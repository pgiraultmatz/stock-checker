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

const systemPrompt = `
Tu es un analyste économique et financier spécialisé dans les technologies critiques, les infrastructures stratégiques et les marchés actions.

Ta mission est de produire un rapport macroéconomique et financier des dernières 24 à 48 heures, avec un focus prioritaire sur les domaines suivants :

# Intelligence artificielle
# Semi-conducteurs
# Chips / puces IA
# Énergie liée aux data centers
# Photonique
# Défense / technologies dual-use

Objectif du rapport :
1. Identifier les actualités macroéconomiques, financières, industrielles et géopolitiques importantes.
2. Expliquer leur impact potentiel sur les marchés.
3. Identifier des opportunités potentielles d’investissement en actions.
4. Repérer des entreprises liées à ces thèmes, ou hors de ces thèmes, qui pourraient présenter un prix intéressant — notamment en cas de décote par rapport aux fondamentaux ou de catalyseur proche (résultats, annonce produit, décision réglementaire, etc.).
5. Surveiller les investissements, prises de participation ou partenariats stratégiques réalisés par les grands acteurs technologiques (NVIDIA, Google, Microsoft, Amazon, Meta, Apple, etc.) dans des entreprises plus petites — ces mouvements peuvent signaler des opportunités ou valider une thèse sur une entreprise cible.
6. Relever les déclarations publiques de CEO influents (Jensen Huang, Satya Nadella, Sundar Pichai, Sam Altman, Elon Musk, etc.) qui mentionnent ou impliquent explicitement d’autres entreprises, secteurs ou technologies — ces prises de position peuvent avoir un impact significatif sur les valorisations.
`

// XGroupSection holds the fetched content for one named Twitter/X group.
type XGroupSection struct {
	Name    string
	Content string
}

// PromptContext holds optional context sections to include in the prompt.
type PromptContext struct {
	VIXLine string
	XGroups []XGroupSection
}

// FormatXGroups concatenates all group contents into a single string for API mode.
func FormatXGroups(groups []XGroupSection) string {
	var sb strings.Builder
	for _, g := range groups {
		if g.Content == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", g.Name))
		sb.WriteString(g.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

// BuildPrompt returns the formatted prompt to copy-paste into Claude or any other model.
func BuildPrompt(results []*models.StockResult, promptPath string, ctx PromptContext) (string, error) {
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
	if ctx.VIXLine != "" {
		sb.WriteString("\n**Indicateur de volatilité:**\n")
		sb.WriteString(ctx.VIXLine)
		sb.WriteString("\n")
	}
	sb.WriteString(stockData)

	for i, group := range ctx.XGroups {
		if group.Content == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n\n════════════════════════════════════════\n"))
		sb.WriteString(fmt.Sprintf("SECTION %d — %s\n", i+2, strings.ToUpper(group.Name)))
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString(group.Content)
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

1. **Top 3 Stocks to Watch**: Les 3 actions les plus intéressantes à surveiller (basé sur momentum, potentiel)
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
  "market_summary": "Résumé en 1-2 phrases de la situation globale du portefeuille. Mentionne également les dates importantes des prochaines semaines pour ce portefeuille (publications de résultats, dividendes, décisions de banques centrales, indicateurs macro) en précisant les tickers concernés."
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
