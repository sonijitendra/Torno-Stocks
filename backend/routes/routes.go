package routes

import (
	"github.com/gin-gonic/gin"
	"tinystock/backend/config"
	"tinystock/backend/handlers"
	"tinystock/backend/middleware"
	"tinystock/backend/services"
)

// Setup configures all routes
func Setup(r *gin.Engine, cfg *config.Config, deps *Dependencies) {
	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.NewRateLimiter(cfg.RateLimit, cfg.RateLimit*2).Middleware())

	// Public API (no auth)
	api := r.Group("/api")
	{
		// Auth
		api.POST("/auth/register", deps.AuthHandler.Register)
		api.POST("/auth/login", deps.AuthHandler.Login)

		// Stock (public - proxy through backend only)
		api.GET("/quote/:symbol", deps.StockHandler.GetQuote)
		api.GET("/history/:symbol", deps.StockHandler.GetHistory)
		api.GET("/search", deps.StockHandler.Search)
	}

	// Protected API (JWT required)
	protected := api.Group("")
	protected.Use(middleware.Auth(deps.AuthService))
	{
		protected.GET("/watchlist", deps.WatchlistHandler.List)
		protected.POST("/watchlist", deps.WatchlistHandler.Add)
		protected.DELETE("/watchlist/:symbol", deps.WatchlistHandler.Remove)

		protected.GET("/portfolio", deps.PortfolioHandler.List)
		protected.POST("/portfolio", deps.PortfolioHandler.Add)
		protected.DELETE("/portfolio/:id", deps.PortfolioHandler.Remove)
	}
}

// Dependencies holds all route dependencies
type Dependencies struct {
	AuthHandler     *handlers.AuthHandler
	StockHandler    *handlers.StockHandler
	WatchlistHandler *handlers.WatchlistHandler
	PortfolioHandler *handlers.PortfolioHandler
	AuthService     *services.AuthService
}
