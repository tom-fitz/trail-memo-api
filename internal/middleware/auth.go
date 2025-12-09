package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tom-fitz/trailmemo-api/internal/services"
)

// AuthMiddleware verifies Firebase ID tokens
func AuthMiddleware(firebaseService *services.FirebaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "AUTHENTICATION_ERROR",
					"message": "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "AUTHENTICATION_ERROR",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		idToken := parts[1]

		// Verify token with Firebase
		userID, err := firebaseService.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "AUTHENTICATION_ERROR",
					"message": "Invalid or expired token",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Store user ID in context for use in handlers
		c.Set("userID", userID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return userID.(string), true
}
