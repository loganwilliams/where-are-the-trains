package main

import (
	"net/http"
	"log"

    "github.com/loganwilliams/where-are-the-trains/server/api"
)

func main() {
    http.HandleFunc("/nyc/api/live", api.LiveTrainsHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}