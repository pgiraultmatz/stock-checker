# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /stock-checker \
    ./cmd/stock-checker

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Copy binary and config
COPY --from=builder /stock-checker /app/stock-checker
COPY config.json /app/config.json
COPY internal/report/templates /app/internal/report/templates

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Default command
ENTRYPOINT ["/app/stock-checker"]
CMD ["-config", "/app/config.json"]
