# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests if needed
RUN apk --no-cache add ca-certificates

# Copy built binary from builder
COPY --from=builder /app/api .

# Expose port
EXPOSE 8080

# Run application
ENTRYPOINT ["./api"]