# =============================================================================
# Nanayam - Distribution Server (Go Gateway) Dockerfile
# =============================================================================

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files first so dependency downloads stay cached across source edits.
# go.sum must be copied too: without it the build cannot verify module checksums.
COPY go.mod go.sum ./
RUN go mod download

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
