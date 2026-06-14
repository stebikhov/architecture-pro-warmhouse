package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"smarthome/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) GetSensors(ctx context.Context) ([]models.Sensor, error) {
	query := `
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
		FROM devices
		ORDER BY id
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors: %w", err)
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var s models.Sensor
		err := rows.Scan(
			&s.ID,
			&s.RoomID,
			&s.Name,
			&s.Type,
			&s.Manufacturer,
			&s.Model,
			&s.SerialNumber,
			&s.FirmwareVersion,
			&s.Status,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sensor rows: %w", err)
	}

	return sensors, nil
}

func (db *DB) GetSensorByID(ctx context.Context, id int) (models.Sensor, error) {
	query := `
		SELECT id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	var s models.Sensor
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.RoomID,
		&s.Name,
		&s.Type,
		&s.Manufacturer,
		&s.Model,
		&s.SerialNumber,
		&s.FirmwareVersion,
		&s.Status,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error getting sensor by ID: %w", err)
	}

	return s, nil
}

func (db *DB) CreateSensor(ctx context.Context, s models.SensorCreate) (models.Sensor, error) {
	query := `
		INSERT INTO devices (room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'inactive', $8, $8)
		RETURNING id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at
	`

	now := time.Now()
	var sensor models.Sensor
	err := db.Pool.QueryRow(ctx, query,
		s.RoomID,
		s.Name,
		s.Type,
		s.Manufacturer,
		s.Model,
		s.SerialNumber,
		s.FirmwareVersion,
		now,
	).Scan(
		&sensor.ID,
		&sensor.RoomID,
		&sensor.Name,
		&sensor.Type,
		&sensor.Manufacturer,
		&sensor.Model,
		&sensor.SerialNumber,
		&sensor.FirmwareVersion,
		&sensor.Status,
		&sensor.CreatedAt,
		&sensor.UpdatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error creating sensor: %w", err)
	}

	return sensor, nil
}

func (db *DB) UpdateSensor(ctx context.Context, id int, s models.SensorUpdate) (models.Sensor, error) {
	_, err := db.GetSensorByID(ctx, id)
	if err != nil {
		return models.Sensor{}, err
	}

	query := "UPDATE devices SET updated_at = $1"
	args := []interface{}{time.Now()}
	argCount := 2

	if s.Name != "" {
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, s.Name)
		argCount++
	}

	if s.Type != "" {
		query += fmt.Sprintf(", type = $%d", argCount)
		args = append(args, s.Type)
		argCount++
	}

	if s.Manufacturer != "" {
		query += fmt.Sprintf(", manufacturer = $%d", argCount)
		args = append(args, s.Manufacturer)
		argCount++
	}

	if s.Model != "" {
		query += fmt.Sprintf(", model = $%d", argCount)
		args = append(args, s.Model)
		argCount++
	}

	if s.SerialNumber != "" {
		query += fmt.Sprintf(", serial_number = $%d", argCount)
		args = append(args, s.SerialNumber)
		argCount++
	}

	if s.FirmwareVersion != "" {
		query += fmt.Sprintf(", firmware_version = $%d", argCount)
		args = append(args, s.FirmwareVersion)
		argCount++
	}

	if s.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, s.Status)
		argCount++
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", argCount) + `
		RETURNING id, room_id, name, type, manufacturer, model, serial_number, firmware_version, status, created_at, updated_at`
	args = append(args, id)

	var sensor models.Sensor
	err = db.Pool.QueryRow(ctx, query, args...).Scan(
		&sensor.ID,
		&sensor.RoomID,
		&sensor.Name,
		&sensor.Type,
		&sensor.Manufacturer,
		&sensor.Model,
		&sensor.SerialNumber,
		&sensor.FirmwareVersion,
		&sensor.Status,
		&sensor.CreatedAt,
		&sensor.UpdatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error updating sensor: %w", err)
	}

	return sensor, nil
}

func (db *DB) DeleteSensor(ctx context.Context, id int) error {
	tx, err := db.Pool.Begin(ctx)
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
		return fmt.Errorf("error deleting sensor: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("sensor not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (db *DB) UpdateSensorValue(ctx context.Context, id int, value float64, status string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateDevice := `UPDATE devices SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := tx.Exec(ctx, updateDevice, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error updating sensor status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("sensor not found")
	}

	insertTelemetry := `
		INSERT INTO telemetry (device_id, value, unit, timestamp)
		VALUES ($1, $2, 'celsius', $3)
	`
	_, err = tx.Exec(ctx, insertTelemetry, id, value, time.Now())
	if err != nil {
		return fmt.Errorf("error inserting telemetry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
