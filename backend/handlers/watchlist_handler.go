package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
	"tinystock/backend/middleware"
	"tinystock/backend/services"
)

// WatchlistHandler handles watchlist endpoints (requires auth)
type WatchlistHandler struct {
	watchlist *services.WatchlistService
}

// NewWatchlistHandler creates a new WatchlistHandler
func NewWatchlistHandler(watchlist *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlist: watchlist}
}

// List handles GET /api/watchlist
func (h *WatchlistHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	items, quotes, err := h.watchlist.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get watchlist")
		return
	}
	response.Success(c, gin.H{"watchlist": items, "quotes": quotes})
}

// Add handles POST /api/watchlist
func (h *WatchlistHandler) Add(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Symbol is required")
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" {
		response.BadRequest(c, "Symbol is required")
		return
	}
	userID := middleware.GetUserID(c)
	err := h.watchlist.Add(c.Request.Context(), userID, symbol)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			response.ErrorResponse(c, http.StatusConflict, "ALREADY_IN_WATCHLIST", "Symbol already in watchlist")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, gin.H{"message": "added", "symbol": symbol})
}

// Remove handles DELETE /api/watchlist/:symbol
func (h *WatchlistHandler) Remove(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Param("symbol")))
	if symbol == "" {
		response.BadRequest(c, "Symbol is required")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.watchlist.Remove(c.Request.Context(), userID, symbol); err != nil {
		response.InternalError(c, "Failed to remove")
		return
	}
	response.Success(c, gin.H{"message": "removed", "symbol": symbol})
}
