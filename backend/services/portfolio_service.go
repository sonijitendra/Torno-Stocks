package services

import (
	"context"
	"errors"
	"strings"

	"tinystock/backend/models"
	"tinystock/backend/repository"
)

var ErrInvalidSymbol = errors.New("invalid symbol")

// PortfolioService handles portfolio business logic and P&L calculations
type PortfolioService struct {
	repo  repository.PortfolioRepository
	stock *StockService
}

// NewPortfolioService creates a new PortfolioService
func NewPortfolioService(repo repository.PortfolioRepository, stock *StockService) *PortfolioService {
	return &PortfolioService{repo: repo, stock: stock}
}

// GetPortfolio returns full portfolio with real-time P&L for a user
func (s *PortfolioService) GetPortfolio(ctx context.Context, userID string) (*models.PortfolioSummary, error) {
	holdings, err := s.repo.ListHoldings(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(holdings) == 0 {
		return &models.PortfolioSummary{Holdings: []models.HoldingWithQuote{}}, nil
	}

	symbols := make([]string, len(holdings))
	for i, h := range holdings {
		symbols[i] = h.Symbol
	}
	quotes, err := s.stock.GetQuotes(ctx, symbols)
	if err != nil {
		quotes = nil
	}
	quoteMap := make(map[string]float64)
	for _, q := range quotes {
		quoteMap[q.Symbol] = q.Price
	}

	var totalValue, totalCost float64
	withQuotes := make([]models.HoldingWithQuote, len(holdings))
	for i, h := range holdings {
		currentPrice := quoteMap[h.Symbol]
		if currentPrice == 0 {
			currentPrice = h.BuyPrice
		}
		marketValue := h.Quantity * currentPrice
		costBasis := h.Quantity * h.BuyPrice
		pnl := marketValue - costBasis
		pnlPercent := 0.0
		if costBasis > 0 {
			pnlPercent = (pnl / costBasis) * 100
		}
		withQuotes[i] = models.HoldingWithQuote{
			Holding:      h,
			CurrentPrice: currentPrice,
			MarketValue:  marketValue,
			CostBasis:    costBasis,
			PnL:          pnl,
			PnLPercent:   pnlPercent,
		}
		totalValue += marketValue
		totalCost += costBasis
	}

	returnPct := 0.0
	if totalCost > 0 {
		returnPct = ((totalValue - totalCost) / totalCost) * 100
	}

	return &models.PortfolioSummary{
		Holdings:   withQuotes,
		TotalValue: totalValue,
		TotalCost:  totalCost,
		TotalPnL:   totalValue - totalCost,
		ReturnPct:  returnPct,
	}, nil
}

// AddHolding adds a holding to user's portfolio
func (s *PortfolioService) AddHolding(ctx context.Context, userID string, symbol string, quantity, buyPrice float64) error {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" || quantity <= 0 || buyPrice <= 0 {
		return ErrInvalidSymbol
	}
	if _, err := s.stock.GetQuote(ctx, symbol); err != nil {
		return err
	}
	return s.repo.AddHolding(ctx, userID, symbol, quantity, buyPrice)
}

// RemoveHolding removes a holding
func (s *PortfolioService) RemoveHolding(ctx context.Context, userID string, id int64) error {
	return s.repo.RemoveHolding(ctx, userID, id)
}
