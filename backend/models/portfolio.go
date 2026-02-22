package models

// Holding represents a portfolio holding
type Holding struct {
	ID       int64   `json:"id"`
	UserID   string  `json:"-"`
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	BuyPrice float64 `json:"buyPrice"`
}

// HoldingWithQuote extends Holding with current quote data for P&L calculation
type HoldingWithQuote struct {
	Holding
	CurrentPrice float64 `json:"currentPrice"`
	MarketValue  float64 `json:"marketValue"`
	CostBasis    float64 `json:"costBasis"`
	PnL          float64 `json:"pnl"`
	PnLPercent   float64 `json:"pnlPercent"`
}

// PortfolioSummary holds the full portfolio view
type PortfolioSummary struct {
	Holdings   []HoldingWithQuote `json:"holdings"`
	TotalValue float64           `json:"totalValue"`
	TotalCost  float64           `json:"totalCost"`
	TotalPnL   float64           `json:"totalPnL"`
	ReturnPct  float64           `json:"returnPct"`
}
