package models

// Quote represents a stock quote
type Quote struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePercent"`
	Volume    int64   `json:"volume"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
}

// HistoryPoint represents a single point in price history
type HistoryPoint struct {
	Date   string  `json:"date"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}
