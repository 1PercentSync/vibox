# ViBox Dockerfile - Multi-stage build with embedded frontend
# Phase 2.5: Frontend is embedded into the Go binary

# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /build/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend for production
RUN npm run build

# Stage 2: Build Go backend (with embedded frontend)
FROM golang:1.25-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy backend source
COPY . .

# Copy built frontend from frontend-builder to static directory
COPY --from=frontend-builder /build/frontend/dist ./internal/static/dist

# Build the application with embedded frontend
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to reduce binary size
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-w -s" \
    -o /build/vibox \
    ./cmd/server

# Stage 3: Runtime stage
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

# Copy binary from backend-builder (includes embedded frontend)
COPY --from=backend-builder /build/vibox /app/vibox

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
ENTRYPOINT ["/app/vibox"]
