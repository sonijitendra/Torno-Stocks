package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
)

// Recovery returns a panic recovery middleware
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				response.InternalError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
