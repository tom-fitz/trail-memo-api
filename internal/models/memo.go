package models

import (
	"time"

	"github.com/google/uuid"
)

// Location represents GPS coordinates
type Location struct {
	Latitude  float64  `json:"latitude" db:"latitude"`
	Longitude float64  `json:"longitude" db:"longitude"`
	Accuracy  *float64 `json:"accuracy,omitempty" db:"location_accuracy"`
	Address   *string  `json:"address,omitempty" db:"address"`
}

// Memo represents a voice memo
type Memo struct {
	MemoID          uuid.UUID  `json:"memo_id" db:"memo_id"`
	UserID          string     `json:"user_id" db:"user_id"`
	UserName        string     `json:"user_name" db:"user_name"`
	Title           *string    `json:"title" db:"title"`
	AudioURL        string     `json:"audio_url" db:"audio_url"`
	Text            string     `json:"text" db:"text"`
	DurationSeconds int        `json:"duration_seconds" db:"duration_seconds"`
	Latitude        *float64   `json:"-" db:"latitude"`
	Longitude       *float64   `json:"-" db:"longitude"`
	LocationAccuracy *float64  `json:"-" db:"location_accuracy"`
	Address         *string    `json:"-" db:"address"`
	ParkName        *string    `json:"park_name" db:"park_name"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	Location        *Location  `json:"location,omitempty" db:"-"`
}

// MemoListItem represents a memo in list views
type MemoListItem struct {
	MemoID          uuid.UUID  `json:"memo_id"`
	UserID          string     `json:"user_id"`
	UserName        string     `json:"user_name"`
	Title           *string    `json:"title"`
	AudioURL        string     `json:"audio_url"`
	Text            string     `json:"text"`
	DurationSeconds int        `json:"duration_seconds"`
	Location        *Location  `json:"location,omitempty"`
	ParkName        *string    `json:"park_name"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateMemoRequest represents the request to create a memo
type CreateMemoRequest struct {
	Text             string   `form:"text" binding:"required"`
	DurationSeconds  int      `form:"duration_seconds" binding:"required"`
	Latitude         *float64 `form:"latitude"`
	Longitude        *float64 `form:"longitude"`
	LocationAccuracy *float64 `form:"location_accuracy"`
	ParkName         *string  `form:"park_name"`
	Title            *string  `form:"title"`
}

// UpdateMemoRequest represents the request to update a memo
type UpdateMemoRequest struct {
	Title    *string `json:"title"`
	Text     *string `json:"text"`
	ParkName *string `json:"park_name"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage  int  `json:"current_page"`
	TotalPages   int  `json:"total_pages"`
	TotalItems   int  `json:"total_items"`
	ItemsPerPage int  `json:"items_per_page"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// MemosListResponse represents the response for listing memos
type MemosListResponse struct {
	Memos      []MemoListItem     `json:"memos"`
	Pagination PaginationResponse `json:"pagination"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results    []MemoListItem     `json:"results"`
	Query      string             `json:"query"`
	Pagination PaginationResponse `json:"pagination"`
}

// NearbyMemo represents a memo with distance info
type NearbyMemo struct {
	MemoID         uuid.UUID  `json:"memo_id"`
	UserName       string     `json:"user_name"`
	Title          *string    `json:"title"`
	ParkName       *string    `json:"park_name"`
	Location       *Location  `json:"location"`
	DistanceMeters float64    `json:"distance_meters"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NearbyMemosResponse represents nearby memos response
type NearbyMemosResponse struct {
	Memos        []NearbyMemo `json:"memos"`
	Center       Location     `json:"center"`
	RadiusMeters int          `json:"radius_meters"`
	TotalFound   int          `json:"total_found"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

