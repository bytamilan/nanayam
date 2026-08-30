#!/usr/bin/env bash
# =============================================================================
# Nanayam - Start Server (one-shot, servers only)
# =============================================================================
# Brings up everything server-side — the Fabric network (orderer, peers,
# CAs), the sample chaincode, and the Go gateway — with a single command.
# Deliberately does not start any client (the Next.js operator console, a
# Flutter app): those are separate concerns you opt into yourself.
#
# Safe to re-run: each stage is skipped if it's already done, so this is
# also the command to bring things back up after a reboot.
#
# Usage:
#   ./scripts/start-server.sh          # bring the server stack up
#   ./scripts/start-server.sh --down   # stop it (keeps generated crypto/data)
#   ./scripts/start-server.sh --clean  # stop it and wipe crypto/channel data
#
# Scope: the "basic" profile (asset-transfer-basic chaincode) only. The
# complaint-workflow profile has its own existing scripts
# (setup-complaint.sh / start-complaint.sh / deploy-complaint.sh /
# start-complaint-apps.sh) and isn't wired into this one-shot entry point.
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

# Resolve the repo root from this script's own location so it works
# regardless of the caller's current directory (unlike the older scripts
# here, which assume they're run from the repo root).
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

APPS_COMPOSE="docker/apps.yaml"

# ---------------------------------------------------------------------------
# docker compose v2 ("docker compose") vs. the legacy standalone v1 binary
# ("docker-compose") — prefer v2, since that's what current Docker Desktop /
# Docker Engine installs ship, but fall back for hosts that still only have
# the old binary (as the rest of this repo's scripts assume).
# ---------------------------------------------------------------------------
compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "$@"
    else
        docker-compose "$@"
    fi
}

container_running() {
    docker ps --format '{{.Names}}' | grep -qx "$1"
}

check_prereqs() {
    command -v docker >/dev/null 2>&1 || {
        log_err "Docker is required. Install it from https://docs.docker.com/get-docker/"
        exit 1
    }
    if ! docker compose version >/dev/null 2>&1 && ! command -v docker-compose >/dev/null 2>&1; then
        log_err "Docker Compose is required (either the 'docker compose' plugin or standalone 'docker-compose')."
        exit 1
    fi
    docker info >/dev/null 2>&1 || {
        log_err "Docker doesn't seem to be running. Start Docker Desktop (or the Docker daemon) and try again."
        exit 1
    }
}

wait_for_gateway() {
    log_info "Waiting for the gateway to become healthy..."
    for _ in $(seq 1 30); do
        if curl -fsS "http://localhost:8080/health" >/dev/null 2>&1; then
            log_ok "Gateway is healthy"
            return 0
        fi
        sleep 2
    done
    log_warn "Gateway didn't report healthy within 60s — check 'docker logs nanayam-gateway'."
}

down() {
    local clean="${1:-false}"
    log_info "Stopping the gateway..."
    compose -f "${APPS_COMPOSE}" stop gateway 2>/dev/null || true

    if [ "${clean}" = true ]; then
        ./scripts/stop-fabric.sh --clean
    else
        ./scripts/stop-fabric.sh
    fi
}

main() {
    if [ "${1:-}" = "--down" ]; then
        check_prereqs
        down false
        exit 0
    fi
    if [ "${1:-}" = "--clean" ]; then
        check_prereqs
        down true
        exit 0
    fi

    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam - Start Server${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    check_prereqs

    if [ ! -d "${REPO_ROOT}/crypto-config" ] || [ ! -d "${REPO_ROOT}/channel-artifacts" ]; then
        log_info "No crypto/channel artifacts found — running setup-fabric.sh..."
        ./scripts/setup-fabric.sh
    else
        log_ok "Crypto/channel artifacts already present, skipping setup"
    fi

    if container_running "peer0.org1.nanayam.com"; then
        log_ok "Fabric network is already running, skipping start-fabric.sh"
    else
        log_info "Fabric network isn't running — running start-fabric.sh..."
        ./scripts/start-fabric.sh
    fi

    if container_running "basic-org1"; then
        log_ok "Chaincode is already deployed, skipping deploy-chaincode.sh"
    else
        log_info "Chaincode isn't deployed — running deploy-chaincode.sh..."
        ./scripts/deploy-chaincode.sh
    fi

    log_info "Starting the gateway (server only — no operator console, no client apps)..."
    compose -f "${APPS_COMPOSE}" up -d --build gateway

    wait_for_gateway

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Nanayam server is up${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "  Gateway REST:  http://localhost:8080"
    echo "  Gateway gRPC:  grpc://localhost:50051"
    echo ""
    echo "Try it:"
    echo "  curl http://localhost:8080/health"
    echo "  curl -X POST http://localhost:8080/v1/Login -d '{\"username\":\"admin\",\"password\":\"admin\"}'"
    echo ""
    echo "Stop it:   ./scripts/start-server.sh --down"
    echo "Wipe it:   ./scripts/start-server.sh --clean"
    echo ""
}

main "$@"
