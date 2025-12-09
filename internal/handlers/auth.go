package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tom-fitz/trailmemo-api/internal/middleware"
	"github.com/tom-fitz/trailmemo-api/internal/models"
	"github.com/tom-fitz/trailmemo-api/internal/repository"
	"github.com/tom-fitz/trailmemo-api/internal/services"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo        *repository.UserRepository
	firebaseService *services.FirebaseService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, firebaseService *services.FirebaseService) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		firebaseService: firebaseService,
	}
}

// Register creates a new user account
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// Get authenticated user ID from Firebase token
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

	// Parse request body
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
				"details": gin.H{
					"reason": err.Error(),
				},
			},
		})
		return
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error checking user existence",
			},
		})
		return
	}

	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "CONFLICT",
				"message": "User already exists",
			},
		})
		return
	}

	// Get user info from Firebase
	firebaseUser, err := h.firebaseService.GetUserByUID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching user from Firebase",
			},
		})
		return
	}

	// Create user in database
	user := &models.User{
		UserID:      userID,
		Email:       firebaseUser.Email,
		DisplayName: req.DisplayName,
		Department:  req.Department,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error creating user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetMe returns information about the currently authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
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

	// Fetch user from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error fetching user",
			},
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
