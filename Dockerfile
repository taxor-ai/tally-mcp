# Multi-stage build for Tally MCP Server
# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tally-mcp .

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/tally-mcp .

# Copy templates directory
COPY --from=builder /app/tally ./tally

# Set environment variables with defaults
ENV TALLY_HOST=${TALLY_HOST:-localhost}
ENV TALLY_PORT=${TALLY_PORT:-9900}
ENV TALLY_COMPANY=${TALLY_COMPANY:-}
ENV TALLY_LOG_LEVEL=${TALLY_LOG_LEVEL:-info}
ENV TALLY_LOG_FILE=${TALLY_LOG_FILE:-}

# MCP servers communicate via stdin/stdout
# The entrypoint is the binary which reads from stdin and writes to stdout
ENTRYPOINT ["/app/tally-mcp"]
