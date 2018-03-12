package gtfsjson

import (
  "log"
  "time"
  "net/http"
  "bytes"
  "os"
  "encoding/csv"
  "bufio"
  "io"
  "strconv"
  "strings"
  "errors"
  "fmt"
  "sync"

  "github.com/loganwilliams/where-are-the-trains/server/transit_realtime"
  "github.com/golang/protobuf/proto"
  "github.com/paulmach/go.geojson"

)

type Location struct {
  Latitude float64
  Longitude float64
}

// The current position of a train.
type Train struct {
  TrainId string
  Line string
  Status string
  StopId string
  Timestamp time.Time
  Direction string
}

type StopLocationCache struct {
  stopLocations map[string]Location
  sync.Mutex
}

var (
  c *StopLocationCache
)

func GetLiveGeoJSON() []byte {
  geometry := makeGeoJSON(GetLiveTrains())
  rawJSON, _ := geometry.MarshalJSON()

  return rawJSON
}

// GetLiveTrains() returns a GeoJSON []byte object with the most recent position of all trains in the NYC Subway, as
// reported by the MTA's GTFS feed.
func GetLiveTrains() []Train {
  // The MTA has several different endpoints for different lines. My API key is in here, but the abuse potential
  // seems low enough that I'm okay with that.
  datafeeds := [](string){
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

  for _, url := range datafeeds {
    transit, err := getGTFS(url, 3)
    if err != nil {
      log.Println("Error getting GTFS feed: ", err)
    }

    for _, entity := range transit.Entity {
      train, err := trainPositionFromTripUpdate(entity)

      if err == nil {
        // Only include trains that have moved in the last 10 minutes, are reporting times in the present/past
        // and have a line associated with them.
        if train.Timestamp.After(cutoff) && train.Timestamp.Before(now) && train.Line != "" {
          trains = append(trains, *train)
        }
      }
    }
  }

  return trains  
}

// trainPositionFromTripUpdate takes a GTFS protobuf entity and returns a Train object. If there is no
// trip update in the GTFS entity, it returns an empty Train and an error.
func trainPositionFromTripUpdate(entity *transit_realtime.FeedEntity) (*Train, error) {
  if entity.TripUpdate == nil {
    return &Train{}, errors.New("No trip update in entity.")
  }

  tripId := entity.GetTripUpdate().GetTrip().GetTripId()
  direction := directionFromId(tripId)      

  routeId := entity.GetTripUpdate().GetTrip().GetRouteId()
  stopTimes := entity.GetTripUpdate().GetStopTimeUpdate();
  timestamp := time.Unix(int64(stopTimes[0].GetArrival().GetTime()), 0)
  stopId := stopTimes[0].GetStopId()

  return &Train{
    TrainId: tripId, 
    Line: routeId, 
    StopId: stopId, 
    Timestamp: timestamp, 
    Direction: direction}, nil

}

// Using the Trip ID, return a direction.
func directionFromId(id string) (direction string) {
  idParts := strings.Split(id, ".")
  direction = string(idParts[len(idParts)-1][0])
  return
}

// stopLocations reads the MTA stop locations from a file ("stops.txt") and constructs a map.
func getStopLocations() map[string]Location {
  if c == nil {
    c = &StopLocationCache{}
  }

  c.Lock()
  defer c.Unlock()

  if c.stopLocations != nil {
    return c.stopLocations
  }

  c.stopLocations = make(map[string]Location)

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

    c.stopLocations[line[0]] = Location{Latitude: latitude, Longitude: longitude}
  }

  return c.stopLocations
}

// makeGeoJSON takes a list of Train objects and constructs a GeoJSON FeatureCollection.
func makeGeoJSON(trains []Train) *geojson.FeatureCollection {
  fc := geojson.NewFeatureCollection()
  stopList := getStopLocations()

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

// getGTFS downloads a GTFS url from the MTA and unmarshals the protobuf.
func getGTFS(url string, retries int) (*transit_realtime.FeedMessage, error) {
  if retries <= 0 {
    return nil, fmt.Errorf("giving up on url %q", url)
  }

  resp, err := http.Get(url)
  defer resp.Body.Close()

  if err != nil {
   fmt.Printf("failed to fetch for url %q", url)
  }

  buf := new(bytes.Buffer)
  buf.ReadFrom(resp.Body)
  gtfs := buf.Bytes()

  transit := &transit_realtime.FeedMessage{}
  if err := proto.Unmarshal(gtfs, transit); err != nil {
      log.Println("Failed to parse GTFS feed", err)
      return getGTFS(url, retries-1)
  }

  return transit, nil
}
