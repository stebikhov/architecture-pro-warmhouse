package model

import "time"

type NotificationChannel string

const (
	ChannelEmail  NotificationChannel = "email"
	ChannelSMS    NotificationChannel = "sms"
	ChannelPush   NotificationChannel = "push"
	ChannelWebhook NotificationChannel = "webhook"
)

type NotificationTemplate struct {
	ID        int                   `json:"id"`
	Name      string                `json:"name"`
	EventType string                `json:"event_type"`
	Channel   NotificationChannel   `json:"channel"`
	Subject   string                `json:"subject"`
	Body      string                `json:"body"`
	IsActive  bool                  `json:"is_active"`
	CreatedAt time.Time             `json:"created_at"`
}

type NotificationTemplateCreate struct {
	Name      string                `json:"name" binding:"required"`
	EventType string                `json:"event_type" binding:"required"`
	Channel   NotificationChannel   `json:"channel" binding:"required"`
	Subject   string                `json:"subject"`
	Body      string                `json:"body" binding:"required"`
	IsActive  *bool                 `json:"is_active"`
}

type NotificationTemplateUpdate struct {
	Name      string                `json:"name"`
	EventType string                `json:"event_type"`
	Channel   NotificationChannel   `json:"channel"`
	Subject   string                `json:"subject"`
	Body      string                `json:"body"`
	IsActive  *bool                 `json:"is_active"`
}
