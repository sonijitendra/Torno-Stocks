package repository

import (
	"fmt"

	"tinystock/backend/config"
	"tinystock/backend/repository/postgres"
	"tinystock/backend/repository/sqlite"
)

// NewDB creates a database connection based on config
func NewDB(cfg *config.Config) (DB, error) {
	switch cfg.DBDriver {
	case "sqlite":
		return sqlite.New(cfg.DatabaseURL)
	case "postgres":
		return postgres.New(cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.DBDriver)
	}
}
