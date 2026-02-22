package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a request logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method

		c.Next()

		statusCode := c.Writer.Status()
		latency := time.Since(start)
		log.Printf("[%s] %d %s %s %v", method, statusCode, path, clientIP, latency)
	}
}
