# Multi-stage Docker build for echo-mcp server
# Stage 1: Builder - Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build static binary with CGO disabled
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags '-extldflags "-static" -s -w' \
    -o echo-mcp \
    ./cmd/echo-mcp

# Stage 2: Tester - Run tests
FROM builder AS tester

# Run tests with coverage
RUN go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
RUN go tool cover -func=coverage.out

# Stage 3: Runtime - Minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S echouser && adduser -S echouser -G echouser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/echo-mcp .

# Change ownership to non-root user
RUN chown -R echouser:echouser /app

# Switch to non-root user
USER echouser

# Expose HTTP port (for remote mode)
EXPOSE 8080

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD echo '{"jsonrpc":"2.0","id":1,"method":"ping"}' | ./echo-mcp --mode=local || exit 1

# Default command (can be overridden)
CMD ["./echo-mcp", "--mode=remote", "--port=8080"]
