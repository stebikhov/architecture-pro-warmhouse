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

type AlertRepository struct {
	dbPool *pgxpool.Pool
}

func NewAlertRepository(database *db.DB) *AlertRepository {
	return &AlertRepository{dbPool: database.Pool}
}

func (r *AlertRepository) GetAll(ctx context.Context) ([]model.Alert, error) {
	query := `
		SELECT id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
		FROM alerts
		ORDER BY created_at DESC
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

func (r *AlertRepository) GetByHouseID(ctx context.Context, houseID int) ([]model.Alert, error) {
	query := `
		SELECT id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
		FROM alerts
		WHERE house_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.dbPool.Query(ctx, query, houseID)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts by house: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

func (r *AlertRepository) GetByUserID(ctx context.Context, userID int) ([]model.Alert, error) {
	query := `
		SELECT id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
		FROM alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.dbPool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts by user: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

func (r *AlertRepository) GetByDeviceID(ctx context.Context, deviceID int) ([]model.Alert, error) {
	query := `
		SELECT id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
		FROM alerts
		WHERE device_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.dbPool.Query(ctx, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("error querying alerts by device: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

func (r *AlertRepository) GetByID(ctx context.Context, id int) (model.Alert, error) {
	query := `
		SELECT id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
		FROM alerts
		WHERE id = $1
	`

	var a model.Alert
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.HouseID, &a.DeviceID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Message, &a.Status, &a.CreatedAt, &a.ResolvedAt,
	)
	if err != nil {
		return model.Alert{}, fmt.Errorf("error getting alert by ID: %w", err)
	}

	return a, nil
}

func (r *AlertRepository) Create(ctx context.Context, a model.AlertCreate) (model.Alert, error) {
	query := `
		INSERT INTO alerts (house_id, device_id, user_id, type, severity, title, message, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at
	`

	now := time.Now()
	status := a.Status
	if status == "" {
		status = model.AlertStatusActive
	}

	var alert model.Alert
	err := r.dbPool.QueryRow(ctx, query, a.HouseID, a.DeviceID, a.UserID, a.Type, a.Severity, a.Title, a.Message, status, now).Scan(
		&alert.ID, &alert.HouseID, &alert.DeviceID, &alert.UserID, &alert.Type, &alert.Severity, &alert.Title, &alert.Message, &alert.Status, &alert.CreatedAt, &alert.ResolvedAt,
	)
	if err != nil {
		return model.Alert{}, fmt.Errorf("error creating alert: %w", err)
	}

	return alert, nil
}

func (r *AlertRepository) Update(ctx context.Context, id int, a model.AlertUpdate) (model.Alert, error) {
	query := "UPDATE alerts SET"
	args := []interface{}{}
	argCount := 1

	if a.Severity != "" {
		query += fmt.Sprintf(" severity = $%d", argCount)
		args = append(args, a.Severity)
		argCount++
	}

	if a.Title != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" title = $%d", argCount)
		args = append(args, a.Title)
		argCount++
	}

	if a.Message != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" message = $%d", argCount)
		args = append(args, a.Message)
		argCount++
	}

	if a.Status != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" status = $%d", argCount)
		args = append(args, a.Status)
		argCount++
	}

	if len(args) == 0 {
		return model.Alert{}, fmt.Errorf("no fields to update")
	}

	if a.Status == model.AlertStatusResolved {
		resolvedAt := time.Now()
		query += fmt.Sprintf(", resolved_at = $%d", argCount)
		args = append(args, resolvedAt)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount) + `
		RETURNING id, house_id, device_id, user_id, type, severity, title, message, status, created_at, resolved_at`
	args = append(args, id)

	var alert model.Alert
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&alert.ID, &alert.HouseID, &alert.DeviceID, &alert.UserID, &alert.Type, &alert.Severity, &alert.Title, &alert.Message, &alert.Status, &alert.CreatedAt, &alert.ResolvedAt,
	)
	if err != nil {
		return model.Alert{}, fmt.Errorf("error updating alert: %w", err)
	}

	return alert, nil
}

func (r *AlertRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM alerts WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("alert not found")
	}

	return nil
}

func scanAlerts(rows pgx.Rows) ([]model.Alert, error) {
	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(&a.ID, &a.HouseID, &a.DeviceID, &a.UserID, &a.Type, &a.Severity, &a.Title, &a.Message, &a.Status, &a.CreatedAt, &a.ResolvedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning alert row: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	return alerts, nil
}
