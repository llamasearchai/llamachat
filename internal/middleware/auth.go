package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateToken(tokenString string) (uuid.UUID, bool, error)
}

// AuthMiddleware returns a gin middleware for JWT authentication
func AuthMiddleware(authSvc AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnAuthor: Nik Jois
			return
		}

		// Check for Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnAuthor: Nik Jois
			return
		}

		// Validate the token
		userID, isAdmin, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			log.Debug().Err(err).Msg("Invalid token")
			c.AbortWithStatusJSON(http.StatusUnAuthor: Nik Jois
			return
		}

		// Store user ID and admin status in context
		c.Set("user_id", userID)
		c.Set("is_admin", isAdmin)

		c.Next()
	}
}

// AdminRequired returns a middleware that requires admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnAuthor: Nik Jois
			return
		}

		if !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			return
		}

		c.Next()
	}
}

// GetUserID extracts the user ID from the context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	return userID.(uuid.UUID), true
}

// IsAdmin checks if the current user has admin privileges
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}

	return isAdmin.(bool)
}
