package services

import (
	"context"
	"strings"

	"tinystock/backend/models"
	"tinystock/backend/repository"
)

// WatchlistService handles watchlist business logic
type WatchlistService struct {
	repo   repository.WatchlistRepository
	stock  *StockService
}

// NewWatchlistService creates a new WatchlistService
func NewWatchlistService(repo repository.WatchlistRepository, stock *StockService) *WatchlistService {
	return &WatchlistService{repo: repo, stock: stock}
}

// List returns watchlist with current quotes for a user
func (s *WatchlistService) List(ctx context.Context, userID string) ([]models.WatchlistItem, []*models.Quote, error) {
	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if len(items) == 0 {
		return items, nil, nil
	}
	symbols := make([]string, len(items))
	for i, w := range items {
		symbols[i] = w.Symbol
	}
	quotes, err := s.stock.GetQuotes(ctx, symbols)
	if err != nil {
		return items, nil, nil // return items without quotes on error
	}
	return items, quotes, nil
}

// Add adds a symbol to user's watchlist
func (s *WatchlistService) Add(ctx context.Context, userID, symbol string) error {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return ErrInvalidSymbol
	}
	// Verify symbol exists
	if _, err := s.stock.GetQuote(ctx, symbol); err != nil {
		return err
	}
	return s.repo.Add(ctx, userID, symbol)
}

// Remove removes a symbol from watchlist
func (s *WatchlistService) Remove(ctx context.Context, userID, symbol string) error {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	return s.repo.Remove(ctx, userID, symbol)
}
