// Package yahoo provides a client for the Yahoo Finance API.
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"stock-checker/internal/config"
)

// Client is a Yahoo Finance API client.
type Client struct {
	httpClient *http.Client
	config     config.YahooAPIConfig
}

// NewClient creates a new Yahoo Finance API client.
func NewClient(cfg config.YahooAPIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		config: cfg,
	}
}

// ChartResponse represents the Yahoo Finance chart API response.
type ChartResponse struct {
	Chart struct {
		Result []ChartResult `json:"result"`
		Error  *ChartError   `json:"error"`
	} `json:"chart"`
}

// ChartResult contains the data for a single stock.
type ChartResult struct {
	Meta       ChartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators ChartIndicators `json:"indicators"`
}

// ChartMeta contains metadata about the stock.
type ChartMeta struct {
	Symbol             string  `json:"symbol"`
	Currency           string  `json:"currency"`
	ExchangeName       string  `json:"exchangeName"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

// ChartIndicators contains the quote data.
type ChartIndicators struct {
	Quote []QuoteData `json:"quote"`
}

// QuoteData contains OHLCV data.
type QuoteData struct {
	Open   []float64 `json:"open"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}

// ChartError represents an error from the Yahoo Finance API.
type ChartError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// StockData contains the processed data for a stock.
type StockData struct {
	Symbol     string
	Currency   string
	Exchange   string
	Closes     []float64
	Timestamps []time.Time
}

// GetChartData fetches chart data for a ticker symbol.
func (c *Client) GetChartData(ctx context.Context, ticker string) (*StockData, error) {
	url := fmt.Sprintf("%s/%s?range=%s&interval=%s",
		c.config.BaseURL,
		ticker,
		c.config.Range,
		c.config.Interval,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var chartResp ChartResponse
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if chartResp.Chart.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s",
			chartResp.Chart.Error.Code,
			chartResp.Chart.Error.Description,
		)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for %s", ticker)
	}

	result := chartResp.Chart.Result[0]

	var closes []float64
	if len(result.Indicators.Quote) > 0 {
		closes = result.Indicators.Quote[0].Close
	}

	var timestamps []time.Time
	for _, ts := range result.Timestamp {
		timestamps = append(timestamps, time.Unix(ts, 0))
	}

	return &StockData{
		Symbol:     result.Meta.Symbol,
		Currency:   result.Meta.Currency,
		Exchange:   result.Meta.ExchangeName,
		Closes:     closes,
		Timestamps: timestamps,
	}, nil
}
