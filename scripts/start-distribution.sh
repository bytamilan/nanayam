#!/usr/bin/env bash
# =============================================================================
# Nanayam - Start Distribution Server
# =============================================================================
# Builds and runs the Go gRPC/REST gateway service (distribution server) and
# the Next.js operator console. Both services connect to the running Fabric
# network via the shared 'nanayam' Docker network.
#
# Usage:
#   ./scripts/start-distribution.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="${PWD}/docker/apps.yaml"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam - Start Distribution Server${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    # Ensure the Fabric network is running
    if ! docker ps --format '{{.Names}}' | grep -q "peer0.org1.nanayam.com"; then
        log_err "Fabric network does not appear to be running. Run ./scripts/start-fabric.sh first."
        exit 1
    fi

    # Ensure the shared network exists
    if ! docker network ls --format '{{.Name}}' | grep -q "^nanayam$"; then
        log_err "Docker network 'nanayam' not found. Run ./scripts/start-fabric.sh first."
        exit 1
    fi

    log_info "Building and starting distribution services..."
    docker-compose -f "${COMPOSE_FILE}" up -d --build

    log_ok "Distribution server is running"
    echo ""
    echo "Services:"
    echo "  Go Gateway (gRPC):   grpc://localhost:50051"
    echo "  Go Gateway (REST):   http://localhost:8080"
    echo "  Operator Console:    http://localhost:3000"
    echo ""
    echo "Test the REST API:"
    echo "  curl http://localhost:8080/v1/ListAssets"
    echo ""
}

main "$@"
