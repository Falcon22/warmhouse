package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// TemperatureService handles fetching temperature data from external API
type TemperatureService struct {
	BaseURL      string
	TelemetryURL string
	HTTPClient   *http.Client
}

// TemperatureResponse represents the response from the temperature API
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

// NewTemperatureService creates a new temperature service.
// If telemetryURL is non-empty, every successful reading is also forwarded to
// the Telemetry microservice (Strangler Fig migration step).
func NewTemperatureService(baseURL, telemetryURL string) *TemperatureService {
	return &TemperatureService{
		BaseURL:      baseURL,
		TelemetryURL: telemetryURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *TemperatureService) forwardTelemetry(t *TemperatureResponse) {
	if s.TelemetryURL == "" {
		return
	}
	body, err := json.Marshal(map[string]interface{}{
		"device_id": t.SensorID,
		"metric":    "temperature",
		"value":     t.Value,
		"unit":      t.Unit,
		"timestamp": t.Timestamp.Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("forward telemetry: marshal failed: %v", err)
		return
	}
	go func() {
		resp, err := s.HTTPClient.Post(
			s.TelemetryURL+"/telemetry",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			log.Printf("forward telemetry: post failed: %v", err)
			return
		}
		_ = resp.Body.Close()
	}()
}

// GetTemperature fetches temperature data for a specific location
func (s *TemperatureService) GetTemperature(location string) (*TemperatureResponse, error) {
	url := fmt.Sprintf("%s/temperature?location=%s", s.BaseURL, location)

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching temperature data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var temperatureResp TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&temperatureResp); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	s.forwardTelemetry(&temperatureResp)

	return &temperatureResp, nil
}

// GetTemperatureByID fetches temperature data for a specific sensor ID
func (s *TemperatureService) GetTemperatureByID(sensorID string) (*TemperatureResponse, error) {
	url := fmt.Sprintf("%s/temperature/%s", s.BaseURL, sensorID)

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching temperature data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var temperatureResp TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&temperatureResp); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	s.forwardTelemetry(&temperatureResp)

	return &temperatureResp, nil
}
