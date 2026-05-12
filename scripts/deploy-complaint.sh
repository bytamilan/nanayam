#!/usr/bin/env bash
# =============================================================================
# Anti-Corruption Complaint System - Deploy Chaincode
# =============================================================================
# Builds and deploys the complaint-system chaincode using CCAAS mode.
# Installs on all 4 orgs, approves with a MAJORITY endorsement policy,
# and initializes the ledger.
#
# Usage:
#   ./scripts/deploy-complaint.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

CC_NAME="complaint"
CC_VERSION="1.0"
CC_SEQ="1"
CC_LABEL="${CC_NAME}_${CC_VERSION}"
CHANNEL="complaint-channel"
CC_ADDRESS="complaint:9999"

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

build_chaincode_image() {
    log_info "Building complaint chaincode Docker image..."
    docker build -t "${CC_LABEL}" -f "${PWD}/chaincode/complaint-system/Dockerfile" "${PWD}/chaincode/complaint-system"
    log_ok "Chaincode image built: ${CC_LABEL}"
}

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

    tar -czf "${pkg_dir}/code.tar.gz" -C "${pkg_dir}" connection.json

    cat > "${pkg_dir}/metadata.json" <<EOF
{
  "type": "ccaas",
  "label": "${CC_LABEL}"
}
EOF

    tar -czf "${PWD}/${CC_NAME}.tar.gz" -C "${pkg_dir}" code.tar.gz metadata.json
    log_ok "CCAAS package created: ${PWD}/${CC_NAME}.tar.gz"
}

install_cc() {
    log_info "Copying chaincode package into cli container..."
    docker cp "${PWD}/${CC_NAME}.tar.gz" cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/${CC_NAME}.tar.gz

    local orgs=(
        "peer0.acb.nanayam.com:7051 ACBMSP /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp"
        "peer0.dept.nanayam.com:9051 DeptMSP /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/peers/peer0.dept.nanayam.com/tls/ca.crt /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/users/Admin@dept.nanayam.com/msp"
        "peer0.oversight.nanayam.com:10051 OversightMSP /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/peers/peer0.oversight.nanayam.com/tls/ca.crt /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/users/Admin@oversight.nanayam.com/msp"
        "peer0.judiciary.nanayam.com:11051 JudiciaryMSP /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/peers/peer0.judiciary.nanayam.com/tls/ca.crt /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/users/Admin@judiciary.nanayam.com/msp"
    )

    for org in "${orgs[@]}"; do
        read -r peer msp tls msp_path <<< "${org}"
        log_info "Installing on ${msp}..."
        docker exec -e CORE_PEER_ADDRESS="${peer}" \
            -e CORE_PEER_LOCALMSPID="${msp}" \
            -e CORE_PEER_TLS_ROOTCERT_FILE="${tls}" \
            -e CORE_PEER_MSPCONFIGPATH="${msp_path}" \
            cli peer lifecycle chaincode install "/opt/gopath/src/github.com/hyperledger/fabric/peer/${CC_NAME}.tar.gz"
        log_ok "Installed on ${msp}"
    done
}

get_package_id() {
    log_info "Querying installed chaincode..."
    docker exec -e CORE_PEER_ADDRESS="peer0.acb.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="ACBMSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp" \
        cli peer lifecycle chaincode queryinstalled >& /tmp/log.txt
    cat /tmp/log.txt
    CC_PACKAGE_ID=$(sed -n 's/.*Package ID: \([^,]*\),.*/\1/p' /tmp/log.txt | head -1)
    log_ok "Package ID: ${CC_PACKAGE_ID}"
}

approve_org() {
    local peer=$1
    local msp=$2
    local tls=$3
    local msp_path=$4

    log_info "Approving chaincode for ${msp}..."
    docker exec -e CORE_PEER_ADDRESS="${peer}" \
        -e CORE_PEER_LOCALMSPID="${msp}" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="${tls}" \
        -e CORE_PEER_MSPCONFIGPATH="${msp_path}" \
        cli peer lifecycle chaincode approveformyorg \
            -o orderer.nanayam.com:7050 \
            --channelID "${CHANNEL}" \
            --name "${CC_NAME}" \
            --version "${CC_VERSION}" \
            --package-id "${CC_PACKAGE_ID}" \
            --sequence "${CC_SEQ}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem \
            --collections-config /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/complaint-system/collections_config.json
    log_ok "Approved for ${msp}"
}

commit_cc() {
    log_info "Committing chaincode..."
    docker exec -e CORE_PEER_ADDRESS="peer0.acb.nanayam.com:7051" \
        -e CORE_PEER_LOCALMSPID="ACBMSP" \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp" \
        cli peer lifecycle chaincode commit \
            -o orderer.nanayam.com:7050 \
            --channelID "${CHANNEL}" \
            --name "${CC_NAME}" \
            --version "${CC_VERSION}" \
            --sequence "${CC_SEQ}" \
            --tls \
            --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem \
            --peerAddresses peer0.acb.nanayam.com:7051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt \
            --peerAddresses peer0.dept.nanayam.com:9051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/peers/peer0.dept.nanayam.com/tls/ca.crt \
            --peerAddresses peer0.oversight.nanayam.com:10051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/peers/peer0.oversight.nanayam.com/tls/ca.crt \
            --peerAddresses peer0.judiciary.nanayam.com:11051 \
            --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/peers/peer0.judiciary.nanayam.com/tls/ca.crt \
            --collections-config /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/complaint-system/collections_config.json
    log_ok "Chaincode committed"
}

start_chaincode_container() {
    log_info "Starting chaincode container..."
    docker run -d \
        --name "${CC_NAME}" \
        --network nanayam \
        -e CHAINCODE_ID="${CC_PACKAGE_ID}" \
        -e CORE_CHAINCODE_ID_NAME="${CC_PACKAGE_ID}" \
        -e CHAINCODE_SERVER_ADDRESS="0.0.0.0:9999" \
        "${CC_LABEL}"
    log_ok "Chaincode container started: ${CC_NAME}"
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Deploy Complaint Chaincode${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    build_chaincode_image
    package_cc
    install_cc
    get_package_id
    start_chaincode_container

    # Approve for all 4 orgs
    approve_org "peer0.acb.nanayam.com:7051" "ACBMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/peers/peer0.acb.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/acb.nanayam.com/users/Admin@acb.nanayam.com/msp"

    approve_org "peer0.dept.nanayam.com:9051" "DeptMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/peers/peer0.dept.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dept.nanayam.com/users/Admin@dept.nanayam.com/msp"

    approve_org "peer0.oversight.nanayam.com:10051" "OversightMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/peers/peer0.oversight.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/oversight.nanayam.com/users/Admin@oversight.nanayam.com/msp"

    approve_org "peer0.judiciary.nanayam.com:11051" "JudiciaryMSP" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/peers/peer0.judiciary.nanayam.com/tls/ca.crt" \
        "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/judiciary.nanayam.com/users/Admin@judiciary.nanayam.com/msp"

    commit_cc

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Complaint Chaincode deployed!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Name:     ${CC_NAME}"
    echo "Version:  ${CC_VERSION}"
    echo "Channel:  ${CHANNEL}"
    echo ""
    echo "Next step:  ./scripts/start-complaint-apps.sh"
    echo ""
}

main "$@"
