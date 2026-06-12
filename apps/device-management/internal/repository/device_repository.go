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

type DeviceRepository struct {
	dbPool *pgxpool.Pool
}

func NewDeviceRepository(database *db.DB) *DeviceRepository {
	return &DeviceRepository{dbPool: database.Pool}
}

func (r *DeviceRepository) GetAll(ctx context.Context) ([]model.Device, error) {
	query := `
		SELECT id, name, type, location, value, unit, status, last_updated, created_at
		FROM devices
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying devices: %w", err)
	}
	defer rows.Close()

	return scanDevices(rows)
}

func (r *DeviceRepository) GetByID(ctx context.Context, id int) (model.Device, error) {
	query := `
		SELECT id, name, type, location, value, unit, status, last_updated, created_at
		FROM devices
		WHERE id = $1
	`

	var d model.Device
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.Name, &d.Type, &d.Location, &d.Value, &d.Unit, &d.Status, &d.LastUpdated, &d.CreatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error getting device by ID: %w", err)
	}

	return d, nil
}

func (r *DeviceRepository) Create(ctx context.Context, d model.DeviceCreate) (model.Device, error) {
	query := `
		INSERT INTO devices (name, type, location, value, unit, status, last_updated, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, type, location, value, unit, status, last_updated, created_at
	`

	now := time.Now()
	status := d.Status
	if status == "" {
		status = model.StatusInactive
	}

	var device model.Device
	err := r.dbPool.QueryRow(ctx, query, d.Name, d.Type, d.Location, 0.0, "", status, now, now).Scan(
		&device.ID, &device.Name, &device.Type, &device.Location, &device.Value, &device.Unit, &device.Status, &device.LastUpdated, &device.CreatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error creating device: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) Update(ctx context.Context, id int, d model.DeviceUpdate) (model.Device, error) {
	query := "UPDATE devices SET last_updated = $1"
	args := []interface{}{time.Now()}
	argCount := 2

	if d.Name != "" {
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, d.Name)
		argCount++
	}

	if d.Type != "" {
		query += fmt.Sprintf(", type = $%d", argCount)
		args = append(args, d.Type)
		argCount++
	}

	if d.Location != "" {
		query += fmt.Sprintf(", location = $%d", argCount)
		args = append(args, d.Location)
		argCount++
	}

	if d.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, d.Status)
		argCount++
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", argCount) + `
		RETURNING id, name, type, location, value, unit, status, last_updated, created_at`
	args = append(args, id)

	var device model.Device
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&device.ID, &device.Name, &device.Type, &device.Location, &device.Value, &device.Unit, &device.Status, &device.LastUpdated, &device.CreatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error updating device: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM devices WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("device not found")
	}

	return nil
}

func scanDevices(rows pgx.Rows) ([]model.Device, error) {
	var devices []model.Device
	for rows.Next() {
		var d model.Device
		err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Location, &d.Value, &d.Unit, &d.Status, &d.LastUpdated, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning device row: %w", err)
		}
		devices = append(devices, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return devices, nil
}
