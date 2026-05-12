#!/usr/bin/env bash
# =============================================================================
# Nanayam - Hyperledger Fabric Setup Script
# =============================================================================
# Downloads Fabric binaries (cryptogen, configtxgen, peer, etc.), pulls Docker
# images, and generates cryptographic materials & channel artifacts.
#
# Usage:
#   ./scripts/setup-fabric.sh
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

CONFIGTX_TMP_DIR=""

cleanup_configtx_env() {
    if [ -n "${CONFIGTX_TMP_DIR}" ] && [ -d "${CONFIGTX_TMP_DIR}" ]; then
        rm -rf "${CONFIGTX_TMP_DIR}"
    fi
}

prepare_configtx_env() {
    local source_file="$1"
    CONFIGTX_TMP_DIR=$(mktemp -d)
    cp "${source_file}" "${CONFIGTX_TMP_DIR}/configtx.yaml"
    export FABRIC_CFG_PATH="${CONFIGTX_TMP_DIR}"
}

trap cleanup_configtx_env EXIT

check_prereqs() {
    log_info "Checking prerequisites..."
    command -v docker &>/dev/null || { log_err "Docker is required. Install from https://docs.docker.com/get-docker/"; exit 1; }
    docker compose version &>/dev/null || docker-compose version &>/dev/null || { log_err "Docker Compose is required."; exit 1; }
    command -v curl &>/dev/null || { log_err "curl is required."; exit 1; }
    command -v jq &>/dev/null || log_warn "jq not found. Some diagnostics may be limited."
    log_ok "Prerequisites satisfied"
}

pull_images() {
    log_info "Pulling Hyperledger Fabric Docker images (v${FABRIC_VERSION})..."
    local images=(
        "hyperledger/fabric-peer:${FABRIC_VERSION}"
        "hyperledger/fabric-orderer:${FABRIC_VERSION}"
        "hyperledger/fabric-tools:${FABRIC_VERSION}"
        "hyperledger/fabric-ca:${CA_VERSION}"
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

    # Download Fabric binaries (peer, orderer, cryptogen, configtxgen, etc.)
    local fabric_url="https://github.com/hyperledger/fabric/releases/download/v${FABRIC_VERSION}/hyperledger-fabric-${platform}-${arch}-${FABRIC_VERSION}.tar.gz"
    log_info "  → ${fabric_url}"
    curl -L "${fabric_url}" -o /tmp/fabric-binaries.tar.gz
    rm -rf /tmp/fabric-extract && mkdir -p /tmp/fabric-extract
    tar -xzf /tmp/fabric-binaries.tar.gz -C /tmp/fabric-extract
    mv /tmp/fabric-extract/bin/* "${BIN_DIR}/" 2>/dev/null || true

    # Download Fabric CA binaries (fabric-ca-client, fabric-ca-server)
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
    log_info "Generating cryptographic materials..."
    export FABRIC_CFG_PATH="${PWD}/config"

    rm -rf "${CRYPTO_DIR}"
    mkdir -p "${CRYPTO_DIR}"

    "${BIN_DIR}/cryptogen" generate --config="${PWD}/config/crypto-config.yaml" --output="${CRYPTO_DIR}"
    log_ok "Cryptographic materials generated in ${CRYPTO_DIR}"
}

generate_channel_artifacts() {
    log_info "Creating channel artifacts..."
    prepare_configtx_env "${PWD}/config/configtx.yaml"

    rm -rf "${CHANNEL_DIR}"
    mkdir -p "${CHANNEL_DIR}"

    # Genesis block
    "${BIN_DIR}/configtxgen" -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock "${CHANNEL_DIR}/genesis.block"

    # Channel creation tx
    "${BIN_DIR}/configtxgen" -profile TwoOrgsChannel -outputCreateChannelTx "${CHANNEL_DIR}/channel.tx" -channelID mychannel

    # Anchor peer updates
    "${BIN_DIR}/configtxgen" -profile TwoOrgsChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/Org1MSPanchors.tx" -channelID mychannel -asOrg Org1MSP
    "${BIN_DIR}/configtxgen" -profile TwoOrgsChannel -outputAnchorPeersUpdate "${CHANNEL_DIR}/Org2MSPanchors.tx" -channelID mychannel -asOrg Org2MSP

    log_ok "Channel artifacts created in ${CHANNEL_DIR}"
}

print_summary() {
    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Fabric Setup Complete!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start the network:  ./scripts/start-fabric.sh"
    echo "  2. Stop the network:   ./scripts/stop-fabric.sh"
    echo "  3. Deploy chaincode:   ./scripts/deploy-chaincode.sh"
    echo ""
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam - Fabric Setup${NC}"
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
