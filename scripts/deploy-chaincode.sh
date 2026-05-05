#!/usr/bin/env bash
# =============================================================================
# Nanayam - Deploy Chaincode
# =============================================================================
# Packages, installs, approves, and commits the asset-transfer-basic chaincode
# on the 'mychannel' channel.
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
CC_PATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/asset-transfer-basic"
CC_LABEL="${CC_NAME}_${CC_VERSION}"
CHANNEL="mychannel"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

# Ensure sample chaincode exists
ensure_chaincode() {
    if [ ! -d "${PWD}/chaincode/asset-transfer-basic" ]; then
        log_info "Downloading sample chaincode..."
        mkdir -p "${PWD}/chaincode"
        curl -L https://github.com/hyperledger/fabric-samples/archive/main.tar.gz -o /tmp/fabric-samples.tar.gz
        tar -xzf /tmp/fabric-samples.tar.gz -C /tmp/
        cp -r "/tmp/fabric-samples-main/asset-transfer-basic/chaincode-go" "${PWD}/chaincode/asset-transfer-basic"
        rm -rf /tmp/fabric-samples-main /tmp/fabric-samples.tar.gz
        log_ok "Sample chaincode downloaded"
    fi
}

# Package chaincode
package_cc() {
    log_info "Packaging chaincode..."
    docker exec cli peer lifecycle chaincode package "${CC_NAME}.tar.gz" \
        --path "${CC_PATH}" \
        --lang golang \
        --label "${CC_LABEL}"
    log_ok "Chaincode packaged"
}

# Install on Org1 peer
install_org1() {
    log_info "Installing chaincode on Org1 peer..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org1.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.nanayam.com/users/Admin@org1.nanayam.com/msp" \
        cli peer lifecycle chaincode install "${CC_NAME}.tar.gz"
    log_ok "Installed on Org1"
}

# Install on Org2 peer
install_org2() {
    log_info "Installing chaincode on Org2 peer..."
    docker exec -e CORE_PEER_ADDRESS="peer0.org2.nanayam.com:9051" \
        -e CORE_PEER_LOCALMSPID="Org2MSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.nanayam.com/users/Admin@org2.nanayam.com/msp" \
        cli peer lifecycle chaincode install "${CC_NAME}.tar.gz"
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

# Init chaincode (optional, but good for asset-transfer-basic)
init_cc() {
    log_info "Initializing chaincode..."
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
    echo -e "${BLUE}  Nanayam - Deploy Chaincode${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    ensure_chaincode
    package_cc
    install_org1
    install_org2
    get_package_id
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
    echo ""
    echo "Next step:  ./scripts/start-distribution.sh"
    echo ""
}

main "$@"
