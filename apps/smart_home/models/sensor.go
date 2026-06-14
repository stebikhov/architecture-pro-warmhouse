package models

import (
	"time"
)

type SensorType string

const (
	Temperature SensorType = "temperature"
)

type Sensor struct {
	ID              int        `json:"id"`
	RoomID          *int       `json:"room_id"`
	Name            string     `json:"name"`
	Type            SensorType `json:"type"`
	Manufacturer    string     `json:"manufacturer"`
	Model           string     `json:"model"`
	SerialNumber    string     `json:"serial_number"`
	FirmwareVersion string     `json:"firmware_version"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type SensorCreate struct {
	RoomID          *int       `json:"room_id"`
	Name            string     `json:"name" binding:"required"`
	Type            SensorType `json:"type" binding:"required"`
	Manufacturer    string     `json:"manufacturer"`
	Model           string     `json:"model"`
	SerialNumber    string     `json:"serial_number"`
	FirmwareVersion string     `json:"firmware_version"`
}

type SensorUpdate struct {
	Name            string     `json:"name"`
	Type            SensorType `json:"type"`
	Manufacturer    string     `json:"manufacturer"`
	Model           string     `json:"model"`
	SerialNumber    string     `json:"serial_number"`
	FirmwareVersion string     `json:"firmware_version"`
	Status          string     `json:"status"`
}
