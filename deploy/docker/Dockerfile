FROM golang:alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git gcc musl-dev

# Copy dependencies list
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API binary (GOMAXPROCS=1 to prevent OOM on small instances)
RUN GOMAXPROCS=1 CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api cmd/api/main.go

# Build Consumer binary
RUN GOMAXPROCS=1 CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o consumer cmd/consumer/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy compiled binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/consumer .
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8080

# Default command (can be overridden in docker-compose / k8s to run consumer)
CMD ["./api"]
