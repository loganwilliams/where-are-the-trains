package api

import (
  "fmt"
  "time"
  "net/http"
  "sync"

  "github.com/loganwilliams/where-are-the-trains/server/gtfsjson"
  "github.com/loganwilliams/where-are-the-trains/server/pretty"
)

// Cache for computed API response
type Cache struct {
  response string
  lastUpdated time.Time
  sync.Mutex
}

var (
  c *Cache
)

func NewCache() *Cache { // return a pointer to prevent unintentional struct copies
  return &Cache{
    // it's ok to leave response as the zero val because lastUpdated will cause the initial liveJson to never be seen
    // it's ok to leave lastUpdated as the zero val because that's more than 10s ago
    // it's ok to leave the Mutex as the zero val because that's how you initialize an unlocked lock by convention
  }
}

func LiveTrainsHandler(w http.ResponseWriter, r *http.Request) {
  // Send the correct headers to enable CORS
  w.Header().Set("Content-Type", "text/json; charset=ascii")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
  
  // respond with train positions
  fmt.Fprintf(w, "%s", c.get())
}

func (c *Cache) get() string {
  if c == nil {
    c = NewCache()
  }

  c.Lock()
  defer c.Unlock()

  // query the MTA GTFS API at most once every 10 seconds
  if time.Since(c.lastUpdated) > time.Duration(10 * time.Second) {
    c.response = pretty.Json(string(gtfsjson.GetLiveGeoJSON()))
    c.lastUpdated = time.Now()
  }

  return c.response
}
