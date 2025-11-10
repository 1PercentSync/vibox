# ViBox Dockerfile - Multi-stage build
# Stage 1: Build stage

FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/server \
    ./cmd/server

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash \
    curl

# Create non-root user for security
RUN addgroup -g 1000 vibox && \
    adduser -D -u 1000 -G vibox vibox

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server /app/server

# Change ownership to non-root user
RUN chown -R vibox:vibox /app

# Switch to non-root user
USER vibox

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Run the application
ENTRYPOINT ["/app/server"]
