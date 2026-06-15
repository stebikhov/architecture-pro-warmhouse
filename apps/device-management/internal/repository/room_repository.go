package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"device-management/internal/db"
	"device-management/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	dbPool *pgxpool.Pool
}

func NewRoomRepository(database *db.DB) *RoomRepository {
	return &RoomRepository{dbPool: database.Pool}
}

func (r *RoomRepository) GetAll(ctx context.Context) ([]model.Room, error) {
	query := `
		SELECT id, house_id, name, floor, area_sqm, created_at
		FROM rooms
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying rooms: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}

func (r *RoomRepository) GetByHouseID(ctx context.Context, houseID int) ([]model.Room, error) {
	query := `
		SELECT id, house_id, name, floor, area_sqm, created_at
		FROM rooms
		WHERE house_id = $1
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query, houseID)
	if err != nil {
		return nil, fmt.Errorf("error querying rooms by house: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}

func (r *RoomRepository) GetByID(ctx context.Context, id int) (model.Room, error) {
	query := `
		SELECT id, house_id, name, floor, area_sqm, created_at
		FROM rooms
		WHERE id = $1
	`

	var rm model.Room
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&rm.ID, &rm.HouseID, &rm.Name, &rm.Floor, &rm.AreaSqm, &rm.CreatedAt,
	)
	if err != nil {
		return model.Room{}, fmt.Errorf("error getting room by ID: %w", err)
	}

	return rm, nil
}

func (r *RoomRepository) Create(ctx context.Context, rm model.RoomCreate) (model.Room, error) {
	query := `
		INSERT INTO rooms (house_id, name, floor, area_sqm, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, house_id, name, floor, area_sqm, created_at
	`

	now := time.Now()
	var room model.Room
	err := r.dbPool.QueryRow(ctx, query, rm.HouseID, rm.Name, rm.Floor, rm.AreaSqm, now).Scan(
		&room.ID, &room.HouseID, &room.Name, &room.Floor, &room.AreaSqm, &room.CreatedAt,
	)
	if err != nil {
		return model.Room{}, fmt.Errorf("error creating room: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) Update(ctx context.Context, id int, rm model.RoomUpdate) (model.Room, error) {
	query := "UPDATE rooms SET"
	args := []interface{}{}
	argCount := 1

	if rm.Name != "" {
		query += fmt.Sprintf(" name = $%d", argCount)
		args = append(args, rm.Name)
		argCount++
	}

	if rm.Floor != nil {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" floor = $%d", argCount)
		args = append(args, rm.Floor)
		argCount++
	}

	if rm.AreaSqm != nil {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" area_sqm = $%d", argCount)
		args = append(args, rm.AreaSqm)
		argCount++
	}

	if len(args) == 0 {
		return model.Room{}, fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount) + `
		RETURNING id, house_id, name, floor, area_sqm, created_at`
	args = append(args, id)

	var room model.Room
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&room.ID, &room.HouseID, &room.Name, &room.Floor, &room.AreaSqm, &room.CreatedAt,
	)
	if err != nil {
		return model.Room{}, fmt.Errorf("error updating room: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM rooms WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting room: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("room not found")
	}

	return nil
}

func scanRooms(rows pgx.Rows) ([]model.Room, error) {
	var rooms []model.Room
	for rows.Next() {
		var rm model.Room
		err := rows.Scan(&rm.ID, &rm.HouseID, &rm.Name, &rm.Floor, &rm.AreaSqm, &rm.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning room row: %w", err)
		}
		rooms = append(rooms, rm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating room rows: %w", err)
	}

	return rooms, nil
}
