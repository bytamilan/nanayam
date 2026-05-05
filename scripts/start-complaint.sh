#!/usr/bin/env bash
# =============================================================================
# Anti-Corruption Complaint System - Start Network
# =============================================================================
# Brings up the 4-org Docker Compose network, creates 'complaint-channel',
# and joins all peers.
#
# Usage:
#   ./scripts/start-complaint.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="${PWD}/docker/complaint-network.yaml"
BIN_DIR="${PWD}/bin"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

export PATH="${BIN_DIR}:${PATH}"
export FABRIC_CFG_PATH="${PWD}/config"

start_containers() {
    log_info "Starting 4-org complaint network containers..."
    docker-compose -f "${COMPOSE_FILE}" up -d
    log_ok "Containers started"

    log_info "Waiting for services to initialize..."
    sleep 8
}

create_channel() {
    log_info "Creating channel 'complaint-channel'..."

    docker exec cli peer channel create \
        -o orderer.nanayam.com:7050 \
        -c complaint-channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/complaint-channel.tx \
        --tls \
        --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem

    log_ok "Channel 'complaint-channel' created"
}

join_channel() {
    local peer=$1
    local msp=$2
    local tls=$3
    local msp_path=$4

    log_info "Joining ${peer} to 'complaint-channel'..."

    docker exec -e CORE_PEER_ADDRESS="${peer}" \
        -e CORE_PEER_LOCALMSPID="${msp}" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="${tls}" \
        -e CORE_PEER_MSPCONFIGPATH="${msp_path}" \
        cli peer channel join -b complaint-channel.block

    log_ok "${peer} joined 'complaint-channel'"
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
            -c complaint-channel \
            -f "${anchor_tx}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem

    log_ok "Anchor peers updated for ${msp}"
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Anti-Corruption Complaint System - Start${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    if [ ! -d "${PWD}/crypto-config" ] || [ ! -d "${PWD}/channel-artifacts" ]; then
        log_err "Missing crypto-config or channel-artifacts. Run ./scripts/setup-complaint.sh first."
        exit 1
    fi

    start_containers
    create_channel

    # ACB
    join_channel \
        "peer0.acb.nanayam.com:7051" \
        "ACBMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp"

    # Dept
    join_channel \
        "peer0.dept.nanayam.com:9051" \
        "DeptMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/peers/peer0.dept.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/users/Admin@dept.nanayam.com/msp"

    # Oversight
    join_channel \
        "peer0.oversight.nanayam.com:10051" \
        "OversightMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/peers/peer0.oversight.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/users/Admin@oversight.nanayam.com/msp"

    # Judiciary
    join_channel \
        "peer0.judiciary.nanayam.com:11051" \
        "JudiciaryMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/peers/peer0.judiciary.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/users/Admin@judiciary.nanayam.com/msp"

    update_anchor_peers \
        "peer0.acb.nanayam.com:7051" \
        "ACBMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/ACBMSPanchors.tx"

    update_anchor_peers \
        "peer0.dept.nanayam.com:9051" \
        "DeptMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/peers/peer0.dept.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/users/Admin@dept.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/DeptMSPanchors.tx"

    update_anchor_peers \
        "peer0.oversight.nanayam.com:10051" \
        "OversightMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/peers/peer0.oversight.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/users/Admin@oversight.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/OversightMSPanchors.tx"

    update_anchor_peers \
        "peer0.judiciary.nanayam.com:11051" \
        "JudiciaryMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/peers/peer0.judiciary.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/users/Admin@judiciary.nanayam.com/msp" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/JudiciaryMSPanchors.tx"

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Complaint Network is up and running!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Channel:         complaint-channel"
    echo "Orderer:         orderer.nanayam.com:7050"
    echo "ACB Peer:        peer0.acb.nanayam.com:7051"
    echo "Dept Peer:       peer0.dept.nanayam.com:9051"
    echo "Oversight Peer:  peer0.oversight.nanayam.com:10051"
    echo "Judiciary Peer:  peer0.judiciary.nanayam.com:11051"
    echo ""
    echo "Next step:  ./scripts/deploy-complaint.sh"
    echo ""
}

main "$@"
