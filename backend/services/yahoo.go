package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"tinystock/backend/models"
)

// YahooFinanceClient fetches stock data from Yahoo Finance
type YahooFinanceClient struct {
	client *http.Client
}

// NewYahooFinanceClient creates a new Yahoo Finance client
func NewYahooFinanceClient() *YahooFinanceClient {
	return &YahooFinanceClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
	}
}

// doRequest performs HTTP GET with browser-like headers (Yahoo may block default Go user-agent)
func (c *YahooFinanceClient) doRequest(u string) (*http.Response, error) {
	return c.doRequestWithContext(context.Background(), u)
}

// doRequestWithContext performs HTTP GET with context for timeout/cancellation
func (c *YahooFinanceClient) doRequestWithContext(ctx context.Context, u string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	return c.client.Do(req)
}

// yahooQuoteResponse represents the Yahoo Finance quote API response
type yahooQuoteResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol               string  `json:"symbol"`
			ShortName            string  `json:"shortName"`
			RegularMarketPrice   float64 `json:"regularMarketPrice"`
			RegularMarketChange  float64 `json:"regularMarketChange"`
			RegularMarketOpen    float64 `json:"regularMarketOpen"`
			RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
			RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
			RegularMarketVolume  int64   `json:"regularMarketVolume"`
		} `json:"result"`
	} `json:"quoteResponse"`
}

// yahooChartResponse represents the Yahoo Finance chart API response
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol                     string  `json:"symbol"`
				ShortName                  string  `json:"shortName"`
				LongName                   string  `json:"longName"`
				RegularMarketPrice         float64 `json:"regularMarketPrice"`
				RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
				RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
				RegularMarketVolume        int64   `json:"regularMarketVolume"`
				RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
				ChartPreviousClose         float64 `json:"chartPreviousClose"`
			} `json:"meta"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
			Timestamp []int64 `json:"timestamp"`
		} `json:"result"`
	} `json:"chart"`
}

// GetQuote fetches the current quote for a symbol
func (c *YahooFinanceClient) GetQuote(symbol string) (*models.Quote, error) {
	return c.GetQuoteWithContext(context.Background(), symbol)
}

// GetQuoteWithContext fetches quote with context support
func (c *YahooFinanceClient) GetQuoteWithContext(ctx context.Context, symbol string) (*models.Quote, error) {
	u := "https://query1.finance.yahoo.com/v7/finance/quote?symbols=" + url.PathEscape(symbol)
	resp, err := c.doRequestWithContext(ctx, u)
	if err != nil {
		return c.getQuoteFromChartWithContext(ctx, symbol)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.getQuoteFromChartWithContext(ctx, symbol)
	}

	var data yahooQuoteResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return c.getQuoteFromChartWithContext(ctx, symbol)
	}

	if len(data.QuoteResponse.Result) == 0 {
		return c.getQuoteFromChartWithContext(ctx, symbol)
	}

	r := data.QuoteResponse.Result[0]
	changePct := 0.0
	if r.RegularMarketPrice > 0 && r.RegularMarketOpen > 0 {
		changePct = ((r.RegularMarketPrice - r.RegularMarketOpen) / r.RegularMarketOpen) * 100
	}

	return &models.Quote{
		Symbol:    r.Symbol,
		Name:      r.ShortName,
		Price:     r.RegularMarketPrice,
		Change:    r.RegularMarketChange,
		ChangePct: changePct,
		Volume:    r.RegularMarketVolume,
		High:      r.RegularMarketDayHigh,
		Low:       r.RegularMarketDayLow,
	}, nil
}

// getQuoteFromChartWithContext fetches a single-quote view from Yahoo chart endpoint.
func (c *YahooFinanceClient) getQuoteFromChartWithContext(ctx context.Context, symbol string) (*models.Quote, error) {
	u := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1d&interval=1d", url.PathEscape(symbol))
	resp, err := c.doRequestWithContext(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("fetch quote: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var data yahooChartResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	result := data.Chart.Result[0]
	meta := result.Meta
	if meta.Symbol == "" {
		meta.Symbol = symbol
	}

	name := meta.ShortName
	if name == "" {
		name = meta.LongName
	}

	price := meta.RegularMarketPrice
	open := meta.RegularMarketPreviousClose
	if open <= 0 {
		open = meta.ChartPreviousClose
	}

	volume := meta.RegularMarketVolume
	high := meta.RegularMarketDayHigh
	low := meta.RegularMarketDayLow

	if len(result.Indicators.Quote) > 0 {
		q := result.Indicators.Quote[0]
		if len(q.Close) > 0 && q.Close[len(q.Close)-1] > 0 {
			price = q.Close[len(q.Close)-1]
		}
		if open <= 0 && len(q.Open) > 0 && q.Open[len(q.Open)-1] > 0 {
			open = q.Open[len(q.Open)-1]
		}
		if high <= 0 && len(q.High) > 0 && q.High[len(q.High)-1] > 0 {
			high = q.High[len(q.High)-1]
		}
		if low <= 0 && len(q.Low) > 0 && q.Low[len(q.Low)-1] > 0 {
			low = q.Low[len(q.Low)-1]
		}
		if volume <= 0 && len(q.Volume) > 0 {
			volume = q.Volume[len(q.Volume)-1]
		}
	}

	change := 0.0
	changePct := 0.0
	if open > 0 {
		change = price - open
		changePct = (change / open) * 100
	}

	return &models.Quote{
		Symbol:    meta.Symbol,
		Name:      name,
		Price:     price,
		Change:    change,
		ChangePct: changePct,
		Volume:    volume,
		High:      high,
		Low:       low,
	}, nil
}

// GetHistory fetches historical price data for a symbol
func (c *YahooFinanceClient) GetHistory(symbol string, range_ string, interval string) ([]models.HistoryPoint, error) {
	return c.GetHistoryWithContext(context.Background(), symbol, range_, interval)
}

// GetHistoryWithContext fetches history with context support
func (c *YahooFinanceClient) GetHistoryWithContext(ctx context.Context, symbol string, range_ string, interval string) ([]models.HistoryPoint, error) {
	u := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s",
		url.PathEscape(symbol), url.QueryEscape(range_), url.QueryEscape(interval))

	resp, err := c.doRequestWithContext(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("fetch history: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var data yahooChartResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no history for symbol: %s", symbol)
	}

	result := data.Chart.Result[0]
	quotes := result.Indicators.Quote
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote data for symbol: %s", symbol)
	}

	closes := quotes[0].Close
	volumes := quotes[0].Volume
	timestamps := result.Timestamp

	points := make([]models.HistoryPoint, 0, len(timestamps))
	for i, ts := range timestamps {
		if i >= len(closes) {
			break
		}
		vol := int64(0)
		if i < len(volumes) {
			vol = volumes[i]
		}
		points = append(points, models.HistoryPoint{
			Date:   time.Unix(ts, 0).Format("2006-01-02"),
			Close:  closes[i],
			Volume: vol,
		})
	}

	return points, nil
}

// GetQuotes fetches quotes for multiple symbols
func (c *YahooFinanceClient) GetQuotes(symbols []string) ([]*models.Quote, error) {
	return c.GetQuotesWithContext(context.Background(), symbols)
}

// GetQuotesWithContext fetches multiple quotes with context support
func (c *YahooFinanceClient) GetQuotesWithContext(ctx context.Context, symbols []string) ([]*models.Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	symbolStr := ""
	for i, s := range symbols {
		if i > 0 {
			symbolStr += ","
		}
		symbolStr += url.PathEscape(s)
	}

	u := "https://query1.finance.yahoo.com/v7/finance/quote?symbols=" + symbolStr
	resp, err := c.doRequestWithContext(ctx, u)
	if err == nil {
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			var data yahooQuoteResponse
			if jsonErr := json.Unmarshal(body, &data); jsonErr == nil && len(data.QuoteResponse.Result) > 0 {
				quotes := make([]*models.Quote, 0, len(data.QuoteResponse.Result))
				for _, r := range data.QuoteResponse.Result {
					changePct := 0.0
					if r.RegularMarketPrice > 0 && r.RegularMarketOpen > 0 {
						changePct = ((r.RegularMarketPrice - r.RegularMarketOpen) / r.RegularMarketOpen) * 100
					}
					quotes = append(quotes, &models.Quote{
						Symbol:    r.Symbol,
						Name:      r.ShortName,
						Price:     r.RegularMarketPrice,
						Change:    r.RegularMarketChange,
						ChangePct: changePct,
						Volume:    r.RegularMarketVolume,
						High:      r.RegularMarketDayHigh,
						Low:       r.RegularMarketDayLow,
					})
				}
				return quotes, nil
			}
		}
	}

	// Fallback when Yahoo quote endpoint is blocked/empty: fetch each symbol via chart endpoint.
	quotes := make([]*models.Quote, 0, len(symbols))
	var firstErr error
	for _, symbol := range symbols {
		q, qErr := c.getQuoteFromChartWithContext(ctx, symbol)
		if qErr != nil {
			if firstErr == nil {
				firstErr = qErr
			}
			continue
		}
		quotes = append(quotes, q)
	}
	if len(quotes) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return quotes, nil
}

// SearchSymbols searches for stock symbols (simplified - Yahoo search endpoint)
func (c *YahooFinanceClient) SearchSymbols(query string, limit int) ([]models.Quote, error) {
	return c.SearchSymbolsWithContext(context.Background(), query, limit)
}

// SearchSymbolsWithContext searches with context support
func (c *YahooFinanceClient) SearchSymbolsWithContext(ctx context.Context, query string, limit int) ([]models.Quote, error) {
	u := "https://query1.finance.yahoo.com/v1/finance/search?q=" + url.QueryEscape(query) + "&quotesCount=" + strconv.Itoa(limit)
	resp, err := c.doRequestWithContext(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var data struct {
		Quotes []struct {
			Symbol    string  `json:"symbol"`
			ShortName string  `json:"shortname"`
			LongName  string  `json:"longname"`
			Price     float64 `json:"regularMarketPrice"`
		} `json:"quotes"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]models.Quote, 0, len(data.Quotes))
	for _, q := range data.Quotes {
		if q.Symbol == "" || q.Symbol == "-" {
			continue
		}
		name := q.ShortName
		if name == "" {
			name = q.LongName
		}
		results = append(results, models.Quote{
			Symbol: q.Symbol,
			Name:   name,
			Price:  q.Price,
		})
	}

	return results, nil
}
