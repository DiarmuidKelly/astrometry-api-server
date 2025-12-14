# Build stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy the local Astrometry-Go-Client dependency first
COPY Astrometry-Go-Client /Astrometry-Go-Client

# Copy go mod files
COPY Astrometry-API-Server/go.mod ./
COPY Astrometry-API-Server/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY Astrometry-API-Server/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/server \
    ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    docker-cli \
    && rm -rf /var/cache/apk/*

# Create non-root user and docker group
RUN addgroup -g 984 docker && \
    addgroup -g 1000 astrometry && \
    adduser -D -u 1000 -G astrometry astrometry && \
    adduser astrometry docker

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server /app/server

# Change ownership
RUN chown -R astrometry:astrometry /app

# Create entrypoint script to fix volume permissions
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'mkdir -p /shared-data' >> /entrypoint.sh && \
    echo 'chown -R astrometry:astrometry /shared-data' >> /entrypoint.sh && \
    echo 'exec su-exec astrometry /app/server' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Install su-exec for dropping privileges
RUN apk add --no-cache su-exec

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the entrypoint script as root (it will drop to astrometry user)
ENTRYPOINT ["/entrypoint.sh"]
