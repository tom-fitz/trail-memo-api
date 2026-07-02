package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tom-fitz/trailmemo-api/internal/models"
)

// ApprovedUserRepository handles the sign-in allowlist
type ApprovedUserRepository struct {
	db *sqlx.DB
}

// NewApprovedUserRepository creates a new approved user repository
func NewApprovedUserRepository(db *sqlx.DB) *ApprovedUserRepository {
	return &ApprovedUserRepository{db: db}
}

// IsApproved reports whether an email is on the allowlist (case-insensitive)
func (r *ApprovedUserRepository) IsApproved(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM approved_users WHERE email = $1)`
	if err := r.db.GetContext(ctx, &exists, query, strings.ToLower(email)); err != nil {
		return false, fmt.Errorf("error checking approved email: %v", err)
	}
	return exists, nil
}

// List returns all approved emails with registration status
func (r *ApprovedUserRepository) List(ctx context.Context) ([]models.ApprovedUser, error) {
	approved := []models.ApprovedUser{}
	query := `
		SELECT a.email, a.added_by, a.created_at,
		       (u.user_id IS NOT NULL) AS registered,
		       u.display_name AS display_name
		FROM approved_users a
		LEFT JOIN users u ON LOWER(u.email) = a.email
		ORDER BY a.created_at DESC
	`
	if err := r.db.SelectContext(ctx, &approved, query); err != nil {
		return nil, fmt.Errorf("error listing approved users: %v", err)
	}
	return approved, nil
}

// Add puts an email on the allowlist (idempotent)
func (r *ApprovedUserRepository) Add(ctx context.Context, email, addedBy string) error {
	query := `INSERT INTO approved_users (email, added_by) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, query, strings.ToLower(email), addedBy); err != nil {
		return fmt.Errorf("error adding approved user: %v", err)
	}
	return nil
}

// Remove deletes an email from the allowlist
func (r *ApprovedUserRepository) Remove(ctx context.Context, email string) error {
	query := `DELETE FROM approved_users WHERE email = $1`
	if _, err := r.db.ExecContext(ctx, query, strings.ToLower(email)); err != nil {
		return fmt.Errorf("error removing approved user: %v", err)
	}
	return nil
}
