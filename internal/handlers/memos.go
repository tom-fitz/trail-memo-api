package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tom-fitz/trailmemo-api/internal/middleware"
	"github.com/tom-fitz/trailmemo-api/internal/models"
	"github.com/tom-fitz/trailmemo-api/internal/repository"
	"github.com/tom-fitz/trailmemo-api/internal/services"
)

// MemoHandler handles memo-related requests
type MemoHandler struct {
	memoRepo        *repository.MemoRepository
	userRepo        *repository.UserRepository
	firebaseService *services.FirebaseService
	maxUploadSize   int64
}

// NewMemoHandler creates a new memo handler
func NewMemoHandler(
	memoRepo *repository.MemoRepository,
	userRepo *repository.UserRepository,
	firebaseService *services.FirebaseService,
	maxUploadSize int64,
) *MemoHandler {
	return &MemoHandler{
		memoRepo:        memoRepo,
		userRepo:        userRepo,
		firebaseService: firebaseService,
		maxUploadSize:   maxUploadSize,
	}
}

// Create creates a new memo with audio upload
// POST /api/v1/memos
func (h *MemoHandler) Create(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_ERROR",
				"message": "Authentication required",
			},
		})
		return
	}

	// Get user info
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching user information",
			},
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(h.maxUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Error parsing form data",
				"details": gin.H{
					"reason": err.Error(),
				},
			},
		})
		return
	}

	// Get audio file (optional for MVP)
	audioFile, err := c.FormFile("audio")
	var audioURL string

	if err == nil {
		// Audio file provided - validate size
		if audioFile.Size > h.maxUploadSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "File size exceeds maximum allowed size",
					"details": gin.H{
						"max_size_mb": h.maxUploadSize / (1024 * 1024),
					},
				},
			})
			return
		}

		// Upload audio file to Firebase Storage
		audioURL, err = h.firebaseService.UploadAudioFile(c.Request.Context(), audioFile, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Error uploading audio file",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			return
		}
	} else {
		// No audio file provided - use placeholder for MVP
		audioURL = "https://placeholder.com/audio.m4a"
	}

	// Parse form data
	var req models.CreateMemoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid form data",
				"details": gin.H{
					"reason": err.Error(),
				},
			},
		})
		return
	}

	// Create memo in database
	memo := &models.Memo{
		UserID:           userID,
		UserName:         user.DisplayName,
		Title:            req.Title,
		AudioURL:         audioURL,
		Text:             req.Text,
		DurationSeconds:  req.DurationSeconds,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		LocationAccuracy: req.LocationAccuracy,
		ParkName:         req.ParkName,
	}

	if err := h.memoRepo.Create(c.Request.Context(), memo); err != nil {
		// Try to delete uploaded file on failure
		_ = h.firebaseService.DeleteAudioFile(c.Request.Context(), audioURL)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error creating memo",
			},
		})
		return
	}

	// Build location object if coordinates exist
	if memo.Latitude != nil && memo.Longitude != nil {
		memo.Location = &models.Location{
			Latitude:  *memo.Latitude,
			Longitude: *memo.Longitude,
			Accuracy:  memo.LocationAccuracy,
			Address:   memo.Address,
		}
	}

	c.JSON(http.StatusCreated, memo)
}

// List retrieves all memos with optional filters
// GET /api/v1/memos
func (h *MemoHandler) List(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}

	// Build filters
	filters := make(map[string]interface{})
	if parkName := c.Query("park_name"); parkName != "" {
		filters["park_name"] = parkName
	}
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}

	// Fetch memos
	memos, total, err := h.memoRepo.List(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching memos",
			},
		})
		return
	}

	// Build pagination response
	totalPages := (total + limit - 1) / limit
	pagination := models.PaginationResponse{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: limit,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}

	c.JSON(http.StatusOK, models.MemosListResponse{
		Memos:      memos,
		Pagination: pagination,
	})
}

// GetByID retrieves a specific memo
// GET /api/v1/memos/:id
func (h *MemoHandler) GetByID(c *gin.Context) {
	// Parse memo ID
	memoIDStr := c.Param("id")
	memoID, err := uuid.Parse(memoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid memo ID",
			},
		})
		return
	}

	// Fetch memo
	memo, err := h.memoRepo.GetByID(c.Request.Context(), memoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching memo",
			},
		})
		return
	}

	if memo == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Memo not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, memo)
}

// Update updates a memo
// PUT /api/v1/memos/:id
func (h *MemoHandler) Update(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_ERROR",
				"message": "Authentication required",
			},
		})
		return
	}

	// Parse memo ID
	memoIDStr := c.Param("id")
	memoID, err := uuid.Parse(memoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid memo ID",
			},
		})
		return
	}

	// Fetch memo to check ownership
	memo, err := h.memoRepo.GetByID(c.Request.Context(), memoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching memo",
			},
		})
		return
	}

	if memo == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Memo not found",
			},
		})
		return
	}

	// Check if user owns the memo
	if memo.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "AUTHORIZATION_ERROR",
				"message": "You can only update your own memos",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateMemoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = req.Title
	}
	if req.Text != nil {
		updates["text"] = req.Text
	}
	if req.ParkName != nil {
		updates["park_name"] = req.ParkName
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "No fields to update",
			},
		})
		return
	}

	// Update memo
	updatedMemo, err := h.memoRepo.Update(c.Request.Context(), memoID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error updating memo",
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedMemo)
}

// Delete deletes a memo
// DELETE /api/v1/memos/:id
func (h *MemoHandler) Delete(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_ERROR",
				"message": "Authentication required",
			},
		})
		return
	}

	// Parse memo ID
	memoIDStr := c.Param("id")
	memoID, err := uuid.Parse(memoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid memo ID",
			},
		})
		return
	}

	// Fetch memo to check ownership
	memo, err := h.memoRepo.GetByID(c.Request.Context(), memoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching memo",
			},
		})
		return
	}

	if memo == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Memo not found",
			},
		})
		return
	}

	// Check if user owns the memo
	if memo.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "AUTHORIZATION_ERROR",
				"message": "You can only delete your own memos",
			},
		})
		return
	}

	// Delete audio file from storage
	if err := h.firebaseService.DeleteAudioFile(c.Request.Context(), memo.AudioURL); err != nil {
		// Log error but continue with database deletion
		// In production, you might want to queue this for retry
	}

	// Delete memo from database
	if err := h.memoRepo.Delete(c.Request.Context(), memoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error deleting memo",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetNearby finds memos near a location
// GET /api/v1/memos/nearby
func (h *MemoHandler) GetNearby(c *gin.Context) {
	// Parse query parameters
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius_meters", "1000")
	limitStr := c.DefaultQuery("limit", "50")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "latitude and longitude are required",
			},
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid latitude value",
			},
		})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid longitude value",
			},
		})
		return
	}

	radius, err := strconv.Atoi(radiusStr)
	if err != nil || radius < 0 || radius > 50000 {
		radius = 1000
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}

	// Fetch nearby memos
	memos, err := h.memoRepo.GetNearby(c.Request.Context(), lat, lon, radius, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching nearby memos",
			},
		})
		return
	}

	response := models.NearbyMemosResponse{
		Memos: memos,
		Center: models.Location{
			Latitude:  lat,
			Longitude: lon,
		},
		RadiusMeters: radius,
		TotalFound:   len(memos),
	}

	c.JSON(http.StatusOK, response)
}

// Search performs full-text search on memos
// GET /api/v1/memos/search
func (h *MemoHandler) Search(c *gin.Context) {
	// Parse query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Search query (q) is required",
			},
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Perform search
	memos, total, err := h.memoRepo.SearchByText(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error performing search",
			},
		})
		return
	}

	// Build pagination response
	totalPages := (total + limit - 1) / limit
	pagination := models.PaginationResponse{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: limit,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Results:    memos,
		Query:      query,
		Pagination: pagination,
	})
}
