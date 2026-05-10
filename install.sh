#!/usr/bin/env bash
# =============================================================================
# Nanayam CLI Installer
# =============================================================================
# One-liner install script for macOS & Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash -s -- --with-fabric --setup
# =============================================================================

set -euo pipefail

REPO="bytamilan/nanayam"
BINARY_NAME="nanayam"
INSTALL_DIR="${HOME}/.nanayam/bin"
FABRIC_BIN_DIR="${HOME}/.nanayam/fabric-bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

# Parse arguments
WITH_FABRIC=false
SETUP=false
VERSION="latest"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --with-fabric) WITH_FABRIC=true; shift ;;
        --setup) SETUP=true; shift ;;
        --version) VERSION="$2"; shift 2 ;;
        *) shift ;;
    esac
done

detect_platform() {
    local os=$(uname | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    case "${arch}" in
        x86_64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_err "Unsupported architecture: ${arch}"; exit 1 ;;
    esac
    echo "${os}-${arch}"
}

download_binary() {
    local platform="$1"
    local dest="$2"
    local tag="${VERSION}"

    if [[ "${tag}" == "latest" ]]; then
        tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ -z "${tag}" ]]; then
            log_err "Could not determine latest release version"
            exit 1
        fi
    fi

    local url="https://github.com/${REPO}/releases/download/${tag}/${BINARY_NAME}-${platform}"
    log_info "Downloading ${BINARY_NAME} ${tag} for ${platform}..."
    log_info "  → ${url}"

    curl -fsSL "${url}" -o "${dest}"
    chmod +x "${dest}"
    log_ok "Binary downloaded to ${dest}"
}

add_to_path() {
    local shell_cfg=""
    case "${SHELL}" in
        */zsh) shell_cfg="${HOME}/.zshrc" ;;
        */bash) shell_cfg="${HOME}/.bashrc" ;;
        *) shell_cfg="${HOME}/.profile" ;;
    esac

    if [[ -f "${shell_cfg}" ]]; then
        if ! grep -q "\.nanayam/bin" "${shell_cfg}"; then
            echo "" >> "${shell_cfg}"
            echo "# Nanayam CLI" >> "${shell_cfg}"
            echo 'export PATH="$HOME/.nanayam/bin:$PATH"' >> "${shell_cfg}"
            log_ok "Added ~/.nanayam/bin to PATH in ${shell_cfg}"
            log_warn "Run 'source ${shell_cfg}' or restart your terminal to apply PATH changes"
        else
            log_info "~/.nanayam/bin already in PATH"
        fi
    fi
}

download_fabric_binaries() {
    local fabric_version="2.5.9"
    local ca_version="1.5.12"
    local platform=$(uname | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    case "${arch}" in
        x86_64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac

    mkdir -p "${FABRIC_BIN_DIR}"

    log_info "Downloading Fabric binaries (v${fabric_version})..."
    local fabric_url="https://github.com/hyperledger/fabric/releases/download/v${fabric_version}/hyperledger-fabric-${platform}-${arch}-${fabric_version}.tar.gz"
    curl -fsSL "${fabric_url}" -o /tmp/fabric-binaries.tar.gz
    rm -rf /tmp/fabric-extract && mkdir -p /tmp/fabric-extract
    tar -xzf /tmp/fabric-binaries.tar.gz -C /tmp/fabric-extract
    mv /tmp/fabric-extract/bin/* "${FABRIC_BIN_DIR}/" 2>/dev/null || true

    log_info "Downloading Fabric CA binaries (v${ca_version})..."
    local ca_url="https://github.com/hyperledger/fabric-ca/releases/download/v${ca_version}/hyperledger-fabric-ca-${platform}-${arch}-${ca_version}.tar.gz"
    curl -fsSL "${ca_url}" -o /tmp/fabric-ca-binaries.tar.gz
    rm -rf /tmp/fabric-ca-extract && mkdir -p /tmp/fabric-ca-extract
    tar -xzf /tmp/fabric-ca-binaries.tar.gz -C /tmp/fabric-ca-extract
    mv /tmp/fabric-ca-extract/bin/* "${FABRIC_BIN_DIR}/" 2>/dev/null || true

    chmod +x "${FABRIC_BIN_DIR}"/*
    log_ok "Fabric binaries installed to ${FABRIC_BIN_DIR}"
}

main() {
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${BLUE}  Nanayam CLI Installer${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo ""

    local platform=$(detect_platform)
    log_info "Detected platform: ${platform}"

    mkdir -p "${INSTALL_DIR}"
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"

    # If running from repo in dev mode, symlink current binary
    if [[ -n "${NANAYAM_DEV:-}" ]] && command -v go &>/dev/null; then
        log_info "Dev mode detected — building from source..."
        (cd "$(dirname "$0")/cli" && go build -o "${binary_path}" .)
    else
        download_binary "${platform}" "${binary_path}"
    fi

    add_to_path

    if [[ "${WITH_FABRIC}" == "true" ]]; then
        echo ""
        download_fabric_binaries
    fi

    if [[ "${SETUP}" == "true" ]]; then
        echo ""
        log_info "Running prerequisites check..."
        "${binary_path}" prerequisites --auto
    fi

    echo ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo ""
    echo "Run: nanayam version"
    echo "Help: nanayam --help"
    echo ""
}

main "$@"
