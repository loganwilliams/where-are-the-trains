package gtfsjson

import (
  "testing"
  "os"
  "io/ioutil"
  "time"

  "github.com/loganwilliams/where-are-the-trains/server/transit_realtime"
  "github.com/golang/protobuf/proto"
)

// This test should be mocked out rather than requiring a network connection to the MTA, but it's okay for now.
func TestGetLiveTrains(t *testing.T) {
  trains := GetLiveTrains()

  if len(trains) < 50 {
    t.Errorf("A suspiciously low number of trains were returned.")
  }
}

// Test that we read the expected data from a local GTFS protobuf.
func TestGetTrainPositionFromTripUpdate(t *testing.T) {
  gtfsFile, err := os.Open("./ACE_test.gtfs")
  if err != nil {
    t.Errorf("Couldn't open test GTFS file: %v", err)
  }

  gtfsTest, err := ioutil.ReadAll(gtfsFile)
  if err != nil {
    t.Errorf("Couldn't read from GTFS file: %v", err)
  }

  transit := &transit_realtime.FeedMessage{}
  if err := proto.Unmarshal(gtfsTest, transit); err != nil {
    t.Errorf("Couldn't unmarshal protobuf: %v", err)
  }

  trains := trainList(transit, time.Unix(1520867016, 0))

  if len(trains) != 70 {
    t.Errorf("Train list had %v trains, epected 70", len(trains))
  }

  if (trains[0].StopId != "A07N") {
    t.Errorf("Expected first train to have StopId of A07N, saw %v", trains[0].StopId)
  }

  trains = trainList(transit, time.Unix(1520867016, 0).Add(60*time.Minute))

  if len(trains) != 0 {
    t.Errorf("Expected train update to show 0 trains after 60 minutes, saw %v trains", len(trains))
  }

  trains = trainList(transit, time.Unix(1520867016, 0).Add(-30*time.Minute))

  if len(trains) != 0 {
    t.Errorf("Expected train update to show 0 trains after -30 minutes, saw %v trains", len(trains))
  }
}

// Test that we can read the stop locations correctly.
func TestGetStopLocations(t *testing.T) {
  stopLocations := getStopLocations()

  if stopLocations["124S"].Latitude != 40.77344 {
    t.Errorf("Stop location incorrect. 124S.Latitude = %v, wanted 40.77344", stopLocations["124S"].Latitude)
  }

  if len(stopLocations) != 1503 {
    t.Errorf("Incorrect number of stops: %v, wanted 1503", len(stopLocations))
  }

  // do it again to test from cache
  stopLocations = getStopLocations()

  if stopLocations["124S"].Latitude != 40.77344 {
    t.Errorf("Stop location incorrect from cache. 124S.Latitude = %v, wanted 40.77344", stopLocations["124S"].Latitude)
  }

  if len(stopLocations) != 1503 {
    t.Errorf("Incorrect number of stops from cache: %v, wanted 1503", len(stopLocations))
  }
}
