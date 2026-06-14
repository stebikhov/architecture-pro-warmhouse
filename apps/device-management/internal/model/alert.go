package model

import "time"

type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

type Alert struct {
	ID         int           `json:"id"`
	HouseID    int           `json:"house_id"`
	DeviceID   *int          `json:"device_id"`
	UserID     int           `json:"user_id"`
	Type       string        `json:"type"`
	Severity   AlertSeverity `json:"severity"`
	Title      string        `json:"title"`
	Message    string        `json:"message"`
	Status     AlertStatus   `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	ResolvedAt *time.Time    `json:"resolved_at"`
}

type AlertCreate struct {
	HouseID  int           `json:"house_id" binding:"required"`
	DeviceID *int          `json:"device_id"`
	UserID   int           `json:"user_id" binding:"required"`
	Type     string        `json:"type" binding:"required"`
	Severity AlertSeverity `json:"severity" binding:"required"`
	Title    string        `json:"title" binding:"required"`
	Message  string        `json:"message"`
	Status   AlertStatus   `json:"status"`
}

type AlertUpdate struct {
	Severity AlertSeverity `json:"severity"`
	Title    string        `json:"title"`
	Message  string        `json:"message"`
	Status   AlertStatus   `json:"status"`
}
