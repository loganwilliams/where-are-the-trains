package main

import (
	"github.com/loganwilliams/where-are-the-trains/server/gtfsjson"
    "github.com/loganwilliams/where-are-the-trains/server/pretty"
	"net/http"
	"log"
	"fmt"
    "os"
    "time"
    "io/ioutil"
)

func currentTrainsHandler(w http.ResponseWriter, r *http.Request) {
    var json string

    info, err := os.Stat("trains.json")
    if err != nil {
        log.Fatal("currentTrainsHandler: ", err)
    }

    if time.Since(info.ModTime()) > time.Duration(10 * time.Second) {
        json = pretty.Json(string(gtfsjson.GetTrains()))

        f, err := os.Create("trains.json")
        if err != nil {
            log.Fatal("Error writing trains.json for write: ", err)
        }
        f.Write([]byte(json))
        f.Close()
    } else {
        f, err := os.Open("trains.json")
        if err != nil {
            log.Fatal("Error openiing trains.json for read: ", err)
        }
        bytes, err := ioutil.ReadAll(f)
        json = string(bytes)
        if err != nil {
            log.Fatal("Error reading trains.json: ", err)
        }
        f.Close()
    }

	w.Header().Set("Content-Type", "text/json; charset=ascii")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
    fmt.Fprintf(w, "%s", json)
}

func main() {
    http.HandleFunc("/live", currentTrainsHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}