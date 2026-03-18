package main

import (
	"stock-checker/internal/models"
)

// mockStockResults returns fake stock results for report generation testing.
func mockStockResults() []*models.StockResult {
	return []*models.StockResult{
		{
			Stock:         models.Stock{Ticker: "AAPL", Name: "Apple Inc.", Category: "Tech"},
			CurrentPrice:  189.45,
			ChangePercent: 1.23,
			RSI:           58.4,
		},
		{
			Stock:         models.Stock{Ticker: "MSFT", Name: "Microsoft Corp.", Category: "Tech"},
			CurrentPrice:  415.20,
			ChangePercent: -0.87,
			RSI:           72.1,
		},
		{
			Stock:         models.Stock{Ticker: "NVDA", Name: "NVIDIA Corp.", Category: "Tech"},
			CurrentPrice:  875.30,
			ChangePercent: 3.45,
			RSI:           78.9,
		},
		{
			Stock:         models.Stock{Ticker: "GOOGL", Name: "Alphabet Inc.", Category: "Tech"},
			CurrentPrice:  165.80,
			ChangePercent: 0.52,
			RSI:           55.2,
		},
		{
			Stock:         models.Stock{Ticker: "JPM", Name: "JPMorgan Chase", Category: "Finance"},
			CurrentPrice:  198.60,
			ChangePercent: -1.34,
			RSI:           42.7,
		},
		{
			Stock:         models.Stock{Ticker: "GS", Name: "Goldman Sachs", Category: "Finance"},
			CurrentPrice:  452.10,
			ChangePercent: -2.10,
			RSI:           28.3,
		},
		{
			Stock:         models.Stock{Ticker: "XOM", Name: "ExxonMobil Corp.", Category: "Energy"},
			CurrentPrice:  112.45,
			ChangePercent: 0.78,
			RSI:           50.1,
		},
		{
			Stock:         models.Stock{Ticker: "BTC-USD", Name: "Bitcoin", Category: "Crypto"},
			CurrentPrice:  68420.00,
			ChangePercent: 4.20,
			RSI:           65.3,
		},
	}
}
