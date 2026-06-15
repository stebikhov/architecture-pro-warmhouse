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

type NotificationTemplateRepository struct {
	dbPool *pgxpool.Pool
}

func NewNotificationTemplateRepository(database *db.DB) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{dbPool: database.Pool}
}

func (r *NotificationTemplateRepository) GetAll(ctx context.Context) ([]model.NotificationTemplate, error) {
	query := `
		SELECT id, name, event_type, channel, subject, body, is_active, created_at
		FROM notification_templates
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying notification templates: %w", err)
	}
	defer rows.Close()

	return scanNotificationTemplates(rows)
}

func (r *NotificationTemplateRepository) GetByEventType(ctx context.Context, eventType string) ([]model.NotificationTemplate, error) {
	query := `
		SELECT id, name, event_type, channel, subject, body, is_active, created_at
		FROM notification_templates
		WHERE event_type = $1 AND is_active = true
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query, eventType)
	if err != nil {
		return nil, fmt.Errorf("error querying notification templates by event type: %w", err)
	}
	defer rows.Close()

	return scanNotificationTemplates(rows)
}

func (r *NotificationTemplateRepository) GetByChannel(ctx context.Context, channel model.NotificationChannel) ([]model.NotificationTemplate, error) {
	query := `
		SELECT id, name, event_type, channel, subject, body, is_active, created_at
		FROM notification_templates
		WHERE channel = $1 AND is_active = true
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query, channel)
	if err != nil {
		return nil, fmt.Errorf("error querying notification templates by channel: %w", err)
	}
	defer rows.Close()

	return scanNotificationTemplates(rows)
}

func (r *NotificationTemplateRepository) GetByID(ctx context.Context, id int) (model.NotificationTemplate, error) {
	query := `
		SELECT id, name, event_type, channel, subject, body, is_active, created_at
		FROM notification_templates
		WHERE id = $1
	`

	var nt model.NotificationTemplate
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&nt.ID, &nt.Name, &nt.EventType, &nt.Channel, &nt.Subject, &nt.Body, &nt.IsActive, &nt.CreatedAt,
	)
	if err != nil {
		return model.NotificationTemplate{}, fmt.Errorf("error getting notification template by ID: %w", err)
	}

	return nt, nil
}

func (r *NotificationTemplateRepository) Create(ctx context.Context, nt model.NotificationTemplateCreate) (model.NotificationTemplate, error) {
	query := `
		INSERT INTO notification_templates (name, event_type, channel, subject, body, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, event_type, channel, subject, body, is_active, created_at
	`

	now := time.Now()
	isActive := true
	if nt.IsActive != nil {
		isActive = *nt.IsActive
	}

	var template model.NotificationTemplate
	err := r.dbPool.QueryRow(ctx, query, nt.Name, nt.EventType, nt.Channel, nt.Subject, nt.Body, isActive, now).Scan(
		&template.ID, &template.Name, &template.EventType, &template.Channel, &template.Subject, &template.Body, &template.IsActive, &template.CreatedAt,
	)
	if err != nil {
		return model.NotificationTemplate{}, fmt.Errorf("error creating notification template: %w", err)
	}

	return template, nil
}

func (r *NotificationTemplateRepository) Update(ctx context.Context, id int, nt model.NotificationTemplateUpdate) (model.NotificationTemplate, error) {
	query := "UPDATE notification_templates SET"
	args := []interface{}{}
	argCount := 1

	if nt.Name != "" {
		query += fmt.Sprintf(" name = $%d", argCount)
		args = append(args, nt.Name)
		argCount++
	}

	if nt.EventType != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" event_type = $%d", argCount)
		args = append(args, nt.EventType)
		argCount++
	}

	if nt.Channel != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" channel = $%d", argCount)
		args = append(args, nt.Channel)
		argCount++
	}

	if nt.Subject != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" subject = $%d", argCount)
		args = append(args, nt.Subject)
		argCount++
	}

	if nt.Body != "" {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" body = $%d", argCount)
		args = append(args, nt.Body)
		argCount++
	}

	if nt.IsActive != nil {
		if len(args) > 0 {
			query += ","
		}
		query += fmt.Sprintf(" is_active = $%d", argCount)
		args = append(args, *nt.IsActive)
		argCount++
	}

	if len(args) == 0 {
		return model.NotificationTemplate{}, fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount) + `
		RETURNING id, name, event_type, channel, subject, body, is_active, created_at`
	args = append(args, id)

	var template model.NotificationTemplate
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&template.ID, &template.Name, &template.EventType, &template.Channel, &template.Subject, &template.Body, &template.IsActive, &template.CreatedAt,
	)
	if err != nil {
		return model.NotificationTemplate{}, fmt.Errorf("error updating notification template: %w", err)
	}

	return template, nil
}

func (r *NotificationTemplateRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM notification_templates WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting notification template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("notification template not found")
	}

	return nil
}

func scanNotificationTemplates(rows pgx.Rows) ([]model.NotificationTemplate, error) {
	var templates []model.NotificationTemplate
	for rows.Next() {
		var nt model.NotificationTemplate
		err := rows.Scan(&nt.ID, &nt.Name, &nt.EventType, &nt.Channel, &nt.Subject, &nt.Body, &nt.IsActive, &nt.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning notification template row: %w", err)
		}
		templates = append(templates, nt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification template rows: %w", err)
	}

	return templates, nil
}
