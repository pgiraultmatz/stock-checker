.PHONY: build run run-no-twitter test clean lint fmt help report report-mock

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Binary name
BINARY_NAME=stock-checker
BINARY_PATH=./$(BINARY_NAME)

# Paths
CMD_PATH=./cmd/stock-checker
CONFIG_PATH=config.json

# Build flags
LDFLAGS=-ldflags "-s -w"

## help: Display this help message
help:
	@echo "Stock Checker - Production Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## run: Generate HTML report and open it
run: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -output report.html -verbose
	@echo "Report saved to report.html"
	@open report.html

## run-no-twitter: Generate HTML report without fetching tweets
run-no-twitter: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -output report.html -no-twitter -verbose
	@echo "Report saved to report.html"
	@open report.html

## check: Check a single stock (random if TICKER not set)
check: build
ifdef TICKER
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -ticker $(TICKER)
else
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -check
endif

## report: Generate HTML report to file (no browser)
report: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -output report.html -verbose
	@echo "Report saved to report.html"

## prompt: Fetch stock data + Twitter context, build full AI prompt and write to prompt.txt (copies to clipboard on macOS)
prompt: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -output /dev/null -prompt-output prompt.txt -verbose
	@echo "Prompt saved to prompt.txt"
	@command -v pbcopy >/dev/null 2>&1 && cat prompt.txt | pbcopy && echo "Copied to clipboard" || true

## prompt-twitter: Fetch tweets only and build a standalone trader analysis prompt (copies to clipboard on macOS)
prompt-twitter: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -twitter-only -prompt-output prompt-twitter.txt -verbose
	@echo "Twitter prompt saved to prompt-twitter.txt"
	@command -v pbcopy >/dev/null 2>&1 && cat prompt-twitter.txt | pbcopy && echo "Copied to clipboard" || true

## check-alerts: Check intraday price alerts (writes alerts.html only if triggered)
check-alerts: build
	@./$(BINARY_NAME) -config $(CONFIG_PATH) -check-alerts -alerts-output alerts.html -verbose

## report-mock: Generate HTML report using mock data (no API calls)
report-mock: build
	@./$(BINARY_NAME) -mock -output report.html -verbose
	@echo "Mock report saved to report.html"

## test: Run all tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linting checks
lint:
	@echo "Running linter..."
	@$(GOVET) ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_PATH)
	@rm -f report.html
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOCMD) mod download
	@$(GOCMD) mod tidy
	@echo "Dependencies ready"

## verify: Run all checks (fmt, lint, test)
verify: fmt lint test
	@echo "All checks passed!"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GOCMD) install $(CMD_PATH)
	@echo "Installed to $(shell $(GOCMD) env GOPATH)/bin/$(BINARY_NAME)"
