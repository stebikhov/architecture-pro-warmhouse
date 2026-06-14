package service

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"device-management/internal/model"
	"device-management/internal/repository"
)

type DeviceService struct {
	repo       *repository.DeviceRepository
	webhookURL string
	httpClient *http.Client
}

func NewDeviceService(repo *repository.DeviceRepository, webhookURL string) *DeviceService {
	return &DeviceService{
		repo:       repo,
		webhookURL: webhookURL,
		httpClient: &http.Client{},
	}
}

func (s *DeviceService) GetAllDevices(ctx context.Context) ([]model.Device, error) {
	return s.repo.GetAll(ctx)
}

func (s *DeviceService) GetDeviceByID(ctx context.Context, id int) (model.Device, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DeviceService) CreateDevice(ctx context.Context, d model.DeviceCreate) (model.Device, error) {
	device, err := s.repo.Create(ctx, d)
	if err != nil {
		return device, err
	}

	s.sendWebhook(ctx, "device.created", device)
	return device, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, id int, d model.DeviceUpdate) (model.Device, error) {
	device, err := s.repo.Update(ctx, id, d)
	if err != nil {
		return device, err
	}

	s.sendWebhook(ctx, "device.updated", device)
	return device, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, id int) error {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.sendWebhook(ctx, "device.deleted", device)
	return nil
}

func (s *DeviceService) GetDeviceStatus(ctx context.Context, id int) (model.Device, error) {
	return s.repo.GetByID(ctx, id)
}

type webhookPayload struct {
	Event     string        `json:"event"`
	DeviceID  int           `json:"device_id"`
	Device    model.Device  `json:"device"`
	Timestamp string        `json:"timestamp"`
}

func (s *DeviceService) sendWebhook(ctx context.Context, event string, device model.Device) {
	if s.webhookURL == "" {
		log.Printf("[WARN] Webhook URL is empty, skipping webhook for event: %s", event)
		return
	}

	payload := webhookPayload{
		Event:     event,
		DeviceID:  device.ID,
		Device:    device,
		Timestamp: device.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal webhook payload for event %s: %v", event, err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] Failed to create webhook request for event %s: %v", event, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Webhook request failed for event %s: %v", event, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[WARN] Webhook returned non-success status %d for event %s", resp.StatusCode, event)
	} else {
		log.Printf("[INFO] Webhook sent successfully for event %s", event)
	}
}
