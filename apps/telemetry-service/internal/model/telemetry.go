package model

import "time"

type Telemetry struct {
	ID        int       `json:"id"`
	DeviceID  int       `json:"device_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type TelemetryCreate struct {
	DeviceID int    `json:"device_id" binding:"required"`
	Value    float64 `json:"value" binding:"required"`
	Unit     string `json:"unit"`
}

type TelemetryStats struct {
	DeviceID int     `json:"device_id"`
	Count    int     `json:"count"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Avg      float64 `json:"avg"`
	Unit     string  `json:"unit"`
}

type DeviceWebhook struct {
	Event     string `json:"event"`
	DeviceID  int    `json:"device_id"`
	Timestamp string `json:"timestamp"`
}
