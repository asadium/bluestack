# Multi-stage build for Bluestack
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X github.com/asad/bluestack/internal/cli.Version=${VERSION:-dev}" \
    -o bluestack \
    ./cmd/bluestack

# Stage 2: Minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS support (if needed in future)
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 bluestack && \
    adduser -D -u 1000 -G bluestack bluestack

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bluestack .

# Create data directory
RUN mkdir -p /app/data && \
    chown -R bluestack:bluestack /app

# Switch to non-root user
USER bluestack

# Expose edge port (default 4566)
EXPOSE 4566

# Set environment variables
ENV EDGE_PORT=4566
ENV DATA_DIR=/app/data
ENV ENABLED_SERVICES=blob
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4566/health || exit 1

# Entrypoint
ENTRYPOINT ["./bluestack", "start"]

