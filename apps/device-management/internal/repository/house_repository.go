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

type HouseRepository struct {
	dbPool *pgxpool.Pool
}

func NewHouseRepository(database *db.DB) *HouseRepository {
	return &HouseRepository{dbPool: database.Pool}
}

func (r *HouseRepository) GetAll(ctx context.Context) ([]model.House, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM houses
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying houses: %w", err)
	}
	defer rows.Close()

	return scanHouses(rows)
}

func (r *HouseRepository) GetByOwnerID(ctx context.Context, ownerID int) ([]model.House, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM houses
		WHERE owner_id = $1
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error querying houses by owner: %w", err)
	}
	defer rows.Close()

	return scanHouses(rows)
}

func (r *HouseRepository) GetByID(ctx context.Context, id int) (model.House, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM houses
		WHERE id = $1
	`

	var h model.House
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.OwnerID, &h.Name, &h.Address, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return model.House{}, fmt.Errorf("error getting house by ID: %w", err)
	}

	return h, nil
}

func (r *HouseRepository) Create(ctx context.Context, h model.HouseCreate) (model.House, error) {
	query := `
		INSERT INTO houses (owner_id, name, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_id, name, address, created_at, updated_at
	`

	now := time.Now()
	var house model.House
	err := r.dbPool.QueryRow(ctx, query, h.OwnerID, h.Name, h.Address, now, now).Scan(
		&house.ID, &house.OwnerID, &house.Name, &house.Address, &house.CreatedAt, &house.UpdatedAt,
	)
	if err != nil {
		return model.House{}, fmt.Errorf("error creating house: %w", err)
	}

	return house, nil
}

func (r *HouseRepository) Update(ctx context.Context, id int, h model.HouseUpdate) (model.House, error) {
	query := "UPDATE houses SET updated_at = $1"
	args := []interface{}{time.Now()}
	argCount := 2

	if h.Name != "" {
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, h.Name)
		argCount++
	}

	if h.Address != "" {
		query += fmt.Sprintf(", address = $%d", argCount)
		args = append(args, h.Address)
		argCount++
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", argCount) + `
		RETURNING id, owner_id, name, address, created_at, updated_at`
	args = append(args, id)

	var house model.House
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&house.ID, &house.OwnerID, &house.Name, &house.Address, &house.CreatedAt, &house.UpdatedAt,
	)
	if err != nil {
		return model.House{}, fmt.Errorf("error updating house: %w", err)
	}

	return house, nil
}

func (r *HouseRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM houses WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting house: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("house not found")
	}

	return nil
}

func scanHouses(rows pgx.Rows) ([]model.House, error) {
	var houses []model.House
	for rows.Next() {
		var h model.House
		err := rows.Scan(&h.ID, &h.OwnerID, &h.Name, &h.Address, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning house row: %w", err)
		}
		houses = append(houses, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating house rows: %w", err)
	}

	return houses, nil
}
