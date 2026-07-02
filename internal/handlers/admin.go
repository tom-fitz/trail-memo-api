package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tom-fitz/trailmemo-api/internal/middleware"
	"github.com/tom-fitz/trailmemo-api/internal/repository"
)

// AdminHandler handles allowlist management
type AdminHandler struct {
	approvedRepo *repository.ApprovedUserRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(approvedRepo *repository.ApprovedUserRepository) *AdminHandler {
	return &AdminHandler{approvedRepo: approvedRepo}
}

// ListApprovedUsers returns the sign-in allowlist
// GET /api/v1/admin/approved-users
func (h *AdminHandler) ListApprovedUsers(c *gin.Context) {
	approved, err := h.approvedRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error listing approved users",
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"approved_users": approved})
}

type addApprovedUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// AddApprovedUser adds an email to the allowlist (idempotent)
// POST /api/v1/admin/approved-users
func (h *AdminHandler) AddApprovedUser(c *gin.Context) {
	var req addApprovedUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "A valid email is required",
			},
		})
		return
	}

	adminID, _ := middleware.GetUserID(c)
	if err := h.approvedRepo.Add(c.Request.Context(), req.Email, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error adding approved user",
			},
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"email": req.Email})
}

// RemoveApprovedUser removes an email from the allowlist
// DELETE /api/v1/admin/approved-users?email=<email>
func (h *AdminHandler) RemoveApprovedUser(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "email query parameter is required",
			},
		})
		return
	}

	if err := h.approvedRepo.Remove(c.Request.Context(), email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Error removing approved user",
			},
		})
		return
	}
	c.Status(http.StatusNoContent)
}
