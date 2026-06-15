package model

import "time"

type House struct {
	ID        int        `json:"id"`
	OwnerID   int        `json:"owner_id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type HouseCreate struct {
	OwnerID int    `json:"owner_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

type HouseUpdate struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
