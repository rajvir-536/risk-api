# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o risk-api ./cmd/api

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary and migrations directory
COPY --from=builder /app/risk-api .
COPY --from=builder /app/migrations ./migrations

# Expose server port
EXPOSE 8080

# Environment variables setup
ENV PORT=8080
ENV DB_PATH=/data/risk.db
ENV MIGRATIONS_DIR=/app/migrations

# Run the API server
ENTRYPOINT ["./risk-api"]
