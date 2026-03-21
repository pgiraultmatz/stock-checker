package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"stock-checker/internal/models"
)

// Analysis represents the AI-generated market analysis.
type Analysis struct {
	TopStocks       []TopStock       `json:"top_stocks"`
	NewsContext     []NewsItem       `json:"news_context"`
	Recommendations []Recommendation `json:"recommendations"`
	MarketSummary   string           `json:"market_summary"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

// TopStock represents a stock highlighted by AI analysis.
type TopStock struct {
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	Reasoning string `json:"reasoning"`
	Signal    string `json:"signal"` // "bullish", "bearish", "neutral"
}

// NewsItem represents a relevant market news item.
type NewsItem struct {
	Headline    string   `json:"headline"`
	Impact      string   `json:"impact"` // "positive", "negative", "neutral"
	AffectedBy  []string `json:"affected_by,omitempty"`
	Description string   `json:"description"`
}

// Recommendation represents an actionable recommendation.
type Recommendation struct {
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
	Action string `json:"action"` // "buy", "sell", "hold", "watch"
	Reason string `json:"reason"`
	Risk   string `json:"risk"` // "low", "medium", "high"
}

// Analyzer performs AI-powered stock analysis.
type Analyzer struct {
	client Client
}

// NewAnalyzer creates a new AI analyzer.
func NewAnalyzer(client Client) *Analyzer {
	return &Analyzer{client: client}
}

const systemPrompt = `Tu es un analyste financier expert spécialisé dans l'analyse technique et fondamentale des marchés.
Tu fournis des analyses claires, concises et actionnables basées sur les données RSI et les variations de prix.

Règles importantes:
- Sois direct et factuel, évite le jargon inutile
- Base tes recommandations sur les données fournies (RSI, variations)
- Mentionne les risques associés à chaque recommandation
- RSI < 30 = survendu (potentiel achat), RSI > 70 = suracheté (potentiel vente)
- Utilise tes connaissances des événements récents du marché (dernières 24-48h)

Tu dois TOUJOURS répondre en JSON valide selon le format demandé.`

// BuildPrompt returns the formatted prompt to copy-paste into Claude or any other model.
// It reads the template from promptPath and appends the stock data.
// If twitterContext is non-empty, it is added as a separate clearly delimited section.
func BuildPrompt(results []*models.StockResult, promptPath string, twitterContext string) (string, error) {
	template, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("reading prompt template %q: %w", promptPath, err)
	}
	var a Analyzer
	stockData := a.prepareStockData(results)

	var sb strings.Builder
	sb.WriteString("════════════════════════════════════════\n")
	sb.WriteString("SECTION 1 — DONNÉES DE MARCHÉ\n")
	sb.WriteString("════════════════════════════════════════\n\n")
	sb.WriteString(string(template))
	sb.WriteString(stockData)

	if twitterContext != "" {
		sb.WriteString("\n\n════════════════════════════════════════\n")
		sb.WriteString("SECTION 2 — SENTIMENT DES TRADERS\n")
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString(twitterContext)
	}

	return sb.String(), nil
}

// Analyze performs AI analysis on stock results.
// twitterContext is optional: if non-empty, it is included in the prompt as additional context.
func (a *Analyzer) Analyze(ctx context.Context, results []*models.StockResult, twitterContext string) (*Analysis, error) {
	// Prepare stock data for the prompt
	stockData := a.prepareStockData(results)

	twitterSection := ""
	if twitterContext != "" {
		twitterSection = "\n\nContexte additionnel — analyses récentes d'un trader quantitatif crypto:\n" + twitterContext
	}

	userPrompt := fmt.Sprintf(`Analyse ces données de marché et fournis:

1. **Top 3 Stocks to Watch**: Les 3 actions les plus intéressantes à surveiller (basé sur RSI, momentum, potentiel)
2. **Recent News Context**: 2-3 événements de marché récents (dernières 24-48h) qui pourraient impacter ces positions
3. **Actionable Recommendations**: Recommandations concrètes (buy/sell/hold/watch) pour les positions les plus significatives

Données actuelles du portefeuille:
%s

Réponds UNIQUEMENT en JSON valide avec cette structure exacte:
{
  "top_stocks": [
    {"ticker": "XXX", "name": "Nom", "reasoning": "Explication courte", "signal": "bullish|bearish|neutral"}
  ],
  "news_context": [
    {"headline": "Titre court", "impact": "positive|negative|neutral", "affected_by": ["TICKER1", "TICKER2"], "description": "Description de l'impact"}
  ],
  "recommendations": [
    {"ticker": "XXX", "name": "Nom", "action": "buy|sell|hold|watch", "reason": "Raison courte", "risk": "low|medium|high"}
  ],
  "market_summary": "Résumé en 1-2 phrases de la situation globale du portefeuille"
}`, stockData+twitterSection)

	response, err := a.client.Complete(ctx, systemPrompt, userPrompt, 2000)
	if err != nil {
		return nil, fmt.Errorf("AI completion failed: %w", err)
	}

	// Parse JSON response
	analysis, err := a.parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("parsing AI response: %w", err)
	}

	analysis.GeneratedAt = time.Now()
	return analysis, nil
}

// prepareStockData formats stock results for the AI prompt.
func (a *Analyzer) prepareStockData(results []*models.StockResult) string {
	var sb strings.Builder

	// Group by category
	categories := make(map[string][]*models.StockResult)
	for _, r := range results {
		categories[r.Stock.Category] = append(categories[r.Stock.Category], r)
	}

	// Sort categories
	catNames := make([]string, 0, len(categories))
	for name := range categories {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	for _, catName := range catNames {
		sb.WriteString(fmt.Sprintf("\n## %s\n", catName))
		for _, r := range categories[catName] {
			status := ""
			if r.IsOversold() {
				status = " [OVERSOLD]"
			} else if r.IsOverbought() {
				status = " [OVERBOUGHT]"
			}

			changeSign := ""
			if r.ChangePercent > 0 {
				changeSign = "+"
			}

			sb.WriteString(fmt.Sprintf("- %s (%s): %.2f | %s%.2f%% | RSI: %.1f%s\n",
				r.Stock.Name, r.Stock.Ticker, r.CurrentPrice,
				changeSign, r.ChangePercent, r.RSI, status))
		}
	}

	return sb.String()
}

// parseResponse extracts the Analysis from Claude's response.
func (a *Analyzer) parseResponse(response string) (*Analysis, error) {
	// Try to extract JSON from the response
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w\nResponse was: %s", err, response[:min(500, len(response))])
	}

	return &analysis, nil
}
