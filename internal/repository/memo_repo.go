package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/tom-fitz/trailmemo-api/internal/models"
)

// MemoRepository handles memo database operations
type MemoRepository struct {
	db *sqlx.DB
}

// NewMemoRepository creates a new memo repository
func NewMemoRepository(db *sqlx.DB) *MemoRepository {
	return &MemoRepository{db: db}
}

// Create creates a new memo
func (r *MemoRepository) Create(ctx context.Context, memo *models.Memo) error {
	query := `
		INSERT INTO memos (
			user_id, user_name, title, audio_url, text, duration_seconds,
			latitude, longitude, location_accuracy, address, park_name
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING memo_id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		memo.UserID,
		memo.UserName,
		memo.Title,
		memo.AudioURL,
		memo.Text,
		memo.DurationSeconds,
		memo.Latitude,
		memo.Longitude,
		memo.LocationAccuracy,
		memo.Address,
		memo.ParkName,
	).Scan(&memo.MemoID, &memo.CreatedAt, &memo.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating memo: %v", err)
	}

	return nil
}

// GetByID retrieves a memo by its ID
func (r *MemoRepository) GetByID(ctx context.Context, memoID uuid.UUID) (*models.Memo, error) {
	var memo models.Memo
	query := `
		SELECT 
			memo_id, user_id, user_name, title, audio_url, text, duration_seconds,
			latitude, longitude, location_accuracy, address, park_name,
			created_at, updated_at
		FROM memos
		WHERE memo_id = $1
	`

	err := r.db.GetContext(ctx, &memo, query, memoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting memo: %v", err)
	}

	// Populate location if coordinates exist
	if memo.Latitude != nil && memo.Longitude != nil {
		memo.Location = &models.Location{
			Latitude:  *memo.Latitude,
			Longitude: *memo.Longitude,
			Accuracy:  memo.LocationAccuracy,
			Address:   memo.Address,
		}
	}

	return &memo, nil
}

// List retrieves all memos with pagination and optional filters
func (r *MemoRepository) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]models.MemoListItem, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if parkName, ok := filters["park_name"].(string); ok && parkName != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("park_name = $%d", argPos))
		args = append(args, parkName)
		argPos++
	}

	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, userID)
		argPos++
	}

	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, startDate)
		argPos++
	}

	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argPos))
		args = append(args, endDate)
		argPos++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM memos %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting memos: %v", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Query memos
	query := fmt.Sprintf(`
		SELECT 
			memo_id, user_id, user_name, title, audio_url, text, duration_seconds,
			latitude, longitude, location_accuracy, address, park_name,
			created_at, updated_at
		FROM memos
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying memos: %v", err)
	}
	defer rows.Close()

	memos := []models.MemoListItem{}
	for rows.Next() {
		var m models.Memo
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("error scanning memo: %v", err)
		}

		// Build location if coordinates exist
		var location *models.Location
		if m.Latitude != nil && m.Longitude != nil {
			location = &models.Location{
				Latitude:  *m.Latitude,
				Longitude: *m.Longitude,
				Accuracy:  m.LocationAccuracy,
				Address:   m.Address,
			}
		}

		memos = append(memos, models.MemoListItem{
			MemoID:          m.MemoID,
			UserID:          m.UserID,
			UserName:        m.UserName,
			Title:           m.Title,
			AudioURL:        m.AudioURL,
			Text:            m.Text,
			DurationSeconds: m.DurationSeconds,
			Location:        location,
			ParkName:        m.ParkName,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		})
	}

	return memos, total, nil
}

// Update updates a memo
func (r *MemoRepository) Update(ctx context.Context, memoID uuid.UUID, updates map[string]interface{}) (*models.Memo, error) {
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if title, ok := updates["title"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argPos))
		args = append(args, title)
		argPos++
	}

	if text, ok := updates["text"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("text = $%d", argPos))
		args = append(args, text)
		argPos++
	}

	if parkName, ok := updates["park_name"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("park_name = $%d", argPos))
		args = append(args, parkName)
		argPos++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE memos
		SET %s
		WHERE memo_id = $%d
	`, strings.Join(setClauses, ", "), argPos)

	args = append(args, memoID)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error updating memo: %v", err)
	}

	// Fetch and return updated memo
	return r.GetByID(ctx, memoID)
}

// Delete deletes a memo
func (r *MemoRepository) Delete(ctx context.Context, memoID uuid.UUID) error {
	query := `DELETE FROM memos WHERE memo_id = $1`

	result, err := r.db.ExecContext(ctx, query, memoID)
	if err != nil {
		return fmt.Errorf("error deleting memo: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rows == 0 {
		return fmt.Errorf("memo not found")
	}

	return nil
}

// SearchByText performs full-text search on memos
func (r *MemoRepository) SearchByText(ctx context.Context, query string, page, limit int) ([]models.MemoListItem, int, error) {
	// Count total matches
	countQuery := `
		SELECT COUNT(*)
		FROM memos
		WHERE to_tsvector('english', text) @@ plainto_tsquery('english', $1)
	`
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, query)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting search results: %v", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Search query
	searchQuery := `
		SELECT 
			memo_id, user_id, user_name, title, audio_url, text, duration_seconds,
			latitude, longitude, location_accuracy, address, park_name,
			created_at, updated_at,
			ts_rank(to_tsvector('english', text), plainto_tsquery('english', $1)) as rank
		FROM memos
		WHERE to_tsvector('english', text) @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryxContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error searching memos: %v", err)
	}
	defer rows.Close()

	memos := []models.MemoListItem{}
	for rows.Next() {
		var m models.Memo
		var rank float64

		if err := rows.Scan(
			&m.MemoID, &m.UserID, &m.UserName, &m.Title, &m.AudioURL, &m.Text,
			&m.DurationSeconds, &m.Latitude, &m.Longitude, &m.LocationAccuracy,
			&m.Address, &m.ParkName, &m.CreatedAt, &m.UpdatedAt, &rank,
		); err != nil {
			return nil, 0, fmt.Errorf("error scanning memo: %v", err)
		}

		// Build location if coordinates exist
		var location *models.Location
		if m.Latitude != nil && m.Longitude != nil {
			location = &models.Location{
				Latitude:  *m.Latitude,
				Longitude: *m.Longitude,
				Accuracy:  m.LocationAccuracy,
				Address:   m.Address,
			}
		}

		memos = append(memos, models.MemoListItem{
			MemoID:          m.MemoID,
			UserID:          m.UserID,
			UserName:        m.UserName,
			Title:           m.Title,
			AudioURL:        m.AudioURL,
			Text:            m.Text,
			DurationSeconds: m.DurationSeconds,
			Location:        location,
			ParkName:        m.ParkName,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		})
	}

	return memos, total, nil
}

// GetNearby finds memos near a location using Haversine formula
func (r *MemoRepository) GetNearby(ctx context.Context, lat, lon float64, radiusMeters, limit int) ([]models.NearbyMemo, error) {
	// Haversine formula in SQL - use subquery to filter by distance
	query := `
		SELECT 
			memo_id, user_name, title, park_name,
			latitude, longitude, location_accuracy, address,
			created_at, distance_meters
		FROM (
			SELECT 
				memo_id, user_name, title, park_name,
				latitude, longitude, location_accuracy, address,
				created_at,
				(
					6371000 * acos(
						cos(radians($1)) * cos(radians(latitude)) *
						cos(radians(longitude) - radians($2)) +
						sin(radians($1)) * sin(radians(latitude))
					)
				) AS distance_meters
			FROM memos
			WHERE latitude IS NOT NULL AND longitude IS NOT NULL
		) AS nearby
		WHERE distance_meters <= $3
		ORDER BY distance_meters ASC
		LIMIT $4
	`

	rows, err := r.db.QueryxContext(ctx, query, lat, lon, radiusMeters, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying nearby memos: %v", err)
	}
	defer rows.Close()

	nearbyMemos := []models.NearbyMemo{}
	for rows.Next() {
		var nm models.NearbyMemo
		var lat, lon float64
		var accuracy *float64
		var address *string

		if err := rows.Scan(
			&nm.MemoID, &nm.UserName, &nm.Title, &nm.ParkName,
			&lat, &lon, &accuracy, &address,
			&nm.CreatedAt, &nm.DistanceMeters,
		); err != nil {
			return nil, fmt.Errorf("error scanning nearby memo: %v", err)
		}

		nm.Location = &models.Location{
			Latitude:  lat,
			Longitude: lon,
			Accuracy:  accuracy,
			Address:   address,
		}

		// Round distance to 2 decimal places
		nm.DistanceMeters = math.Round(nm.DistanceMeters*100) / 100

		nearbyMemos = append(nearbyMemos, nm)
	}

	return nearbyMemos, nil
}
