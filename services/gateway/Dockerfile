# =============================================================================
# Nanayam - Distribution Server (Go Gateway) Dockerfile
# =============================================================================

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
RUN go mod tidy && go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway .

# =============================================================================
# Runtime
# =============================================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/gateway /app/gateway

# Expose gRPC and REST ports
EXPOSE 50051 8080

ENTRYPOINT ["/app/gateway"]
