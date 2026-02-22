package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"tinystock/backend/config"
	"tinystock/backend/handlers"
	"tinystock/backend/repository"
	"tinystock/backend/routes"
	"tinystock/backend/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config:", err)
	}

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatal("database:", err)
	}
	defer db.Close()

	authService := services.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiry)
	stockService := services.NewStockService()
	watchlistService := services.NewWatchlistService(db, stockService)
	portfolioService := services.NewPortfolioService(db, stockService)

	deps := &routes.Dependencies{
		AuthHandler:      handlers.NewAuthHandler(authService),
		StockHandler:     handlers.NewStockHandler(stockService),
		WatchlistHandler: handlers.NewWatchlistHandler(watchlistService),
		PortfolioHandler: handlers.NewPortfolioHandler(portfolioService),
		AuthService:     authService,
	}

	if cfg.DBDriver == "postgres" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	routes.Setup(r, cfg, deps)

	log.Printf("TinyStock API listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
