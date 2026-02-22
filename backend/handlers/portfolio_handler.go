package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
	"tinystock/backend/middleware"
	"tinystock/backend/services"
)

// PortfolioHandler handles portfolio endpoints (requires auth)
type PortfolioHandler struct {
	portfolio *services.PortfolioService
}

// NewPortfolioHandler creates a new PortfolioHandler
func NewPortfolioHandler(portfolio *services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{portfolio: portfolio}
}

// List handles GET /api/portfolio
func (h *PortfolioHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	summary, err := h.portfolio.GetPortfolio(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get portfolio")
		return
	}
	response.Success(c, summary)
}

// Add handles POST /api/portfolio
func (h *PortfolioHandler) Add(c *gin.Context) {
	var req struct {
		Symbol   string  `json:"symbol"`
		Quantity float64 `json:"quantity"`
		BuyPrice float64 `json:"buyPrice"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "symbol, quantity, and buyPrice are required")
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" || req.Quantity <= 0 || req.BuyPrice <= 0 {
		response.BadRequest(c, "symbol, quantity, and buyPrice must be positive")
		return
	}
	userID := middleware.GetUserID(c)
	err := h.portfolio.AddHolding(c.Request.Context(), userID, symbol, req.Quantity, req.BuyPrice)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, gin.H{"message": "added", "symbol": symbol})
}

// Remove handles DELETE /api/portfolio/:id
func (h *PortfolioHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid holding id")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.portfolio.RemoveHolding(c.Request.Context(), userID, id); err != nil {
		response.InternalError(c, "Failed to remove")
		return
	}
	response.Success(c, gin.H{"message": "removed"})
}
