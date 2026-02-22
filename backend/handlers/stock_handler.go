package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"tinystock/backend/internal/response"
	"tinystock/backend/services"
)

// StockHandler handles stock API endpoints (public, no auth)
type StockHandler struct {
	stock *services.StockService
}

// NewStockHandler creates a new StockHandler
func NewStockHandler(stock *services.StockService) *StockHandler {
	return &StockHandler{stock: stock}
}

// GetQuote handles GET /api/quote/:symbol
func (h *StockHandler) GetQuote(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Param("symbol")))
	if symbol == "" {
		response.BadRequest(c, "Symbol is required")
		return
	}
	quote, err := h.stock.GetQuote(c.Request.Context(), symbol)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, quote)
}

// GetHistory handles GET /api/history/:symbol
func (h *StockHandler) GetHistory(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Param("symbol")))
	if symbol == "" {
		response.BadRequest(c, "Symbol is required")
		return
	}
	range_ := c.DefaultQuery("range", "1mo")
	interval := c.DefaultQuery("interval", "1d")
	history, err := h.stock.GetHistory(c.Request.Context(), symbol, range_, interval)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"symbol": symbol, "history": history})
}

// Search handles GET /api/search
func (h *StockHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.BadRequest(c, "Query (q) is required")
		return
	}
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}
	results, err := h.stock.SearchSymbols(c.Request.Context(), query, limit)
	if err != nil {
		response.InternalError(c, "Search failed")
		return
	}
	response.Success(c, gin.H{"results": results})
}
