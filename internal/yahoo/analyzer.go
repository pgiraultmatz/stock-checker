// Package yahoo provides a client for the Yahoo Finance API.
package yahoo

import (
	"context"
	"log/slog"
	"math"
	"sync"

	"stock-checker/internal/analysis"
	"stock-checker/internal/config"
	"stock-checker/internal/models"
)

// Analyzer fetches and analyzes stock data.
type Analyzer struct {
	client        *Client
	rsiCalculator *analysis.RSICalculator
	concurrency   int
	logger        *slog.Logger
}

// NewAnalyzer creates a new stock analyzer.
func NewAnalyzer(cfg *config.Config, logger *slog.Logger) *Analyzer {
	if logger == nil {
		logger = slog.Default()
	}

	return &Analyzer{
		client:        NewClient(cfg.YahooAPI),
		rsiCalculator: analysis.NewRSICalculator(14),
		concurrency:   cfg.Concurrency,
		logger:        logger,
	}
}

// AnalyzeStock analyzes a single stock.
func (a *Analyzer) AnalyzeStock(ctx context.Context, stock models.Stock) *models.StockResult {
	result := &models.StockResult{Stock: stock}

	data, err := a.client.GetChartData(ctx, stock.Ticker)
	if err != nil {
		a.logger.Warn("failed to fetch stock data",
			"ticker", stock.Ticker,
			"error", err,
		)
		result.Error = err
		return result
	}

	if len(data.Closes) < 15 {
		a.logger.Warn("insufficient data for analysis",
			"ticker", stock.Ticker,
			"weeks", len(data.Closes),
		)
		result.Error = err
		return result
	}

	result.CurrentPrice = data.CurrentPrice
	result.Currency = data.Currency

	j1, j2, err := a.client.GetPreviousDayClose(ctx, stock.Ticker)
	if err != nil {
		a.logger.Warn("failed to fetch daily closes", "ticker", stock.Ticker, "error", err)
	} else if j1 > 0 {
		change := ((data.CurrentPrice - j1) / j1) * 100
		if math.Abs(change) < 0.05 && j2 > 0 {
			change = ((j1 - j2) / j2) * 100
		}
		result.ChangePercent = change
	}

	result.RSI = a.rsiCalculator.Calculate(data.Closes)

	return result
}

// AnalyzeAll analyzes multiple stocks concurrently.
func (a *Analyzer) AnalyzeAll(ctx context.Context, stocks []models.Stock) []*models.StockResult {
	results := make([]*models.StockResult, len(stocks))

	// Use a semaphore to limit concurrency
	sem := make(chan struct{}, a.concurrency)
	var wg sync.WaitGroup

	for i, stock := range stocks {
		wg.Add(1)
		go func(idx int, s models.Stock) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Check for context cancellation
			select {
			case <-ctx.Done():
				results[idx] = &models.StockResult{
					Stock: s,
					Error: ctx.Err(),
				}
				return
			default:
			}

			results[idx] = a.AnalyzeStock(ctx, s)
		}(i, stock)
	}

	wg.Wait()

	// Filter out nil results and failed analyses
	validResults := make([]*models.StockResult, 0, len(results))
	for _, r := range results {
		if r != nil && r.Error == nil {
			validResults = append(validResults, r)
		}
	}

	return validResults
}
