#!/usr/bin/env bash
# =============================================================================
# Anti-Corruption Complaint System - Start Applications
# =============================================================================
# Builds and starts one Go gateway + Next.js org-console PER organization.
# All four orgs (ACB, Dept, Oversight, Judiciary) get their own stack.
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
    echo -e "${BLUE}  Start Complaint System Apps (4-Org)${NC}"
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

    log_info "Building and starting complaint system services (4 gateways + 4 consoles)..."
    docker-compose -f "${COMPOSE_FILE}" up -d --build

    log_ok "Complaint system apps are running"
    echo ""
    echo "Organization Consoles:"
    echo "  ACB Console:        http://localhost:3000  (Gateway REST: http://localhost:8080)"
    echo "  Dept Console:       http://localhost:3001  (Gateway REST: http://localhost:8081)"
    echo "  Oversight Console:  http://localhost:3002  (Gateway REST: http://localhost:8082)"
    echo "  Judiciary Console:  http://localhost:3003  (Gateway REST: http://localhost:8083)"
    echo ""
    echo "gRPC Gateways:"
    echo "  ACB:        grpc://localhost:50051"
    echo "  Dept:       grpc://localhost:50052"
    echo "  Oversight:  grpc://localhost:50053"
    echo "  Judiciary:  grpc://localhost:50054"
    echo ""
    echo "Test the REST API:"
    echo "  curl http://localhost:8080/v1/health"
    echo "  curl http://localhost:8080/v1/ListComplaints"
    echo ""
}

main "$@"
