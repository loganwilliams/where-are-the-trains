package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestLiveTrainsHandler(t *testing.T) {
  // Create a request to test our handler with.
  req, err := http.NewRequest("GET", "/live", nil)
  if err != nil {
    t.Fatal(err)
  }

  // Create a ResponseRecorder to record the response.
  rr := httptest.NewRecorder()
  handler := http.HandlerFunc(LiveTrainsHandler)

  // Call handler directly
  handler.ServeHTTP(rr, req)

  // Check the status code is 200 - OK
  if status := rr.Code; status != http.StatusOK {
    t.Errorf("handler returned wrong status code: got %v want %v",
      status, http.StatusOK)
  }
}