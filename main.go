package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "math"
    "net/http"
    "sort"
)

// Codes couleurs ANSI
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorGray   = "\033[90m"
    ColorCyan   = "\033[36m"
    ColorBold   = "\033[1m"
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

func getRSIStatus(rsi float64) string {
    if rsi < 30 {
        return ColorRed + "🔴 OVERSOLD" + ColorReset
    } else if rsi > 70 {
        return ColorGreen + "🟢 OVERBOUGHT" + ColorReset  // Changé en vert
    }
    return ""
}

func getStockAnalysis(stock Stock) *StockResult {
    url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1y&interval=1wk", stock.Ticker)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Printf("Erreur %s: %v\n", stock.Ticker, err)
        return nil
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
    req.Header.Set("Accept", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Erreur %s: %v\n", stock.Ticker, err)
        return nil
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        log.Printf("Erreur %s: %s\n", stock.Ticker, resp.Status)
        return nil
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Erreur %s: %v\n", stock.Ticker, err)
        return nil
    }

    var data YahooResponse
    err = json.Unmarshal(body, &data)
    if err != nil {
        log.Printf("Erreur %s: %v\n", stock.Ticker, err)
        return nil
    }

    if len(data.Chart.Result) == 0 {
        log.Printf("Pas de données pour %s\n", stock.Ticker)
        return nil
    }

    result := data.Chart.Result[0]

    var closes []float64
    if len(result.Indicators.Quote) > 0 {
        closes = result.Indicators.Quote[0].Close
    }

    if len(closes) < 3 {
        log.Printf("Pas assez de données pour %s\n", stock.Ticker)
        return nil
    }

    currentPrice := closes[len(closes)-1]

    var previousClose float64
    for i := len(closes) - 2; i >= 0; i-- {
        if closes[i] != currentPrice {
            previousClose = closes[i]
            break
        }
    }

    if previousClose == 0 {
        previousClose = closes[len(closes)-2]
    }

    changePercent := 0.0
    if previousClose != 0 {
        changePercent = ((currentPrice - previousClose) / previousClose) * 100
    }

    rsi := calculateRSI(closes, 14)

    return &StockResult{
        Stock:         stock,
        CurrentPrice:  currentPrice,
        ChangePercent: changePercent,
        RSI:           rsi,
    }
}

func displayStockResult(result *StockResult) {
    if result == nil {
        return
    }

    rsiStatus := getRSIStatus(result.RSI)

    // Couleur et icône selon la variation
    var variationColor, variationIcon string
    if result.ChangePercent > 0.01 {
        variationColor = ColorGreen
        variationIcon = "📈"
    } else if result.ChangePercent < -0.01 {
        variationColor = ColorRed
        variationIcon = "📉"
    } else {
        variationColor = ColorGray
        variationIcon = "📊"
    }

    fmt.Printf("%-30s | %10.2f | %s %s%+6.2f%%%s | RSI: %5.1f | %s\n",
        result.Stock.Name,
        result.CurrentPrice,
        variationIcon,
        variationColor,
        result.ChangePercent,
        ColorReset,
        result.RSI,
        rsiStatus)
}

func main() {
    stocks := []Stock{
        // Cryptos
        {Ticker: "BTC-USD", Name: "BTC (USD)", Category: "Cryptos"},
        {Ticker: "ETH-USD", Name: "ETH (USD)", Category: "Cryptos"},

        // France
        {Ticker: "MC.PA", Name: "LVMH", Category: "France"},
        {Ticker: "OR.PA", Name: "L'Oréal", Category: "France"},
        {Ticker: "AI.PA", Name: "Air Liquide", Category: "France"},
        {Ticker: "SAN.PA", Name: "Sanofi", Category: "France"},
        {Ticker: "BNP.PA", Name: "BNP Paribas", Category: "France"},
        {Ticker: "SU.PA", Name: "Schneider Electric S.E.", Category: "France"},
        {Ticker: "RI.PA", Name: "Pernod Ricard SA", Category: "France"},

        // Energie
        {Ticker: "TTE.PA", Name: "TotalEnergies", Category: "Energie"},
        {Ticker: "VST", Name: "Vistra Corp.", Category: "Energie"},
        {Ticker: "5MVW.DE", Name: "iShares MSCI World Energy", Category: "Energie"},
        {Ticker: "NUCL.MI", Name: "VanEck Uranium Nuclear", Category: "Energie"},

        // Métaux
        {Ticker: "ISLN.L", Name: "iShares Physical Silver", Category: "Métaux"},
        {Ticker: "SGLD.L", Name: "Invesco Physical Gold", Category: "Métaux"},

        // USA
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

        // Defense
        {Ticker: "DFNS.MI", Name: "VanEck Defense UCITS ETF", Category: "Defense"},
        {Ticker: "LMT", Name: "Lockheed Martin", Category: "Defense"},
        {Ticker: "KTOS", Name: "Kratos Defense", Category: "Defense"},
        {Ticker: "PLTR", Name: "Palantir Technologies", Category: "Defense"},
        {Ticker: "EXA.PA", Name: "Exail Technologies SA", Category: "Defense"},

        // Autres
        {Ticker: "CHDVD.SW", Name: "iShares Swiss Dividend", Category: "Autres"},
        {Ticker: "ICHNZ.XC", Name: "iShares MSCI China", Category: "Autres"},
    }

    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println("📊 Analyse Boursière par Catégorie")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

    // Récupérer toutes les données
    results := make([]*StockResult, 0)
    for _, stock := range stocks {
        result := getStockAnalysis(stock)
        if result != nil {
            results = append(results, result)
        }
    }

    // Trier par catégorie dans l'ordre souhaité
    categoryOrder := map[string]int{
        "Métaux":  1,
        "Cryptos": 2,
        "Energie": 3,
        "USA":     4,
        "Defense": 5,
        "France":  6,
        "Autres":  7,
    }

    sort.Slice(results, func(i, j int) bool {
        catI := categoryOrder[results[i].Stock.Category]
        catJ := categoryOrder[results[j].Stock.Category]
        if catI != catJ {
            return catI < catJ
        }
        // Si même catégorie, trier par nom
        return results[i].Stock.Name < results[j].Stock.Name
    })

    // Afficher par catégorie
    currentCategory := ""
    for _, result := range results {
        if result.Stock.Category != currentCategory {
            currentCategory = result.Stock.Category

            // Emoji par catégorie
            var emoji string
            switch currentCategory {
            case "Métaux":
                emoji = "🥇"
            case "Energie":
                emoji = "⚡"
            case "USA":
                emoji = "🇺🇸"
            case "Defense":
                emoji = "🛡️"
            case "France":
                emoji = "🇫🇷"
            case "Autres":
                emoji = "🌍"
            default:
                emoji = "📊"
            }

            fmt.Printf("\n%s%s %s %s\n", ColorBold, emoji, currentCategory, ColorReset)
            fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────")
        }
        displayStockResult(result)
    }

    fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}