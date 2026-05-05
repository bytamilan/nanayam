#!/usr/bin/env bash
# =============================================================================
# Nanayam - Deploy Chaincode (Chaincode as a Service)
# =============================================================================
# Uses CCAAS to avoid Docker-in-Docker build issues on macOS/Docker Desktop.
# The chaincode runs as its own container; peers connect to it via gRPC.
#
# Usage:
#   ./scripts/deploy-chaincode.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

CC_NAME="basic"
CC_VERSION="1.0"
CC_SEQ="1"
CC_LABEL="${CC_NAME}_${CC_VERSION}"
CHANNEL="mychannel"
CC_ADDRESS="chaincode-basic-org1:9999"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

# Build chaincode Docker image on the host
build_chaincode_image() {
    log_info "Building chaincode Docker image..."
    docker build -t "${CC_LABEL}" -f "${PWD}/chaincode/asset-transfer-basic/Dockerfile" "${PWD}/chaincode/asset-transfer-basic"
    log_ok "Chaincode image built: ${CC_LABEL}"
}

# Create CCAAS package (code.tar.gz containing connection.json + metadata.json)
package_cc() {
    log_info "Packaging chaincode for CCAAS..."
    local pkg_dir="/tmp/ccaas-${CC_NAME}"
    rm -rf "${pkg_dir}" && mkdir -p "${pkg_dir}"

    cat > "${pkg_dir}/connection.json" <<EOF
{
  "address": "${CC_ADDRESS}",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF

    # code.tar.gz contains just connection.json
    tar -czf "${pkg_dir}/code.tar.gz" -C "${pkg_dir}" connection.json

    cat > "${pkg_dir}/metadata.json" <<EOF
{
  "type": "ccaas",
  "label": "${CC_LABEL}"
}
EOF

    # Final package: code.tar.gz + metadata.json
    tar -czf "${PWD}/${CC_NAME}.tar.gz" -C "${pkg_dir}" code.tar.gz metadata.json
    log_ok "CCAAS package created: ${PWD}/${CC_NAME}.tar.gz"
}

# Copy package into cli container and install on peers
install_cc() {
    log_info "Copying chaincode package into cli container..."
    docker cp "${PWD}/${CC_NAME}.tar.gz" cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/${CC_NAME}.tar.gz

    log_info "Installing chaincode on Org1 peer..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer lifecycle chaincode install "/opt/gopath/src/github.com/hyperledger/fabric/peer/${CC_NAME}.tar.gz"
    log_ok "Installed on Org1"

    log_info "Installing chaincode on Org2 peer..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org2.nanayam.com:9051" \
        -e CORE_PEER_LOCALMSPID="Org2MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/users/Admin@org2.nanayam.com/msp" \
        cli peer lifecycle chaincode install "/opt/gopath/src/github.com/hyperledger/fabric/peer/${CC_NAME}.tar.gz"
    log_ok "Installed on Org2"
}

# Query installed to get package ID
get_package_id() {
    log_info "Querying installed chaincode..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer lifecycle chaincode queryinstalled >& /tmp/log.txt
    cat /tmp/log.txt
    CC_PACKAGE_ID=$(grep -oP "Package ID: \K[^:]+" /tmp/log.txt | head -1)
    log_ok "Package ID: ${CC_PACKAGE_ID}"
}

# Approve for Org1
approve_org1() {
    log_info "Approving chaincode for Org1..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer lifecycle chaincode approveformyorg \
            -o orderer.nanayam.com:7050 \
            --channelID "${CHANNEL}" \
            --name "${CC_NAME}" \
            --version "${CC_VERSION}" \
            --package-id "${CC_PACKAGE_ID}" \
            --sequence "${CC_SEQ}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem
    log_ok "Approved for Org1"
}

# Approve for Org2
approve_org2() {
    log_info "Approving chaincode for Org2..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org2.nanayam.com:9051" \
        -e CORE_PEER_LOCALMSPID="Org2MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/users/Admin@org2.nanayam.com/msp" \
        cli peer lifecycle chaincode approveformyorg \
            -o orderer.nanayam.com:7050 \
            --channelID "${CHANNEL}" \
            --name "${CC_NAME}" \
            --version "${CC_VERSION}" \
            --package-id "${CC_PACKAGE_ID}" \
            --sequence "${CC_SEQ}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem
    log_ok "Approved for Org2"
}

# Commit chaincode
commit_cc() {
    log_info "Committing chaincode..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer lifecycle chaincode commit \
            -o orderer.nanayam.com:7050 \
            --channelID "${CHANNEL}" \
            --name "${CC_NAME}" \
            --version "${CC_VERSION}" \
            --sequence "${CC_SEQ}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem \
            --peerAddresses peer0.org1.nanayam.com:7051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt \
            --peerAddresses peer0.org2.nanayam.com:9051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt
    log_ok "Chaincode committed"
}

# Start chaincode container
start_chaincode_container() {
    log_info "Starting chaincode container..."
    # Copy the package into the container path so the chaincode can read its ID
    docker run -d \
        --name "${CC_NAME}-org1" \
        --network nanayam \
        -e CHAINCODE_ID="${CC_PACKAGE_ID}" \
        -e CHAINCODE_SERVER_ADDRESS="0.0.0.0:9999" \
        "${CC_LABEL}"
    log_ok "Chaincode container started: ${CC_NAME}-org1"
}

# Init chaincode
init_cc() {
    log_info "Initializing chaincode..."
    sleep 3  # Give chaincode time to start listening
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer chaincode invoke \
            -o orderer.nanayam.com:7050 \
            -C "${CHANNEL}" \
            -n "${CC_NAME}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem \
            --peerAddresses peer0.org1.nanayam.com:7051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt \
            -c '{"function":"InitLedger","Args":[]}'
    log_ok "Chaincode initialized"
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam - Deploy Chaincode (CCAAS)${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    build_chaincode_image
    package_cc
    install_cc
    get_package_id
    start_chaincode_container
    approve_org1
    approve_org2
    commit_cc
    init_cc

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Chaincode deployed successfully!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Name:     ${CC_NAME}"
    echo "Version:  ${CC_VERSION}"
    echo "Channel:  ${CHANNEL}"
    echo "Mode:     Chaincode as a Service (CCAAS)"
    echo ""
    echo "Next step:  ./scripts/start-distribution.sh"
    echo ""
}

main "$@"
