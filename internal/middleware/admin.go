package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tom-fitz/trailmemo-api/internal/repository"
)

// AdminMiddleware requires the authenticated user to be an admin.
// Must run after AuthMiddleware.
func AdminMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "AUTHENTICATION_ERROR",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Error fetching user",
				},
			})
			c.Abort()
			return
		}

		if user == nil || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Admin access required",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
