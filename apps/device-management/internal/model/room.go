package model

import "time"

type Room struct {
	ID        int        `json:"id"`
	HouseID   int        `json:"house_id"`
	Name      string     `json:"name"`
	Floor     *int       `json:"floor"`
	AreaSqm   *float64   `json:"area_sqm"`
	CreatedAt time.Time  `json:"created_at"`
}

type RoomCreate struct {
	HouseID int    `json:"house_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Floor   *int   `json:"floor"`
	AreaSqm *float64 `json:"area_sqm"`
}

type RoomUpdate struct {
	Name    string   `json:"name"`
	Floor   *int     `json:"floor"`
	AreaSqm *float64 `json:"area_sqm"`
}
