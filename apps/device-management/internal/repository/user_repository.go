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

type UserRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{dbPool: database.Pool}
}

func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users
		ORDER BY id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	return scanUsers(rows)
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u model.User
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error getting user by ID: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var u model.User
	err := r.dbPool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error getting user by email: %w", err)
	}

	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, u model.UserCreate) (model.User, error) {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, password_hash, first_name, last_name, phone, created_at, updated_at
	`

	now := time.Now()
	var user model.User
	err := r.dbPool.QueryRow(ctx, query, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone, now, now).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id int, u model.UserUpdate) (model.User, error) {
	query := "UPDATE users SET updated_at = $1"
	args := []interface{}{time.Now()}
	argCount := 2

	if u.FirstName != "" {
		query += fmt.Sprintf(", first_name = $%d", argCount)
		args = append(args, u.FirstName)
		argCount++
	}

	if u.LastName != "" {
		query += fmt.Sprintf(", last_name = $%d", argCount)
		args = append(args, u.LastName)
		argCount++
	}

	if u.Phone != "" {
		query += fmt.Sprintf(", phone = $%d", argCount)
		args = append(args, u.Phone)
		argCount++
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", argCount) + `
		RETURNING id, email, password_hash, first_name, last_name, phone, created_at, updated_at`
	args = append(args, id)

	var user model.User
	err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error updating user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := r.dbPool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func scanUsers(rows pgx.Rows) ([]model.User, error) {
	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
