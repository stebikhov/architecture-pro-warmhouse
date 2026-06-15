package repository

import (
	"context"
	"fmt"
	"time"

	"telemetry-service/internal/db"
	"telemetry-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryRepository struct {
	dbPool *pgxpool.Pool
}

func NewTelemetryRepository(database *db.DB) *TelemetryRepository {
	return &TelemetryRepository{dbPool: database.Pool}
}

func (r *TelemetryRepository) GetByDeviceID(ctx context.Context, deviceID int, limit int) ([]model.Telemetry, error) {
	query := `
		SELECT id, device_id, value, unit, timestamp
		FROM telemetry
		WHERE device_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.dbPool.Query(ctx, query, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying telemetry: %w", err)
	}
	defer rows.Close()

	return scanTelemetry(rows)
}

func (r *TelemetryRepository) GetHistory(ctx context.Context, deviceID int, from, to time.Time) ([]model.Telemetry, error) {
	query := `
		SELECT id, device_id, value, unit, timestamp
		FROM telemetry
		WHERE device_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
	`

	rows, err := r.dbPool.Query(ctx, query, deviceID, from, to)
	if err != nil {
		return nil, fmt.Errorf("error querying telemetry history: %w", err)
	}
	defer rows.Close()

	return scanTelemetry(rows)
}

func (r *TelemetryRepository) Create(ctx context.Context, t model.TelemetryCreate) (model.Telemetry, error) {
	query := `
		INSERT INTO telemetry (device_id, value, unit, timestamp)
		VALUES ($1, $2, $3, $4)
		RETURNING id, device_id, value, unit, timestamp
	`

	now := time.Now()
	unit := t.Unit
	if unit == "" {
		unit = "celsius"
	}

	var telemetry model.Telemetry
	err := r.dbPool.QueryRow(ctx, query, t.DeviceID, t.Value, unit, now).Scan(
		&telemetry.ID, &telemetry.DeviceID, &telemetry.Value, &telemetry.Unit, &telemetry.Timestamp,
	)
	if err != nil {
		return model.Telemetry{}, fmt.Errorf("error creating telemetry: %w", err)
	}

	return telemetry, nil
}

func (r *TelemetryRepository) GetStats(ctx context.Context, deviceID int) (model.TelemetryStats, error) {
	query := `
		SELECT 
			device_id,
			COUNT(*) as count,
			MIN(value) as min,
			MAX(value) as max,
			AVG(value) as avg
		FROM telemetry
		WHERE device_id = $1
		GROUP BY device_id
	`

	var stats model.TelemetryStats
	err := r.dbPool.QueryRow(ctx, query, deviceID).Scan(
		&stats.DeviceID, &stats.Count, &stats.Min, &stats.Max, &stats.Avg,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.TelemetryStats{}, nil
		}
		return model.TelemetryStats{}, fmt.Errorf("error querying stats: %w", err)
	}

	return stats, nil
}

func scanTelemetry(rows pgx.Rows) ([]model.Telemetry, error) {
	var telemetry []model.Telemetry
	for rows.Next() {
		var t model.Telemetry
		err := rows.Scan(&t.ID, &t.DeviceID, &t.Value, &t.Unit, &t.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("error scanning telemetry row: %w", err)
		}
		telemetry = append(telemetry, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	return telemetry, nil
}
