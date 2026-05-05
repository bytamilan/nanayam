#!/usr/bin/env bash
# =============================================================================
# Anti-Corruption Complaint System - Setup Script
# =============================================================================
# Downloads Fabric binaries, pulls images, generates crypto for 4 orgs,
# and creates channel artifacts for the complaint channel.
#
# Usage:
#   ./scripts/setup-complaint.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

FABRIC_VERSION="2.5.9"
CA_VERSION="1.5.12"
BIN_DIR="${PWD}/bin"
CRYPTO_DIR="${PWD}/crypto-config"
CHANNEL_DIR="${PWD}/channel-artifacts"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

check_prereqs() {
    log_info "Checking prerequisites..."
    command -v docker &>/dev/null || { log_err "Docker is required."; exit 1; }
    docker compose version &>/dev/null || docker-compose version &>/dev/null || { log_err "Docker Compose is required."; exit 1; }
    command -v curl &>/dev/null || { log_err "curl is required."; exit 1; }
    log_ok "Prerequisites satisfied"
}

pull_images() {
    log_info "Pulling Hyperledger Fabric Docker images (v${FABRIC_VERSION})..."
    local images=(
        "hyperledger/fabric-peer:${FABRIC_VERSION}"
        "hyperledger/fabric-orderer:${FABRIC_VERSION}"
        "hyperledger/fabric-tools:${FABRIC_VERSION}"
        "hyperledger/fabric-ca:${CA_VERSION}"
        "couchdb:3.3.2"
    )
    for img in "${images[@]}"; do
        log_info "  → ${img}"
        docker pull "${img}" || log_warn "Failed to pull ${img}"
    done
    log_ok "Docker images pulled"
}

download_binaries() {
    log_info "Downloading Fabric binaries..."
    mkdir -p "${BIN_DIR}"

    local platform=$(uname | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    case "${arch}" in
        x86_64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac

    local fabric_url="https://github.com/hyperledger/fabric/releases/download/v${FABRIC_VERSION}/hyperledger-fabric-${platform}-${arch}-${FABRIC_VERSION}.tar.gz"
    log_info "  → ${fabric_url}"
    curl -L "${fabric_url}" -o /tmp/fabric-binaries.tar.gz
    rm -rf /tmp/fabric-extract && mkdir -p /tmp/fabric-extract
    tar -xzf /tmp/fabric-binaries.tar.gz -C /tmp/fabric-extract
    mv /tmp/fabric-extract/bin/* "${BIN_DIR}/" 2>/dev/null || true

    local ca_url="https://github.com/hyperledger/fabric-ca/releases/download/v${CA_VERSION}/hyperledger-fabric-ca-${platform}-${arch}-${CA_VERSION}.tar.gz"
    log_info "  → ${ca_url}"
    curl -L "${ca_url}" -o /tmp/fabric-ca-binaries.tar.gz
    rm -rf /tmp/fabric-ca-extract && mkdir -p /tmp/fabric-ca-extract
    tar -xzf /tmp/fabric-ca-binaries.tar.gz -C /tmp/fabric-ca-extract
    mv /tmp/fabric-ca-extract/bin/* "${BIN_DIR}/" 2>/dev/null || true

    chmod +x "${BIN_DIR}"/*
    log_ok "Binaries downloaded to ${BIN_DIR}"
}

generate_crypto() {
    log_info "Generating cryptographic materials for 4-org complaint network..."
    export FABRIC_CFG_PATH="${PWD}/config"

    rm -rf "${CRYPTO_DIR}"
    mkdir -p "${CRYPTO_DIR}"

    "${BIN_DIR}/cryptogen" generate --config="${PWD}/config/crypto-config-complaint.yaml" --output="${CRYPTO_DIR}"
    log_ok "Cryptographic materials generated in ${CRYPTO_DIR}"
}

generate_channel_artifacts() {
    log_info "Creating channel artifacts for complaint-channel..."
    export FABRIC_CFG_PATH="${PWD}/config"

    rm -rf "${CHANNEL_DIR}"
    mkdir -p "${CHANNEL_DIR}"

    # Genesis block
    "${BIN_DIR}/configtxgen" -profile ComplaintOrdererGenesis -channelID system-channel -outputBlock "${CHANNEL_DIR}/genesis.block"

    # Channel creation tx
    "${BIN_DIR}/configtxgen" -profile ComplaintChannel -outputCreateChannelTx "${CHANNEL_DIR}/complaint-channel.tx" -channelID complaint-channel

    # Anchor peer updates
    "${BIN_DIR}/configtxgen" -profile ComplaintChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/ACBMSPanchors.tx" -channelID complaint-channel -asOrg ACBMSP
    "${BIN_DIR}/configtxgen" -profile ComplaintChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/DeptMSPanchors.tx" -channelID complaint-channel -asOrg DeptMSP
    "${BIN_DIR}/configtxgen" -profile ComplaintChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/OversightMSPanchors.tx" -channelID complaint-channel -asOrg OversightMSP
    "${BIN_DIR}/configtxgen" -profile ComplaintChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/JudiciaryMSPanchors.tx" -channelID complaint-channel -asOrg JudiciaryMSP

    log_ok "Channel artifacts created in ${CHANNEL_DIR}"
}

print_summary() {
    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Complaint System Setup Complete!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start the network:  ./scripts/start-complaint.sh"
    echo "  2. Deploy chaincode:   ./scripts/deploy-complaint.sh"
    echo "  3. Start apps:         ./scripts/start-complaint-apps.sh"
    echo ""
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Anti-Corruption Complaint System - Setup${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    check_prereqs
    pull_images
    download_binaries
    generate_crypto
    generate_channel_artifacts
    print_summary
}

main "$@"
