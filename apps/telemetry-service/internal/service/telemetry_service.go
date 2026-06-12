package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"telemetry-service/internal/model"
	"telemetry-service/internal/repository"
)

type TelemetryService struct {
	repo           *repository.TelemetryRepository
	deviceMgmtURL  string
	temperatureURL string
	httpClient     *http.Client
	thresholdMin   float64
	thresholdMax   float64
}

func NewTelemetryService(repo *repository.TelemetryRepository, deviceMgmtURL, tempAPIURL string, thresholdMin, thresholdMax float64) *TelemetryService {
	return &TelemetryService{
		repo:           repo,
		deviceMgmtURL:  deviceMgmtURL,
		temperatureURL: tempAPIURL,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		thresholdMin:   thresholdMin,
		thresholdMax:   thresholdMax,
	}
}

func (s *TelemetryService) GetTelemetry(ctx context.Context, deviceID int, limit int) ([]model.Telemetry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetByDeviceID(ctx, deviceID, limit)
}

func (s *TelemetryService) GetHistory(ctx context.Context, deviceID int, from, to time.Time) ([]model.Telemetry, error) {
	return s.repo.GetHistory(ctx, deviceID, from, to)
}

func (s *TelemetryService) ReceiveTelemetry(ctx context.Context, t model.TelemetryCreate) (model.Telemetry, error) {
	telemetry, err := s.repo.Create(ctx, t)
	if err != nil {
		return telemetry, err
	}

	s.checkThreshold(ctx, telemetry)
	return telemetry, nil
}

func (s *TelemetryService) GetStats(ctx context.Context, deviceID int) (model.TelemetryStats, error) {
	return s.repo.GetStats(ctx, deviceID)
}

func (s *TelemetryService) ProcessDeviceWebhook(ctx context.Context, webhook model.DeviceWebhook) error {
	log.Printf("[INFO] Received device webhook: event=%s, device_id=%d, timestamp=%s", webhook.Event, webhook.DeviceID, webhook.Timestamp)
	return nil
}

func (s *TelemetryService) GetTemperature(ctx context.Context, location string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/temperature?location=%s", s.temperatureURL, location)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching temperature: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	return result, nil
}

type thresholdPayload struct {
	Event     string             `json:"event"`
	Telemetry model.Telemetry    `json:"telemetry"`
	Timestamp string             `json:"timestamp"`
}

func (s *TelemetryService) checkThreshold(ctx context.Context, t model.Telemetry) {
	if t.Value > s.thresholdMax || t.Value < s.thresholdMin {
		payload := thresholdPayload{
			Event:     "threshold.exceeded",
			Telemetry: t,
			Timestamp: t.Timestamp.Format(time.RFC3339),
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[ERROR] Failed to marshal threshold payload: %v", err)
			return
		}

		url := fmt.Sprintf("%s/webhooks/telemetry", s.deviceMgmtURL)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[ERROR] Failed to create threshold webhook request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("[ERROR] Threshold webhook request failed: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("[WARN] Threshold webhook returned status %d", resp.StatusCode)
		}
	}
}
