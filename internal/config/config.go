// Package config handles application configuration loading and validation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"stock-checker/internal/models"
)

// Config holds the application configuration.
type Config struct {
	Stocks      []models.Stock    `json:"stocks"`
	Categories  []models.Category `json:"categories"`
	YahooAPI    YahooAPIConfig    `json:"yahoo_api"`
	AI          AIConfig          `json:"ai"`
	Concurrency int               `json:"concurrency"`
}

// AIConfig holds AI/Anthropic API configuration.
type AIConfig struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`     // "api" or "manual_prompt" (default: "manual_prompt")
	Provider  string `json:"provider"` // "gemini" or "anthropic"
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

// YahooAPIConfig holds Yahoo Finance API configuration.
type YahooAPIConfig struct {
	BaseURL   string `json:"base_url"`
	Range     string `json:"range"`
	Interval  string `json:"interval"`
	UserAgent string `json:"user_agent"`
	Timeout   int    `json:"timeout_seconds"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		YahooAPI: YahooAPIConfig{
			BaseURL:   "https://query1.finance.yahoo.com/v8/finance/chart",
			Range:     "1y",
			Interval:  "1wk",
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			Timeout:   30,
		},
		AI: AIConfig{
			Enabled:   true,
			Mode:      "manual_prompt",
			Provider:  "gemini",
			Model:     "gemini-2.0-flash",
			MaxTokens: 2000,
		},
		Concurrency: 10,
		Categories: []models.Category{
			{Name: "Metals", Emoji: "gold", Order: 1},
			{Name: "Cryptos", Emoji: "bitcoin", Order: 2},
			{Name: "Energy", Emoji: "zap", Order: 3},
			{Name: "USA", Emoji: "us", Order: 4},
			{Name: "Defense", Emoji: "shield", Order: 5},
			{Name: "France", Emoji: "fr", Order: 6},
			{Name: "Others", Emoji: "globe", Order: 7},
		},
	}
}

// Load reads configuration from a JSON file.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Stocks) == 0 {
		return fmt.Errorf("no stocks configured")
	}

	for i, stock := range c.Stocks {
		if stock.Ticker == "" {
			return fmt.Errorf("stock %d: ticker is required", i)
		}
		if stock.Name == "" {
			return fmt.Errorf("stock %d (%s): name is required", i, stock.Ticker)
		}
	}

	if c.Concurrency < 1 {
		c.Concurrency = 10
	}

	return nil
}

// GetCategoryOrder returns a map of category name to order for sorting.
func (c *Config) GetCategoryOrder() map[string]int {
	order := make(map[string]int)
	for _, cat := range c.Categories {
		order[cat.Name] = cat.Order
	}
	return order
}

// GetCategoryEmoji returns a map of category name to emoji.
func (c *Config) GetCategoryEmoji() map[string]string {
	emojis := make(map[string]string)
	for _, cat := range c.Categories {
		emojis[cat.Name] = cat.Emoji
	}
	return emojis
}

// FindConfigFile searches for a config file in common locations.
func FindConfigFile() (string, error) {
	locations := []string{
		"config.json",
		"config/config.json",
		filepath.Join(os.Getenv("HOME"), ".stock-checker", "config.json"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("config file not found in: %v", locations)
}
