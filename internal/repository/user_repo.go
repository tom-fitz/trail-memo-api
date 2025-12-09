package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tom-fitz/trailmemo-api/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (user_id, email, display_name, department)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.UserID,
		user.Email,
		user.DisplayName,
		user.Department,
	).Scan(&user.CreatedAt)

	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT user_id, email, display_name, department, created_at
		FROM users
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT user_id, email, display_name, department, created_at
		FROM users
		WHERE email = $1
	`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %v", err)
	}

	return &user, nil
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET display_name = $1, department = $2
		WHERE user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, user.DisplayName, user.Department, user.UserID)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}
