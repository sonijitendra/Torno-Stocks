package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
	"tinystock/backend/services"
)

const UserIDKey = "user_id"

// Auth returns a JWT auth middleware
func Auth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}
		userID, err := authService.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID extracts user ID from context (must be used after Auth middleware)
func GetUserID(c *gin.Context) string {
	v, _ := c.Get(UserIDKey)
	if id, ok := v.(string); ok {
		return id
	}
	return ""
}
