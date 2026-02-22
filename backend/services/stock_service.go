package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tinystock/backend/models"
)

// StockService provides stock data with caching and context timeout
type StockService struct {
	yahoo *YahooFinanceClient
	cache *MemoryCache
}

// NewStockService creates a new StockService
func NewStockService() *StockService {
	return &StockService{
		yahoo: NewYahooFinanceClient(),
		cache: NewMemoryCache(2 * time.Minute), // 2 min cache for quotes
	}
}

// GetQuote fetches quote with cache and context timeout
func (s *StockService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	cacheKey := "quote:" + symbol
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.(*models.Quote), nil
	}

	// Use context with timeout for Yahoo API call
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	quote, err := s.yahoo.GetQuoteWithContext(ctx, symbol)
	if err != nil {
		return nil, err
	}
	s.cache.Set(cacheKey, quote)
	return quote, nil
}

// GetHistory fetches 30-day history with cache
func (s *StockService) GetHistory(ctx context.Context, symbol string, range_, interval string) ([]models.HistoryPoint, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if range_ == "" {
		range_ = "1mo"
	}
	if interval == "" {
		interval = "1d"
	}

	cacheKey := fmt.Sprintf("history:%s:%s:%s", symbol, range_, interval)
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.([]models.HistoryPoint), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	history, err := s.yahoo.GetHistoryWithContext(ctx, symbol, range_, interval)
	if err != nil {
		return nil, err
	}
	s.cache.Set(cacheKey, history)
	return history, nil
}

// GetQuotes fetches multiple quotes (for watchlist/portfolio)
func (s *StockService) GetQuotes(ctx context.Context, symbols []string) ([]*models.Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	// Build symbol list and check cache
	var toFetch []string
	quoteMap := make(map[string]*models.Quote)
	for _, sym := range symbols {
		sym = strings.ToUpper(strings.TrimSpace(sym))
		if v, ok := s.cache.Get("quote:" + sym); ok {
			quoteMap[sym] = v.(*models.Quote)
		} else {
			toFetch = append(toFetch, sym)
		}
	}
	if len(toFetch) > 0 {
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		quotes, err := s.yahoo.GetQuotesWithContext(ctx, toFetch)
		if err != nil {
			return nil, err
		}
		for _, q := range quotes {
			s.cache.Set("quote:"+q.Symbol, q)
			quoteMap[q.Symbol] = q
		}
	}
	// Return in original symbol order
	result := make([]*models.Quote, 0, len(symbols))
	for _, sym := range symbols {
		sym = strings.ToUpper(strings.TrimSpace(sym))
		if q, ok := quoteMap[sym]; ok {
			result = append(result, q)
		}
	}
	return result, nil
}

// SearchSymbols searches for stock symbols
func (s *StockService) SearchSymbols(ctx context.Context, query string, limit int) ([]models.Quote, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.yahoo.SearchSymbolsWithContext(ctx, query, limit)
}
