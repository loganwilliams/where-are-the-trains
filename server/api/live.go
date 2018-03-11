package api

import (
  "fmt"
  "time"
  "net/http"

  "github.com/loganwilliams/where-are-the-trains/server/gtfsjson"
  "github.com/loganwilliams/where-are-the-trains/server/pretty"
)

// variables for cacheing MTA GTFS API response
var (
  liveJson string
  lastUpdated time.Time
)

func LiveTrainsHandler(w http.ResponseWriter, r *http.Request) {
  // query the MTA GTFS API at most once every 10 seconds
  if time.Since(lastUpdated) > time.Duration(10 * time.Second) {
    liveJson = pretty.Json(string(gtfsjson.GetTrains()))
    lastUpdated = time.Now()
  }

  // Send the correct headers to enable CORS
  w.Header().Set("Content-Type", "text/json; charset=ascii")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
  
  // respond with train positions
  fmt.Fprintf(w, "%s", liveJson)
}
