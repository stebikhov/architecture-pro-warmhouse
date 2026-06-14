package model

import "time"

type UserHouseRole string

const (
	RoleOwner  UserHouseRole = "owner"
	RoleAdmin  UserHouseRole = "admin"
	RoleMember UserHouseRole = "member"
	RoleGuest  UserHouseRole = "guest"
)

type UserHouse struct {
	UserID    int           `json:"user_id"`
	HouseID   int           `json:"house_id"`
	Role      UserHouseRole `json:"role"`
	CreatedAt time.Time     `json:"created_at"`
}

type UserHouseCreate struct {
	UserID  int           `json:"user_id" binding:"required"`
	HouseID int           `json:"house_id" binding:"required"`
	Role    UserHouseRole `json:"role"`
}
