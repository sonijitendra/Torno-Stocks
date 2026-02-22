package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
	"tinystock/backend/models"
	"tinystock/backend/services"
)

// AuthHandler handles auth endpoints
type AuthHandler struct {
	auth *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: email and password (min 6 chars) required")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user, err := h.auth.Register(c.Request.Context(), &req)
	if err != nil {
		if err == services.ErrUserExists {
			response.ErrorResponse(c, http.StatusConflict, "USER_EXISTS", "Email already registered")
			return
		}
		response.InternalError(c, "Registration failed")
		return
	}
	response.Created(c, user)
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: email and password required")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	res, err := h.auth.Login(c.Request.Context(), &req)
	if err != nil {
		if err == services.ErrInvalidCreds {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		response.InternalError(c, "Login failed")
		return
	}
	response.Success(c, res)
}
