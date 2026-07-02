package models

import "time"

// User represents a user in the system
type User struct {
	UserID      string    `json:"user_id" db:"user_id"`
	Email       string    `json:"email" db:"email"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Department  string    `json:"department" db:"department"`
	Color       string    `json:"color" db:"color"` // Hex color code (e.g., #FF5733)
	IsAdmin     bool      `json:"is_admin" db:"is_admin"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ApprovedUser is a sign-in allowlist entry with registration status
type ApprovedUser struct {
	Email       string    `json:"email" db:"email"`
	AddedBy     *string   `json:"added_by" db:"added_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Registered  bool      `json:"registered" db:"registered"`
	DisplayName *string   `json:"display_name" db:"display_name"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	Department  string `json:"department"`
}
