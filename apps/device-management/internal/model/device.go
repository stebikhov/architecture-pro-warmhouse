package model

import "time"

type DeviceType string

const (
	TypeSensor    DeviceType = "sensor"
	TypeLamp      DeviceType = "lamp"
	TypeLock      DeviceType = "lock"
	TypeThermostat DeviceType = "thermostat"
	TypeCamera    DeviceType = "camera"
	TypeSwitch    DeviceType = "switch"
	TypeOutlet    DeviceType = "outlet"
	TypeValve     DeviceType = "valve"
	TypeOther     DeviceType = "other"
)

type DeviceStatus string

const (
	StatusActive   DeviceStatus = "active"
	StatusInactive DeviceStatus = "inactive"
	StatusOffline  DeviceStatus = "offline"
)

type Device struct {
	ID              int          `json:"id"`
	RoomID          *int         `json:"room_id"`
	Name            string       `json:"name"`
	Type            DeviceType   `json:"type"`
	Manufacturer    string       `json:"manufacturer"`
	Model           string       `json:"model"`
	SerialNumber    string       `json:"serial_number"`
	FirmwareVersion string       `json:"firmware_version"`
	Status          DeviceStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type DeviceCreate struct {
	RoomID          *int         `json:"room_id"`
	Name            string       `json:"name" binding:"required"`
	Type            DeviceType   `json:"type" binding:"required"`
	Manufacturer    string       `json:"manufacturer"`
	Model           string       `json:"model"`
	SerialNumber    string       `json:"serial_number"`
	FirmwareVersion string       `json:"firmware_version"`
	Status          DeviceStatus `json:"status"`
}

type DeviceUpdate struct {
	RoomID          *int         `json:"room_id"`
	Name            string       `json:"name"`
	Type            DeviceType   `json:"type"`
	Manufacturer    string       `json:"manufacturer"`
	Model           string       `json:"model"`
	SerialNumber    string       `json:"serial_number"`
	FirmwareVersion string       `json:"firmware_version"`
	Status          DeviceStatus `json:"status"`
}
