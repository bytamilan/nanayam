#!/usr/bin/env bash
# =============================================================================
# Nanayam - Stop Fabric Network
# =============================================================================
# Stops the Fabric containers. Use --clean to also remove volumes and
# generated crypto/channel artifacts.
#
# Usage:
#   ./scripts/stop-fabric.sh
#   ./scripts/stop-fabric.sh --clean
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="${PWD}/docker/fabric-network.yaml"
CLEAN=false

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }

if [ "${1:-}" = "--clean" ]; then
    CLEAN=true
fi

log_info "Stopping Fabric network containers..."
docker-compose -f "${COMPOSE_FILE}" down

if [ "${CLEAN}" = true ]; then
    log_warn "Removing Docker volumes..."
    docker-compose -f "${COMPOSE_FILE}" down -v

    log_warn "Removing generated artifacts..."
    rm -rf "${PWD:?}/crypto-config"
    rm -rf "${PWD:?}/channel-artifacts"
    rm -rf "${PWD:?}/bin"

    log_ok "Cleanup complete. Run ./scripts/setup-fabric.sh to regenerate everything."
else
    log_ok "Fabric network stopped. Use --clean to remove volumes and artifacts."
fi
