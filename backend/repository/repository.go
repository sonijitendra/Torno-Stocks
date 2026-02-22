package repository

import (
	"context"

	"tinystock/backend/models"
)

// UserRepository defines user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

// WatchlistRepository defines watchlist data access
type WatchlistRepository interface {
	Add(ctx context.Context, userID, symbol string) error
	Remove(ctx context.Context, userID, symbol string) error
	List(ctx context.Context, userID string) ([]models.WatchlistItem, error)
}

// PortfolioRepository defines portfolio data access
type PortfolioRepository interface {
	AddHolding(ctx context.Context, userID string, symbol string, quantity, buyPrice float64) error
	RemoveHolding(ctx context.Context, userID string, id int64) error
	ListHoldings(ctx context.Context, userID string) ([]models.Holding, error)
}

// DB wraps all repositories
type DB interface {
	UserRepository
	WatchlistRepository
	PortfolioRepository
	Close() error
}
