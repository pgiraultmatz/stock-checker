// Package report generates HTML reports from stock analysis results.
package report

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"stock-checker/internal/ai"
	"stock-checker/internal/models"
)

//go:embed templates/*.html
var templateFS embed.FS

// Generator creates HTML reports from stock analysis results.
type Generator struct {
	templates      *template.Template
	categoryEmojis map[string]string
	categoryOrder  map[string]int
}

// VIXData holds the VIX index data for display at the top of the report.
type VIXData struct {
	Price         string
	ChangePercent float64
	Change        string
	ChangeClass   string
	Level         string // "low", "moderate", "high", "extreme"
	LevelClass    string
}

// TemplateData contains the data passed to the HTML template.
type TemplateData struct {
	Title           string
	GeneratedAt     string
	CategoryGroups  []CategoryGroupData
	TotalStocks     int
	OversoldCount   int
	OverboughtCount int
	AIAnalysis      *AIAnalysisData
	ManualPrompt    string
	VIX             *VIXData
}

// CategoryGroupData represents a category with its stocks for the template.
type CategoryGroupData struct {
	Name   string
	Emoji  string
	Stocks []StockRowData
}

// StockRowData represents a single stock row for the template.
type StockRowData struct {
	Name          string
	Price         string
	Currency      string
	Change        string
	ChangePercent float64
	ChangeClass   string
	ChangeIcon    string
	RSI           string
	RSIValue      float64
	RSIStatus     string
	RSIStatusHTML template.HTML
}

// AIAnalysisData contains AI analysis data for the template.
type AIAnalysisData struct {
	TopStocks       []TopStockData
	NewsContext     []NewsItemData
	Recommendations []RecommendationData
	MarketSummary   string
}

// TopStockData represents an AI-highlighted stock.
type TopStockData struct {
	Ticker      string
	Name        string
	Reasoning   string
	Signal      string
	SignalClass string
	SignalIcon  string
}

// NewsItemData represents a market news item.
type NewsItemData struct {
	Headline    string
	Impact      string
	ImpactClass string
	AffectedBy  string
	Description string
}

// RecommendationData represents an actionable recommendation.
type RecommendationData struct {
	Ticker      string
	Name        string
	Action      string
	ActionClass string
	ActionIcon  string
	Reason      string
	Risk        string
	RiskClass   string
}

// NewGenerator creates a new report generator.
func NewGenerator(categoryEmojis map[string]string, categoryOrder map[string]int) (*Generator, error) {
	funcMap := template.FuncMap{
		"formatPrice": func(price float64) string {
			return fmt.Sprintf("%.2f", price)
		},
		"formatChange": func(change float64) string {
			return fmt.Sprintf("%+.2f%%", change)
		},
		"formatRSI": func(rsi float64) string {
			return fmt.Sprintf("%.1f", rsi)
		},
		"getChangeClass": func(change float64) string {
			if change > 0.01 {
				return "positive"
			} else if change < -0.01 {
				return "negative"
			}
			return "neutral"
		},
		"getChangeIcon": func(change float64) string {
			if change > 0.01 {
				return "↑"
			} else if change < -0.01 {
				return "↓"
			}
			return "→"
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	return &Generator{
		templates:      tmpl,
		categoryEmojis: categoryEmojis,
		categoryOrder:  categoryOrder,
	}, nil
}

// NewVIXData builds a VIXData from raw price and change values.
func NewVIXData(price, changePercent float64) *VIXData {
	changeClass := "neutral"
	if changePercent > 0.01 {
		changeClass = "negative" // rising VIX = bad
	} else if changePercent < -0.01 {
		changeClass = "positive" // falling VIX = good
	}

	level, levelClass := "Normal", "vix-moderate"
	switch {
	case price >= 30:
		level, levelClass = "Stress", "vix-extreme"
	case price >= 15:
		level, levelClass = "Normal", "vix-moderate"
	default:
		level, levelClass = "Calm", "vix-low"
	}

	return &VIXData{
		Price:         fmt.Sprintf("%.2f", price),
		ChangePercent: changePercent,
		Change:        fmt.Sprintf("%+.2f%%", changePercent),
		ChangeClass:   changeClass,
		Level:         level,
		LevelClass:    levelClass,
	}
}

// Generate creates an HTML report from the analysis results.
func (g *Generator) Generate(results []*models.StockResult) (string, error) {
	return g.GenerateWithAI(results, nil, "", nil)
}

// GenerateWithAI creates an HTML report with optional AI analysis or manual prompt.
func (g *Generator) GenerateWithAI(results []*models.StockResult, aiAnalysis *ai.Analysis, manualPrompt string, vix *VIXData) (string, error) {
	data := g.prepareTemplateData(results)

	// Add AI analysis if provided
	if aiAnalysis != nil {
		data.AIAnalysis = g.convertAIAnalysis(aiAnalysis)
	}

	// Add manual prompt if provided
	if manualPrompt != "" {
		data.ManualPrompt = manualPrompt
	}

	data.VIX = vix

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "report.html", data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// prepareTemplateData transforms analysis results into template-ready data.
func (g *Generator) prepareTemplateData(results []*models.StockResult) TemplateData {
	// Sort results by category order, then by name
	sort.Slice(results, func(i, j int) bool {
		orderI := g.categoryOrder[results[i].Stock.Category]
		orderJ := g.categoryOrder[results[j].Stock.Category]
		if orderI != orderJ {
			return orderI < orderJ
		}
		return results[i].Stock.Name < results[j].Stock.Name
	})

	// Group by category
	categoryMap := make(map[string][]StockRowData)
	categoryOrderList := make([]string, 0)
	oversoldCount := 0
	overboughtCount := 0

	for _, result := range results {
		if _, exists := categoryMap[result.Stock.Category]; !exists {
			categoryOrderList = append(categoryOrderList, result.Stock.Category)
		}

		row := g.createStockRow(result)
		categoryMap[result.Stock.Category] = append(categoryMap[result.Stock.Category], row)

		if result.IsOversold() {
			oversoldCount++
		}
		if result.IsOverbought() {
			overboughtCount++
		}
	}

	// Sort category list by order
	sort.Slice(categoryOrderList, func(i, j int) bool {
		return g.categoryOrder[categoryOrderList[i]] < g.categoryOrder[categoryOrderList[j]]
	})

	// Build category groups
	groups := make([]CategoryGroupData, 0, len(categoryOrderList))
	for _, catName := range categoryOrderList {
		groups = append(groups, CategoryGroupData{
			Name:   catName,
			Emoji:  g.getCategoryEmoji(catName),
			Stocks: categoryMap[catName],
		})
	}

	return TemplateData{
		Title:           "Stock Market Report",
		GeneratedAt:     time.Now().Format("Monday, January 2, 2006"),
		CategoryGroups:  groups,
		TotalStocks:     len(results),
		OversoldCount:   oversoldCount,
		OverboughtCount: overboughtCount,
	}
}

// createStockRow creates a template-ready stock row.
func (g *Generator) createStockRow(result *models.StockResult) StockRowData {
	changeClass := "neutral"
	changeIcon := "→"
	if result.ChangePercent > 0.01 {
		changeClass = "positive"
		changeIcon = "↑"
	} else if result.ChangePercent < -0.01 {
		changeClass = "negative"
		changeIcon = "↓"
	}

	var rsiStatus string
	var rsiStatusHTML template.HTML
	if result.IsOversold() {
		rsiStatus = "OVERSOLD"
		rsiStatusHTML = template.HTML(`<span class="rsi-oversold">OVERSOLD</span>`)
	} else if result.IsOverbought() {
		rsiStatus = "OVERBOUGHT"
		rsiStatusHTML = template.HTML(`<span class="rsi-overbought">OVERBOUGHT</span>`)
	}

	return StockRowData{
		Name:          result.Stock.Name,
		Price:         fmt.Sprintf("%.2f", result.CurrentPrice),
		Currency:      result.Currency,
		Change:        fmt.Sprintf("%+.2f%%", result.ChangePercent),
		ChangePercent: result.ChangePercent,
		ChangeClass:   changeClass,
		ChangeIcon:    changeIcon,
		RSI:           fmt.Sprintf("%.1f", result.RSI),
		RSIValue:      result.RSI,
		RSIStatus:     rsiStatus,
		RSIStatusHTML: rsiStatusHTML,
	}
}

// getCategoryEmoji returns the emoji for a category.
func (g *Generator) getCategoryEmoji(category string) string {
	emojiMap := map[string]string{
		"1st_place_medal": "🥇",
		"coin":            "🪙",
		"zap":             "⚡",
		"us":              "🇺🇸",
		"shield":          "🛡️",
		"fr":              "🇫🇷",
		"earth_americas":  "🌍",
		"gold":            "🥇",
		"bitcoin":         "₿",
		"globe":           "🌍",
	}

	if emojiName, ok := g.categoryEmojis[category]; ok {
		if emoji, exists := emojiMap[emojiName]; exists {
			return emoji
		}
		return emojiName
	}

	return "📊"
}

// convertAIAnalysis converts AI analysis to template-ready data.
func (g *Generator) convertAIAnalysis(analysis *ai.Analysis) *AIAnalysisData {
	data := &AIAnalysisData{
		MarketSummary: analysis.MarketSummary,
	}

	// Convert top stocks
	for _, ts := range analysis.TopStocks {
		signalClass := "neutral"
		signalIcon := "→"
		switch ts.Signal {
		case "bullish":
			signalClass = "positive"
			signalIcon = "↑"
		case "bearish":
			signalClass = "negative"
			signalIcon = "↓"
		}
		data.TopStocks = append(data.TopStocks, TopStockData{
			Ticker:      ts.Ticker,
			Name:        ts.Name,
			Reasoning:   ts.Reasoning,
			Signal:      ts.Signal,
			SignalClass: signalClass,
			SignalIcon:  signalIcon,
		})
	}

	// Convert news context
	for _, news := range analysis.NewsContext {
		impactClass := "neutral"
		switch news.Impact {
		case "positive":
			impactClass = "positive"
		case "negative":
			impactClass = "negative"
		}
		affectedBy := ""
		if len(news.AffectedBy) > 0 {
			affectedBy = strings.Join(news.AffectedBy, ", ")
		}
		data.NewsContext = append(data.NewsContext, NewsItemData{
			Headline:    news.Headline,
			Impact:      news.Impact,
			ImpactClass: impactClass,
			AffectedBy:  affectedBy,
			Description: news.Description,
		})
	}

	// Convert recommendations
	for _, rec := range analysis.Recommendations {
		actionClass := "neutral"
		actionIcon := "●"
		switch rec.Action {
		case "buy":
			actionClass = "positive"
			actionIcon = "↑"
		case "sell":
			actionClass = "negative"
			actionIcon = "↓"
		case "watch":
			actionClass = "watch"
			actionIcon = "👁"
		}
		riskClass := "medium"
		switch rec.Risk {
		case "low":
			riskClass = "low"
		case "high":
			riskClass = "high"
		}
		data.Recommendations = append(data.Recommendations, RecommendationData{
			Ticker:      rec.Ticker,
			Name:        rec.Name,
			Action:      rec.Action,
			ActionClass: actionClass,
			ActionIcon:  actionIcon,
			Reason:      rec.Reason,
			Risk:        rec.Risk,
			RiskClass:   riskClass,
		})
	}

	return data
}
