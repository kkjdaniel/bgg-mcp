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

# Required MCP Registry label
LABEL io.modelcontextprotocol.server.name="io.github.kkjdaniel/bgg-mcp"

# Default command - runs in stdio mode
ENTRYPOINT ["/usr/local/bin/bgg-mcp"]
