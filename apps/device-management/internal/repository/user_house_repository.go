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

type UserHouseRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserHouseRepository(database *db.DB) *UserHouseRepository {
	return &UserHouseRepository{dbPool: database.Pool}
}

func (r *UserHouseRepository) GetAll(ctx context.Context) ([]model.UserHouse, error) {
	query := `
		SELECT user_id, house_id, role, created_at
		FROM user_house
		ORDER BY house_id, user_id
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying user_house: %w", err)
	}
	defer rows.Close()

	return scanUserHouses(rows)
}

func (r *UserHouseRepository) GetByUserID(ctx context.Context, userID int) ([]model.UserHouse, error) {
	query := `
		SELECT user_id, house_id, role, created_at
		FROM user_house
		WHERE user_id = $1
		ORDER BY house_id
	`

	rows, err := r.dbPool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user_house by user: %w", err)
	}
	defer rows.Close()

	return scanUserHouses(rows)
}

func (r *UserHouseRepository) GetByHouseID(ctx context.Context, houseID int) ([]model.UserHouse, error) {
	query := `
		SELECT user_id, house_id, role, created_at
		FROM user_house
		WHERE house_id = $1
		ORDER BY user_id
	`

	rows, err := r.dbPool.Query(ctx, query, houseID)
	if err != nil {
		return nil, fmt.Errorf("error querying user_house by house: %w", err)
	}
	defer rows.Close()

	return scanUserHouses(rows)
}

func (r *UserHouseRepository) GetByUserIDAndHouseID(ctx context.Context, userID, houseID int) (model.UserHouse, error) {
	query := `
		SELECT user_id, house_id, role, created_at
		FROM user_house
		WHERE user_id = $1 AND house_id = $2
	`

	var uh model.UserHouse
	err := r.dbPool.QueryRow(ctx, query, userID, houseID).Scan(
		&uh.UserID, &uh.HouseID, &uh.Role, &uh.CreatedAt,
	)
	if err != nil {
		return model.UserHouse{}, fmt.Errorf("error getting user_house: %w", err)
	}

	return uh, nil
}

func (r *UserHouseRepository) Create(ctx context.Context, uh model.UserHouseCreate) (model.UserHouse, error) {
	query := `
		INSERT INTO user_house (user_id, house_id, role, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, house_id, role, created_at
	`

	now := time.Now()
	role := uh.Role
	if role == "" {
		role = model.RoleMember
	}

	var userHouse model.UserHouse
	err := r.dbPool.QueryRow(ctx, query, uh.UserID, uh.HouseID, role, now).Scan(
		&userHouse.UserID, &userHouse.HouseID, &userHouse.Role, &userHouse.CreatedAt,
	)
	if err != nil {
		return model.UserHouse{}, fmt.Errorf("error creating user_house: %w", err)
	}

	return userHouse, nil
}

func (r *UserHouseRepository) UpdateRole(ctx context.Context, userID, houseID int, role model.UserHouseRole) (model.UserHouse, error) {
	query := `
		UPDATE user_house SET role = $1
		WHERE user_id = $2 AND house_id = $3
		RETURNING user_id, house_id, role, created_at
	`

	var uh model.UserHouse
	err := r.dbPool.QueryRow(ctx, query, role, userID, houseID).Scan(
		&uh.UserID, &uh.HouseID, &uh.Role, &uh.CreatedAt,
	)
	if err != nil {
		return model.UserHouse{}, fmt.Errorf("error updating user_house role: %w", err)
	}

	return uh, nil
}

func (r *UserHouseRepository) Delete(ctx context.Context, userID, houseID int) error {
	query := "DELETE FROM user_house WHERE user_id = $1 AND house_id = $2"
	result, err := r.dbPool.Exec(ctx, query, userID, houseID)
	if err != nil {
		return fmt.Errorf("error deleting user_house: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("user_house not found")
	}

	return nil
}

func scanUserHouses(rows pgx.Rows) ([]model.UserHouse, error) {
	var userHouses []model.UserHouse
	for rows.Next() {
		var uh model.UserHouse
		err := rows.Scan(&uh.UserID, &uh.HouseID, &uh.Role, &uh.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning user_house row: %w", err)
		}
		userHouses = append(userHouses, uh)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user_house rows: %w", err)
	}

	return userHouses, nil
}
