package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration from environment variables
type Config struct {
	Port        string
	DBDriver    string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration
	RateLimit   int
	CORSOrigins string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" && dbDriver == "sqlite" {
		databaseURL = "./tinystock.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	jwtExpiryStr := os.Getenv("JWT_EXPIRY")
	if jwtExpiryStr == "" {
		jwtExpiryStr = "24h"
	}
	jwtExpiry, _ := time.ParseDuration(jwtExpiryStr)

	rateLimit := 100
	if rl := os.Getenv("RATE_LIMIT"); rl != "" {
		if n, err := strconv.Atoi(rl); err == nil && n > 0 {
			rateLimit = n
		}
	}

	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	return &Config{
		Port:        port,
		DBDriver:    dbDriver,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		JWTExpiry:   jwtExpiry,
		RateLimit:   rateLimit,
		CORSOrigins: corsOrigins,
	}, nil
}
