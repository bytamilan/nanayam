#!/usr/bin/env bash
# =============================================================================
# Anti-Corruption Complaint System - Start Applications
# =============================================================================
# Builds and starts the Go gateway + Next.js operator console for the
# complaint system.
#
# Usage:
#   ./scripts/start-complaint-apps.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="${PWD}/docker/complaint-apps.yaml"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Start Complaint System Apps${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    if ! docker ps --format '{{.Names}}' | grep -q "peer0.acb.nanayam.com"; then
        log_err "Complaint network does not appear to be running. Run ./scripts/start-complaint.sh first."
        exit 1
    fi

    if ! docker network ls --format '{{.Name}}' | grep -q "^nanayam$"; then
        log_err "Docker network 'nanayam' not found. Run ./scripts/start-complaint.sh first."
        exit 1
    fi

    log_info "Building and starting complaint system services..."
    docker-compose -f "${COMPOSE_FILE}" up -d --build

    log_ok "Complaint system apps are running"
    echo ""
    echo "Services:"
    echo "  Go Gateway (gRPC):   grpc://localhost:50051"
    echo "  Go Gateway (REST):   http://localhost:8080"
    echo "  Operator Console:    http://localhost:3000"
    echo ""
    echo "Test the REST API:"
    echo "  curl http://localhost:8080/v1/health"
    echo "  curl http://localhost:8080/v1/ListComplaints"
    echo ""
}

main "$@"
