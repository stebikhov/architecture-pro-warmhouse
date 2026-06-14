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
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
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

func (r *DeviceRepository) GetByRoomID(ctx context.Context, roomID int) ([]model.Device, error) {
	query := `
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
		FROM devices
		WHERE room_id = $1
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("error querying devices by room: %w", err)
	}
	defer rows.Close()

	return scanDevices(rows)
}

func (r *DeviceRepository) GetByID(ctx context.Context, id int) (model.Device, error) {
	query := `
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	var d model.Device
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.RoomID, &d.Name, &d.Type, &d.Manufacturer, &d.Model, &d.SerialNumber, &d.FirmwareVersion, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error getting device by ID: %w", err)
	}

	return d, nil
}

func (r *DeviceRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (model.Device, error) {
	query := `
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
		FROM devices
		WHERE serial_number = $1
	`

	var d model.Device
	err := r.dbPool.QueryRow(ctx, query, serialNumber).Scan(
		&d.ID, &d.RoomID, &d.Name, &d.Type, &d.Manufacturer, &d.Model, &d.SerialNumber, &d.FirmwareVersion, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error getting device by serial number: %w", err)
	}

	return d, nil
}

func (r *DeviceRepository) Create(ctx context.Context, d model.DeviceCreate) (model.Device, error) {
	query := `
		INSERT INTO devices (room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
	`

	now := time.Now()
	status := d.Status
	if status == "" {
		status = model.StatusInactive
	}

	var device model.Device
	err := r.dbPool.QueryRow(ctx, query, d.RoomID, d.Name, d.Type, d.Manufacturer, d.Model, d.SerialNumber, d.FirmwareVersion, status, now, now).Scan(
		&device.ID, &device.RoomID, &device.Name, &device.Type, &device.Manufacturer, &device.Model, &device.SerialNumber, &device.FirmwareVersion, &device.Status, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error creating device: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) Update(ctx context.Context, id int, d model.DeviceUpdate) (model.Device, error) {
	query := "UPDATE devices SET updated_at = $1"
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

	if d.Manufacturer != "" {
		query += fmt.Sprintf(", manufacturer = $%d", argCount)
		args = append(args, d.Manufacturer)
		argCount++
	}

	if d.Model != "" {
		query += fmt.Sprintf(", model = $%d", argCount)
		args = append(args, d.Model)
		argCount++
	}

	if d.SerialNumber != "" {
		query += fmt.Sprintf(", serial_number = $%d", argCount)
		args = append(args, d.SerialNumber)
		argCount++
	}

	if d.FirmwareVersion != "" {
		query += fmt.Sprintf(", firmware_version = $%d", argCount)
		args = append(args, d.FirmwareVersion)
		argCount++
	}

	if d.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, d.Status)
		argCount++
	}

	if d.RoomID != nil {
		query += fmt.Sprintf(", room_id = $%d", argCount)
		args = append(args, d.RoomID)
		argCount++
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", argCount) + `
		RETURNING id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at`
	args = append(args, id)

	var device model.Device
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&device.ID, &device.RoomID, &device.Name, &device.Type, &device.Manufacturer, &device.Model, &device.SerialNumber, &device.FirmwareVersion, &device.Status, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		return model.Device{}, fmt.Errorf("error updating device: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM telemetry WHERE device_id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting telemetry data: %w", err)
	}

	result, err := tx.Exec(ctx, "DELETE FROM devices WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("device not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func scanDevices(rows pgx.Rows) ([]model.Device, error) {
	var devices []model.Device
	for rows.Next() {
		var d model.Device
		err := rows.Scan(&d.ID, &d.RoomID, &d.Name, &d.Type, &d.Manufacturer, &d.Model, &d.SerialNumber, &d.FirmwareVersion, &d.Status, &d.CreatedAt, &d.UpdatedAt)
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
