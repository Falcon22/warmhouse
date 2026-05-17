package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	locationLivingRoom = "Living Room"
	locationBedroom    = "Bedroom"
	locationKitchen    = "Kitchen"
	locationUnknown    = "Unknown"

	sensorIDLivingRoom = "1"
	sensorIDBedroom    = "2"
	sensorIDKitchen    = "3"
	sensorIDUnknown    = "0"
)

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func resolve(location, sensorID string) (string, string) {
	if location == "" {
		switch sensorID {
		case sensorIDLivingRoom:
			location = locationLivingRoom
		case sensorIDBedroom:
			location = locationBedroom
		case sensorIDKitchen:
			location = locationKitchen
		default:
			location = locationUnknown
		}
	}

	if sensorID == "" {
		switch location {
		case locationLivingRoom:
			sensorID = sensorIDLivingRoom
		case locationBedroom:
			sensorID = sensorIDBedroom
		case locationKitchen:
			sensorID = sensorIDKitchen
		default:
			sensorID = sensorIDUnknown
		}
	}

	return location, sensorID
}

func generate(location, sensorID string) TemperatureResponse {
	value := 18.0 + rand.Float64()*8.0
	value = float64(int(value*10)) / 10

	return TemperatureResponse{
		Value:       value,
		Unit:        "°C",
		Timestamp:   time.Now().UTC(),
		Location:    location,
		Status:      "active",
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: fmt.Sprintf("Temperature reading from %s", location),
	}
}

func handleTemperature(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	sensorID := r.URL.Query().Get("sensorId")

	if id := strings.TrimPrefix(r.URL.Path, "/temperature/"); id != r.URL.Path && id != "" {
		sensorID = id
	}

	location, sensorID = resolve(location, sensorID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(generate(location, sensorID)); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/temperature", handleTemperature)
	mux.HandleFunc("/temperature/", handleTemperature)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	addr := ":" + port
	log.Printf("Temperature API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
