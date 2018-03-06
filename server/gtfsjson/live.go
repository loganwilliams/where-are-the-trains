package gtfsjson

import (
	"github.com/loganwilliams/where-are-the-trains/server/transit_realtime"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
	"net/http"
	"bytes"
	"os"
	"encoding/csv"
	"bufio"
	"io"
	"strconv"
	"github.com/paulmach/go.geojson"
	"strings"
)

type Location struct {
	Latitude float64
	Longitude float64
}

type Train struct {
	TrainId string
	Line string
	Status string
	StopId string
	Timestamp time.Time
	Direction string
}

func trainPosition(entity *transit_realtime.FeedEntity) Train {
	if entity.Vehicle != nil && entity.Vehicle.GetStopId() != "" {

		tripId := entity.Vehicle.Trip.GetTripId()
		idParts := strings.Split(tripId, ".")
		direction := string(idParts[len(idParts)-1][0])

		routeId := entity.Vehicle.Trip.GetRouteId()
		timestamp := time.Unix(int64(entity.Vehicle.GetTimestamp()), 0)
		stopId := entity.Vehicle.GetStopId()
		stopStatus := entity.Vehicle.GetCurrentStatus().String()

		train := Train{TrainId: tripId, Line: routeId, StopId: stopId, Status: stopStatus, Timestamp: timestamp, Direction: direction}
		return train
	} else {
		// this is necessary for some lines
		if entity.TripUpdate != nil {

			// fmt.Println(entity.GetTripUpdate())

			tripId := entity.GetTripUpdate().GetTrip().GetTripId()
			idParts := strings.Split(tripId, ".")
			direction := string(idParts[len(idParts)-1][0])

			routeId := entity.GetTripUpdate().GetTrip().GetRouteId()
			stopTimes := entity.GetTripUpdate().GetStopTimeUpdate();
			timestamp := time.Unix(int64(stopTimes[0].GetArrival().GetTime()), 0)
			stopId := stopTimes[0].GetStopId()

			train := Train{TrainId: tripId, Line: routeId, StopId: stopId, Timestamp: timestamp, Direction: direction}
			return train
		} else {
			return Train{}
		}
	}	
}

func stopLocations() map[string]Location {
	stopList := make(map[string]Location)

	stops, error := os.Open("gtfsjson/stops.txt")

	if error != nil {
		log.Fatal("Error opening stops.txt: ", error)
	}

    reader := csv.NewReader(bufio.NewReader(stops))
	line, error := reader.Read()

    for {
        line, error = reader.Read()

        if error == io.EOF {
            break
        } else if error != nil {
            log.Fatal("stopLocations: ", error)
        }

        latitude, _ := strconv.ParseFloat(line[4], 64)
        longitude, _ := strconv.ParseFloat(line[5], 64)

        stopList[line[0]] = Location{Latitude: latitude, Longitude: longitude}
    }

    return stopList
}

func makeGeoJSON(trains []Train) *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()
	stopList := stopLocations()

	for _, train := range(trains) {
		stop := train.StopId
		stopLocation := stopList[stop]

		f := geojson.NewPointFeature([]float64{stopLocation.Longitude, stopLocation.Latitude})
		f.SetProperty("id", train.TrainId)
		f.SetProperty("stopId", train.StopId)
		f.SetProperty("line", train.Line)
		f.SetProperty("time", train.Timestamp)
		f.SetProperty("direction", train.Direction)
		fc.AddFeature(f)
	}

	return fc
}

func getGTFS(url string) *transit_realtime.FeedMessage {

	resp, _ := http.Get(url)
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	gtfs := buf.Bytes()


	transit := &transit_realtime.FeedMessage{}
	if err := proto.Unmarshal(gtfs, transit); err != nil {
	    log.Println("Failed to parse GTFS feed", err)
	    return getGTFS(url)
	}

	return transit
}

func GetTrains() []byte {
	datafeeds := [9](string){
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=1",  // 123456S
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=26", // ACE
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=16", // NQRW
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=21", // BDFM
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=2",  // L
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=11", // SIR
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=31", // G
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=36", // JZ
		"http://datamine.mta.info/mta_esi.php?key=5a28db44c9856c30f98eeac4cd09a345&feed_id=51",  // 7
	}

	var trains []Train
	now := time.Now()
	cutoff := now.Add(-10.0*time.Minute)
	goodCount := 0

	for i := 0; i < 9; i++ {
		transit := getGTFS(datafeeds[i])		

		for _, entity := range transit.Entity {
			train := trainPosition(entity)
			if train.Timestamp != time.Unix(0, 0) && train.Timestamp.After(cutoff) && train.Timestamp.Before(now) && train.Line != "" {
				trains = append(trains, trainPosition(entity))
				goodCount += 1
			}
		}
	}

	geometry := makeGeoJSON(trains)
	rawJSON, _ := geometry.MarshalJSON()

	return rawJSON
}
