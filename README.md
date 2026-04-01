# 📊 Stock Market Analyzer

A Go-based stock market analysis tool that calculates RSI (Relative Strength Index) indicators and sends daily HTML reports via email using GitHub Actions.

![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![GitHub Actions](https://img.shields.io/badge/CI-GitHub%20Actions-2088FF?logo=github-actions)

## ✨ Features

- 📈 **Real-time stock data** from Yahoo Finance API
- 📊 **RSI calculation** (14-period weekly)
- 🎯 **Oversold/Overbought detection** (RSI < 30 / RSI > 70)
- 📧 **Automated daily email reports** (HTML format)
- 🏷️ **Category-based organization** (Metals, Cryptos, Energy, USA, Defense, France, Others)
- 🎨 **Clean HTML design** optimized for email clients
- ⚡ **GitHub Actions automation** (runs daily at 8am Paris time)

## 📊 Example Report

### 🥇 Metals

| Stock | Price | Change | RSI | Alert |
|-------|-------|--------|-----|-------|
| Invesco Physical Gold | 479.95 | 📉 -1.44% | RSI: 52.7 | |
| iShares Physical Silver | 73.90 | 📉 -3.27% | RSI: 52.1 | |

### ₿ Cryptos

| Stock | Price | Change | RSI | Alert |
|-------|-------|--------|-----|-------|
| BTC (USD) | 73219.75 | 📈 +0.75% | RSI: 45.9 | |
| ETH (USD) | 2252.19 | 📊 +0.00% | RSI: 44.2 | |

### 🛡️ Defense

| Stock | Price | Change | RSI | Alert |
|-------|-------|--------|-----|-------|
| Lockheed Martin | 646.00 | 📉 -3.84% | RSI: 83.0 | 🟢 OVERBOUGHT |
| Palantir Technologies | 150.95 | 📉 -3.95% | RSI: 36.7 | |

### 🇺🇸 USA

| Stock | Price | Change | RSI | Alert |
|-------|-------|--------|-----|-------|
| Microsoft | 395.55 | 📉 -3.28% | RSI: 23.9 | 🔴 OVERSOLD |
| MicroStrategy Inc | 139.67 | 📈 +4.60% | RSI: 29.6 | 🔴 OVERSOLD |
| NVIDIA | 180.25 | 📈 +1.37% | RSI: 54.5 | |

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- GitHub account (for automated emails)
- Gmail account with App Password
- A stock configuration set up via [stock-portfolio](https://github.com/pgiraultmatz/stock-portfolio) (see below)

### 1. Configure your stock portfolio

Stock configuration (tickers, categories, etc.) is managed through a GitHub Gist via the **[stock-portfolio](https://github.com/pgiraultmatz/stock-portfolio)** web editor. Set it up first — it will give you a `GIST_ID` and require a GitHub Personal Access Token (`GH_TOKEN`) with the `gist` scope.

### 2. Clone and install

```bash
git clone https://github.com/pgiraultmatz/stock-checker.git
cd stock-checker
go mod download
```

### 3. Create a `.env` file

```bash
# .env
GIST_ID=your_gist_id
GH_TOKEN=your_github_pat

GEMINI_API_KEY=your_gemini_key        # or ANTHROPIC_API_KEY
TWITTER_USERNAMES=trader1,trader2     # optional
```

### 4. Run locally

```bash
make run
```

To generate an HTML report file:

```bash
make report
open report.html
```

## 🔧 Configuration

### Add/Remove Stocks

Stock configuration is managed via the [stock-portfolio](https://github.com/pgiraultmatz/stock-portfolio) web editor. Changes are saved to your GitHub Gist and picked up automatically by stock-checker on the next run — no file editing needed.

### Supported Categories

- `Metals` 🥇
- `Cryptos` ₿
- `Energy` ⚡
- `USA` 🇺🇸
- `Defense` 🛡️
- `France` 🇫🇷
- `Others` 🌍

### Ticker Format Examples

| Asset Type | Example | Format |
|------------|---------|--------|
| US Stock | `AAPL` | Direct ticker |
| French Stock | `MC.PA` | Ticker + `.PA` (Paris) |
| German Stock | `5MVW.DE` | Ticker + `.DE` (Frankfurt) |
| London Stock | `SGLD.L` | Ticker + `.L` (London) |
| Crypto | `BTC-USD` | Crypto + `-USD` |

### Twitter/X Context (optional)

Enrich the AI prompt with recent tweets from a quantitative crypto trader.
Two providers are supported: **Nitter** (no credentials needed) and **Twitter API v2**.

The list of Twitter accounts to follow is set via the `TWITTER_USERNAMES` environment variable (comma-separated) — kept out of `config.json` to avoid committing usernames to the repository.

#### Provider: Nitter (recommended)

Nitter is an open-source Twitter mirror that exposes RSS feeds — no API key required.

```json
"twitter": {
  "enabled": true,
  "max_tweets": 5,
  "provider": "nitter",
  "nitter_instance": "https://nitter.poast.org"
}
```

```bash
export TWITTER_USERNAMES=trader1,trader2,trader3
make run
```

> **Note:** Nitter instances can go down or get blocked by Twitter. If `nitter.poast.org` stops working, replace `nitter_instance` with another instance from the [Nitter instance list](https://github.com/zedeus/nitter/wiki/Instances).

#### Provider: Twitter API v2

Requires a Bearer Token from the [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard) (free Basic tier is sufficient).

```json
"twitter": {
  "enabled": true,
  "max_tweets": 5,
  "provider": "api"
}
```

```bash
export TWITTER_USERNAMES=trader1,trader2,trader3
export TWITTER_BEARER_TOKEN=your_bearer_token
make run
```

If fetching fails for any reason, the program continues without Twitter context.

## ⚙️ GitHub Actions Setup

The workflows run automatically on a schedule. Add the following secrets in your repository settings (`Settings → Secrets and variables → Actions`):

| Secret | Description |
|--------|-------------|
| `GIST_ID` | ID of the GitHub Gist containing your `stock-config.json` (managed via [stock-portfolio](https://github.com/pgiraultmatz/stock-portfolio)) |
| `GH_TOKEN` | GitHub Personal Access Token with `gist` scope |
| `GEMINI_API_KEY` | Google Gemini API key (or use `ANTHROPIC_API_KEY`) |
| `TWITTER_USERNAMES` | Comma-separated Twitter handles (optional) |
| `EMAIL_USERNAME` | Gmail address used to send reports |
| `EMAIL_PASSWORD` | Gmail App Password |
| `EMAIL_RECIPIENT` | Email address to receive reports |