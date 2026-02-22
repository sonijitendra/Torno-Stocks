package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error represents a structured error response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success sends a JSON success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created sends a 201 response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// ErrorResponse sends a structured error response
func ErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": Error{Code: code, Message: message},
	})
}

// BadRequest sends 400
func BadRequest(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends 401
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// NotFound sends 404
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError sends 500
func InternalError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
