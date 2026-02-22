package models

// WatchlistItem represents a stock in the watchlist
type WatchlistItem struct {
	ID     int64  `json:"id"`
	UserID string `json:"-"`
	Symbol string `json:"symbol"`
}
