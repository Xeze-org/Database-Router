# ---- Build Stage ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Download dependencies first — leverages Docker layer cache.
COPY go.mod go.sum ./
RUN go mod download

# Copy all source — .dockerignore excludes config.json, .env, vendor, etc.
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o db-router ./cmd/

# ---- Run Stage ──────────────────────────────────────────────────────────────
FROM alpine:3.19

LABEL org.opencontainers.image.title="db-router" \
      org.opencontainers.image.description="Unified gRPC router for PostgreSQL, MongoDB and Redis" \
      org.opencontainers.image.source="https://github.com/xeze/db-router" \
      org.opencontainers.image.licenses="MIT"

WORKDIR /app

# ca-certificates — needed for TLS to external databases
# grpcurl        — used by the healthcheck probe
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/db-router .

# config.json is NOT embedded — mount it at runtime:
#   -v ./config.json:/app/config.json:ro
# The binary falls back to a default config.json in the work dir if not mounted.
VOLUME ["/app/config.json"]

EXPOSE 50051

CMD ["./db-router"]
