#!/usr/bin/env bash
# =============================================================================
# Nanayam - Start Fabric Network
# =============================================================================
# Brings up the Docker Compose Fabric network, creates the channel 'mychannel',
# and joins Org1 and Org2 peers.
#
# Usage:
#   ./scripts/start-fabric.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="${PWD}/docker/fabric-network.yaml"
BIN_DIR="${PWD}/bin"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

export PATH="${BIN_DIR}:${PATH}"
export FABRIC_CFG_PATH="${PWD}/config"

# Prefer the "docker compose" v2 plugin; fall back to the standalone v1
# binary for hosts that still only have that installed.
compose() {
    if docker compose version &>/dev/null; then
        docker compose "$@"
    else
        docker-compose "$@"
    fi
}

# Helper to execute peer commands inside the cli container
cli_exec() {
    docker exec -e CORE_PEER_ADDRESS=$1 -e CORE_PEER_LOCALMSPID=$2 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=$3 \
        -e CORE_PEER_MSPCONFIGPATH=$4 cli peer channel "$5"
}

start_containers() {
    log_info "Starting Fabric network containers..."
    compose -f "${COMPOSE_FILE}" up -d
    log_ok "Containers started"

    log_info "Waiting for services to initialize..."
    sleep 5
}

create_channel() {
    log_info "Creating channel 'mychannel'..."

    docker exec cli peer channel create \
        -o orderer.nanayam.com:7050 \
        -c mychannel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/channel.tx \
        --tls \
        --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem

    log_ok "Channel 'mychannel' created"
}

join_channel() {
    local peer=$1
    local msp=$2
    local tls=$3
    local msp_path=$4

    log_info "Joining ${peer} to 'mychannel'..."

    docker exec -e CORE_PEER_ADDRESS="${peer}" \
        -e CORE_PEER_LOCALMSPID="${msp}" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="${tls}" \
        -e CORE_PEER_MSPCONFIGPATH="${msp_path}" \
        cli peer channel join -b mychannel.block

    log_ok "${peer} joined 'mychannel'"
}

update_anchor_peers() {
    local peer=$1
    local msp=$2
    local tls=$3
    local msp_path=$4
    local anchor_tx=$5

    log_info "Updating anchor peers for ${msp}..."

    docker exec -e CORE_PEER_ADDRESS="${peer}" \
        -e CORE_PEER_LOCALMSPID="${msp}" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="${tls}" \
        -e CORE_PEER_MSPCONFIGPATH="${msp_path}" \
        cli peer channel update \
            -o orderer.nanayam.com:7050 \
            -c mychannel \
            -f "${anchor_tx}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem

    log_ok "Anchor peers updated for ${msp}"
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam - Start Fabric Network${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    if [ ! -d "${PWD}/crypto-config" ] || [ ! -d "${PWD}/channel-artifacts" ]; then
        log_err "Missing crypto-config or channel-artifacts. Run ./scripts/setup-fabric.sh first."
        exit 1
    fi

    start_containers
    create_channel

    # Org1 peer0
    join_channel \
        "peer0.org1.nanayam.com:7051" \
        "Org1MSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp"

    # Org2 peer0
    join_channel \
        "peer0.org2.nanayam.com:9051" \
        "Org2MSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/users/Admin@org2.nanayam.com/msp"

    update_anchor_peers \
        "peer0.org1.nanayam.com:7051" \
        "Org1MSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org1MSPanchors.tx"

    update_anchor_peers \
        "peer0.org2.nanayam.com:9051" \
        "Org2MSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/users/Admin@org2.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org2MSPanchors.tx"

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Fabric Network is up and running!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Channel:    mychannel"
    echo "Orderer:    orderer.nanayam.com:7050"
    echo "Peer Org1:  peer0.org1.nanayam.com:7051"
    echo "Peer Org2:  peer0.org2.nanayam.com:9051"
    echo ""
    echo "Next step:  ./scripts/deploy-chaincode.sh"
    echo ""
}

main "$@"
