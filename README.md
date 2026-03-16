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

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/pgiraultmatz/stock-checker.git
cd stock-checker
```

2. **Install dependencies**

```bash
go mod download
```

3. **Run locally**

```bash
CGO_ENABLED=0 go run main.go > report.html
open report.html
```

## 🔧 Configuration

### Add/Remove Stocks

Edit `main.go` and modify the `stocks` slice:

```go
stocks := []Stock{
    {Ticker: "AAPL", Name: "Apple Inc", Category: "USA"},
    {Ticker: "TSLA", Name: "Tesla", Category: "USA"},
    {Ticker: "BTC-USD", Name: "Bitcoin", Category: "Cryptos"},
    // Add your stocks here
}
```

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