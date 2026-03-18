// Stock Checker - A production-grade stock market analysis tool.
//
// This application fetches real-time stock data from Yahoo Finance,
// calculates RSI indicators, and generates HTML reports.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"stock-checker/internal/ai"
	"stock-checker/internal/config"
	"stock-checker/internal/models"
	"stock-checker/internal/report"
	"stock-checker/internal/yahoo"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	promptPath := flag.String("prompt", "manual_prompt.txt", "Path to manual prompt template file")
	outputPath := flag.String("output", "", "Path to output HTML file (defaults to stdout)")
	check := flag.Bool("check", false, "Check a single stock (use -ticker to specify, random otherwise)")
	ticker := flag.String("ticker", "", "Ticker symbol to check (implies -check)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	timeout := flag.Duration("timeout", 5*time.Minute, "Timeout for the entire operation")
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Single stock check mode (triggered by -check or -ticker)
	if *check || *ticker != "" {
		if err := runSingleCheck(ctx, cfg, *ticker, logger); err != nil {
			logger.Error("check failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// Full report mode
	if err := runFullReport(ctx, cfg, *outputPath, *promptPath, logger); err != nil {
		logger.Error("analysis failed", "error", err)
		os.Exit(1)
	}
}

func runSingleCheck(ctx context.Context, cfg *config.Config, ticker string, logger *slog.Logger) error {
	// Find the stock to check
	var stock models.Stock

	if ticker == "" {
		// Pick a random stock
		idx := rand.Intn(len(cfg.Stocks))
		stock = cfg.Stocks[idx]
	} else {
		// Find by ticker
		found := false
		for _, s := range cfg.Stocks {
			if s.Ticker == ticker {
				stock = s
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ticker %q not found in config", ticker)
		}
	}

	// Analyze the stock
	analyzer := yahoo.NewAnalyzer(cfg, logger)
	result := analyzer.AnalyzeStock(ctx, stock)

	if result.Error != nil {
		return fmt.Errorf("failed to analyze %s: %w", stock.Ticker, result.Error)
	}

	// Output to console
	printStockResult(result)
	return nil
}

func printStockResult(r *models.StockResult) {
	// Determine trend indicator
	var trend string
	if r.IsPositive() {
		trend = "\033[32m▲\033[0m" // Green up arrow
	} else if r.IsNegative() {
		trend = "\033[31m▼\033[0m" // Red down arrow
	} else {
		trend = "\033[33m●\033[0m" // Yellow dot
	}

	// Determine RSI status
	var rsiStatus string
	if r.IsOversold() {
		rsiStatus = " \033[31m[OVERSOLD]\033[0m"
	} else if r.IsOverbought() {
		rsiStatus = " \033[32m[OVERBOUGHT]\033[0m"
	}

	// Format change with color
	var changeStr string
	if r.ChangePercent >= 0 {
		changeStr = fmt.Sprintf("\033[32m+%.2f%%\033[0m", r.ChangePercent)
	} else {
		changeStr = fmt.Sprintf("\033[31m%.2f%%\033[0m", r.ChangePercent)
	}

	fmt.Println()
	fmt.Printf("  %s %s (%s)\n", trend, r.Stock.Name, r.Stock.Ticker)
	fmt.Printf("  %-12s %s\n", "Category:", r.Stock.Category)
	fmt.Printf("  %-12s %.2f\n", "Price:", r.CurrentPrice)
	fmt.Printf("  %-12s %s\n", "Change:", changeStr)
	fmt.Printf("  %-12s %.1f%s\n", "RSI:", r.RSI, rsiStatus)
	fmt.Println()
}

func runFullReport(ctx context.Context, cfg *config.Config, outputPath string, promptPath string, logger *slog.Logger) error {
	logger.Info("starting stock analysis",
		"stocks", len(cfg.Stocks),
		"concurrency", cfg.Concurrency,
	)

	// Create analyzer and fetch stock data
	analyzer := yahoo.NewAnalyzer(cfg, logger)

	startTime := time.Now()
	results := analyzer.AnalyzeAll(ctx, cfg.Stocks)
	elapsed := time.Since(startTime)

	logger.Info("analysis complete",
		"successful", len(results),
		"failed", len(cfg.Stocks)-len(results),
		"duration", elapsed.Round(time.Millisecond),
	)

	if len(results) == 0 {
		return fmt.Errorf("no stocks were successfully analyzed")
	}

	// Run AI analysis if enabled
	var aiAnalysis *ai.Analysis
	var manualPrompt string
	if cfg.AI.Enabled {
		if cfg.AI.Mode == "manual_prompt" {
			var err error
			manualPrompt, err = ai.BuildPrompt(results, promptPath)
			if err != nil {
				logger.Warn("failed to build manual prompt, continuing without it", "error", err)
			} else {
				logger.Info("manual prompt mode: prompt generated for copy-paste")
			}
		} else {
			apiKey, provider := getAICredentials(cfg.AI.Provider)
			if apiKey != "" {
				logger.Info("running AI analysis", "provider", provider)

				aiClient := ai.NewClient(ai.ClientConfig{
					Provider: provider,
					APIKey:   apiKey,
					Model:    cfg.AI.Model,
				})
				aiAnalyzer := ai.NewAnalyzer(aiClient)

				var err error
				aiAnalysis, err = aiAnalyzer.Analyze(ctx, results)
				if err != nil {
					logger.Warn("AI analysis failed, continuing without it", "error", err)
				} else {
					logger.Info("AI analysis complete")
				}
			} else {
				logger.Warn("AI analysis enabled but no API key found", "provider", cfg.AI.Provider)
			}
		}
	}

	// Generate HTML report
	generator, err := report.NewGenerator(cfg.GetCategoryEmoji(), cfg.GetCategoryOrder())
	if err != nil {
		return fmt.Errorf("creating report generator: %w", err)
	}

	htmlReport, err := generator.GenerateWithAI(results, aiAnalysis, manualPrompt)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Output report
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(htmlReport), 0644); err != nil {
			return fmt.Errorf("writing report to file: %w", err)
		}
		logger.Info("report written", "path", outputPath)
	} else {
		fmt.Println(htmlReport)
	}

	return nil
}

// getAICredentials returns the API key and provider based on config and environment.
// It checks environment variables in order: configured provider first, then fallbacks.
func getAICredentials(configuredProvider string) (string, ai.Provider) {
	// Map of providers to their environment variable names
	providerEnvVars := map[ai.Provider]string{
		ai.ProviderGemini:    "GEMINI_API_KEY",
		ai.ProviderAnthropic: "ANTHROPIC_API_KEY",
	}

	// Try configured provider first
	provider := ai.Provider(configuredProvider)
	if envVar, ok := providerEnvVars[provider]; ok {
		if apiKey := os.Getenv(envVar); apiKey != "" {
			return apiKey, provider
		}
	}

	// Fallback: try all providers in order
	fallbackOrder := []ai.Provider{ai.ProviderGemini, ai.ProviderAnthropic}
	for _, p := range fallbackOrder {
		if apiKey := os.Getenv(providerEnvVars[p]); apiKey != "" {
			return apiKey, p
		}
	}

	return "", ai.Provider(configuredProvider)
}
