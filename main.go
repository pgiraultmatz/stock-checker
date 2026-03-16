package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

type YahooResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol string `json:"symbol"`
			} `json:"meta"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

type Stock struct {
	Ticker   string
	Name     string
	Category string
}

type StockResult struct {
	Stock         Stock
	CurrentPrice  float64
	ChangePercent float64
	RSI           float64
}

func calculateRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	gains := make([]float64, 0)
	losses := make([]float64, 0)

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, math.Abs(change))
		}
	}

	if len(gains) < period {
		return 0
	}

	avgGain := average(gains[len(gains)-period:])
	avgLoss := average(losses[len(losses)-period:])

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func getStockAnalysis(stock Stock) *StockResult {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1y&interval=1wk", stock.Ticker)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error %s: %v\n", stock.Ticker, err)
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %s: %v\n", stock.Ticker, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Error %s: %s\n", stock.Ticker, resp.Status)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %s: %v\n", stock.Ticker, err)
		return nil
	}

	var data YahooResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Error %s: %v\n", stock.Ticker, err)
		return nil
	}

	if len(data.Chart.Result) == 0 {
		log.Printf("No data for %s\n", stock.Ticker)
		return nil
	}

	result := data.Chart.Result[0]

	var closes []float64
	if len(result.Indicators.Quote) > 0 {
		closes = result.Indicators.Quote[0].Close
	}

	if len(closes) < 15 {
		log.Printf("Not enough data for %s (only %d weeks)\n", stock.Ticker, len(closes))
		return nil
	}

	currentPrice := closes[len(closes)-1]

	var previousClose float64
	var changePercent float64

	for i := len(closes) - 2; i >= 0; i-- {
		if closes[i] != currentPrice && closes[i] > 0 {
			previousClose = closes[i]
			changePercent = ((currentPrice - previousClose) / previousClose) * 100

			if math.Abs(changePercent) > 0.01 {
				break
			}
		}
	}

	rsi := calculateRSI(closes, 14)

	return &StockResult{
		Stock:         stock,
		CurrentPrice:  currentPrice,
		ChangePercent: changePercent,
		RSI:           rsi,
	}
}

func generateHTMLReport(results []*StockResult, categoryOrder map[string]int) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stock Market Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f6f8fa;
            color: #24292f;
            padding: 20px;
            margin: 0;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            background-color: #ffffff;
            border: 1px solid #d0d7de;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #238636 0%, #1f6feb 100%);
            color: #ffffff;
            padding: 32px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 32px;
            color: #ffffff;
        }
        .header p {
            margin: 12px 0 0 0;
            opacity: 0.95;
            color: #ffffff;
            font-size: 16px;
        }
        .category {
            border-top: 2px solid #d0d7de;
            padding: 24px 32px;
            background-color: #ffffff;
        }
        .category-title {
            font-size: 22px;
            font-weight: 700;
            margin-bottom: 16px;
            color: #24292f;
        }
        .stock-table {
            width: 100%;
            border-collapse: collapse;
            table-layout: fixed;
        }
        .stock-row {
            border-bottom: 1px solid #d0d7de;
        }
        .stock-row:last-child {
            border-bottom: none;
        }
        .stock-row td {
            padding: 14px 10px;
            color: #24292f;
        }
        .stock-name {
            font-weight: 600;
            width: 35%;
            text-align: left;
            color: #24292f;
        }
        .stock-price {
            text-align: right;
            font-family: 'Courier New', monospace;
            font-weight: 700;
            width: 15%;
            color: #24292f;
        }
        .stock-change {
            text-align: right;
            font-family: 'Courier New', monospace;
            font-weight: 700;
            width: 18%;
            white-space: nowrap;
        }
        .stock-rsi {
            text-align: right;
            font-family: 'Courier New', monospace;
            width: 17%;
            color: #656d76;
        }
        .stock-status {
            text-align: center;
            font-size: 12px;
            font-weight: 700;
            width: 15%;
        }
        .stock-row td.positive { color: #1a7f37; }
        .stock-row td.negative { color: #cf222e; }
        .stock-row td.neutral { color: #656d76; }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 13px;
            color: #656d76;
            border-top: 1px solid #d0d7de;
            background-color: #f6f8fa;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Stock Market Report</h1>
            <p>` + getCurrentDate() + `</p>
        </div>
`)

	sort.Slice(results, func(i, j int) bool {
		catI := categoryOrder[results[i].Stock.Category]
		catJ := categoryOrder[results[j].Stock.Category]
		if catI != catJ {
			return catI < catJ
		}
		return results[i].Stock.Name < results[j].Stock.Name
	})

	currentCategory := ""
	for _, result := range results {
		if result.Stock.Category != currentCategory {
			if currentCategory != "" {
				html.WriteString("        </table>\n")
				html.WriteString("    </div>\n")
			}

			currentCategory = result.Stock.Category
			emoji := getCategoryEmoji(currentCategory)

			html.WriteString(fmt.Sprintf(`    <div class="category">
        <div class="category-title">%s %s</div>
        <table class="stock-table">
`, emoji, currentCategory))
		}

		changeClass := "neutral"
		changeIcon := "📊"
		if result.ChangePercent > 0.01 {
			changeClass = "positive"
			changeIcon = "📈"
		} else if result.ChangePercent < -0.01 {
			changeClass = "negative"
			changeIcon = "📉"
		}

		rsiStatus := ""
		if result.RSI < 30 {
			rsiStatus = `<span style="color: #cf222e; font-weight: 700; font-size: 11px;">🔴 OVERSOLD</span>`
		} else if result.RSI > 70 {
			rsiStatus = `<span style="color: #1a7f37; font-weight: 700; font-size: 11px;">🟢 OVERBOUGHT</span>`
		}

		html.WriteString(fmt.Sprintf(`            <tr class="stock-row">
                <td class="stock-name">%s</td>
                <td class="stock-price">%.2f</td>
                <td class="stock-change %s">%s %+.2f%%</td>
                <td class="stock-rsi">RSI: %.1f</td>
                <td class="stock-status">%s</td>
            </tr>
`, result.Stock.Name, result.CurrentPrice, changeClass, changeIcon, result.ChangePercent, result.RSI, rsiStatus))
	}

	if currentCategory != "" {
		html.WriteString("        </table>\n")
		html.WriteString("    </div>\n")
	}

	html.WriteString(`    <div class="footer">
            Automatically generated by Stock Analyzer
        </div>
    </div>
</body>
</html>`)

	return html.String()
}

func getCurrentDate() string {
	return time.Now().Format("Monday, January 2, 2006 at 3:04 PM")
}

func getCategoryEmoji(category string) string {
	switch category {
	case "Metals":
		return "🥇"
	case "Cryptos":
		return "₿"
	case "Energy":
		return "⚡"
	case "USA":
		return "🇺🇸"
	case "Defense":
		return "🛡️"
	case "France":
		return "🇫🇷"
	case "Others":
		return "🌍"
	default:
		return "📊"
	}
}

func main() {
	stocks := []Stock{
		{Ticker: "BTC-USD", Name: "BTC (USD)", Category: "Cryptos"},
		{Ticker: "ETH-USD", Name: "ETH (USD)", Category: "Cryptos"},
		{Ticker: "MC.PA", Name: "LVMH", Category: "France"},
		{Ticker: "OR.PA", Name: "L'Oréal", Category: "France"},
		{Ticker: "AI.PA", Name: "Air Liquide", Category: "France"},
		{Ticker: "SAN.PA", Name: "Sanofi", Category: "France"},
		{Ticker: "BNP.PA", Name: "BNP Paribas", Category: "France"},
		{Ticker: "SU.PA", Name: "Schneider Electric S.E.", Category: "France"},
		{Ticker: "RI.PA", Name: "Pernod Ricard SA", Category: "France"},
		{Ticker: "TTE.PA", Name: "TotalEnergies", Category: "Energy"},
		{Ticker: "VST", Name: "Vistra Corp.", Category: "Energy"},
		{Ticker: "5MVW.DE", Name: "iShares MSCI World Energy", Category: "Energy"},
		{Ticker: "NUCL.MI", Name: "VanEck Uranium Nuclear", Category: "Energy"},
		{Ticker: "ISLN.L", Name: "iShares Physical Silver", Category: "Metals"},
		{Ticker: "SGLD.L", Name: "Invesco Physical Gold", Category: "Metals"},
		{Ticker: "VOO", Name: "Vanguard S&P 500 ETF", Category: "USA"},
		{Ticker: "NVDA", Name: "NVIDIA", Category: "USA"},
		{Ticker: "GOOGL", Name: "Alphabet (Google)", Category: "USA"},
		{Ticker: "AMZN", Name: "Amazon", Category: "USA"},
		{Ticker: "TSLA", Name: "Tesla", Category: "USA"},
		{Ticker: "AAPL", Name: "Apple Inc", Category: "USA"},
		{Ticker: "UNH", Name: "UnitedHealth Group", Category: "USA"},
		{Ticker: "MSTR", Name: "MicroStrategy Inc", Category: "USA"},
		{Ticker: "CMG", Name: "Chipotle Mexican Grill", Category: "USA"},
		{Ticker: "MSFT", Name: "Microsoft", Category: "USA"},
		{Ticker: "CROX", Name: "Crocs Inc.", Category: "USA"},
		{Ticker: "DFNS.MI", Name: "VanEck Defense UCITS ETF", Category: "Defense"},
		{Ticker: "LMT", Name: "Lockheed Martin", Category: "Defense"},
		{Ticker: "KTOS", Name: "Kratos Defense", Category: "Defense"},
		{Ticker: "PLTR", Name: "Palantir Technologies", Category: "Defense"},
		{Ticker: "EXA.PA", Name: "Exail Technologies SA", Category: "Defense"},
		{Ticker: "CHDVD.SW", Name: "iShares Swiss Dividend", Category: "Others"},
		{Ticker: "MCHI", Name: "iShares MSCI China", Category: "Others"},
	}

	results := make([]*StockResult, 0)
	for _, stock := range stocks {
		result := getStockAnalysis(stock)
		if result != nil {
			results = append(results, result)
		}
	}

	categoryOrder := map[string]int{
		"Metals":  1,
		"Cryptos": 2,
		"Energy":  3,
		"USA":     4,
		"Defense": 5,
		"France":  6,
		"Others":  7,
	}

	htmlReport := generateHTMLReport(results, categoryOrder)
	fmt.Println(htmlReport)
}
