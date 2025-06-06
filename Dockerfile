# Generated by https://smithery.ai. See: https://smithery.ai/docs/build/project-config
# syntax=docker/dockerfile:1

# Builder stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy source
COPY . .
# Build binary
RUN go build -o bgg-mcp main.go

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
# Copy binary
COPY --from=builder /app/bgg-mcp /usr/local/bin/bgg-mcp
# Set executable permissions
RUN chmod +x /usr/local/bin/bgg-mcp
# Default command
ENTRYPOINT ["/usr/local/bin/bgg-mcp"]
