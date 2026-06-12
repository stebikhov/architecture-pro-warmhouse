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
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Type        DeviceType   `json:"type"`
	Location    string       `json:"location"`
	Value       float64      `json:"value"`
	Unit        string       `json:"unit"`
	Status      DeviceStatus `json:"status"`
	LastUpdated time.Time    `json:"last_updated"`
	CreatedAt   time.Time    `json:"created_at"`
}

type DeviceCreate struct {
	Name     string       `json:"name" binding:"required"`
	Type     DeviceType   `json:"type" binding:"required"`
	Location string       `json:"location" binding:"required"`
	Status   DeviceStatus `json:"status"`
}

type DeviceUpdate struct {
	Name     string       `json:"name"`
	Type     DeviceType   `json:"type"`
	Location string       `json:"location"`
	Status   DeviceStatus `json:"status"`
}
