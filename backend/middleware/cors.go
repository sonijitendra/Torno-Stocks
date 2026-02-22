package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware
func CORS(origins string) gin.HandlerFunc {
	allowOrigins := strings.Split(origins, ",")
	for i, o := range allowOrigins {
		allowOrigins[i] = strings.TrimSpace(o)
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := "*"
		if len(allowOrigins) > 0 && allowOrigins[0] != "*" {
			for _, o := range allowOrigins {
				if o == origin || o == "*" {
					allowed = origin
					if o == "*" {
						allowed = "*"
					}
					break
				}
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowed)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
