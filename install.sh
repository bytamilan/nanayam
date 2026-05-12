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
RELEASE_BASE_URL="${NANAYAM_RELEASE_BASE_URL:-https://github.com/${REPO}/releases/download}"

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
DEV_LOCAL=false
REFRESH=false
SOURCE_PATH=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --with-fabric) WITH_FABRIC=true; shift ;;
        --setup) SETUP=true; shift ;;
        --version) VERSION="$2"; shift 2 ;;
        --dev-local) DEV_LOCAL=true; shift ;;
        --refresh) REFRESH=true; shift ;;
        --source) SOURCE_PATH="$2"; shift 2 ;;
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
    echo "${os} ${arch}"
}

resolve_version() {
    local tag="${VERSION}"

    if [[ "${tag}" == "latest" ]]; then
        tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ -z "${tag}" ]]; then
            log_err "Could not determine latest release version"
            exit 1
        fi
    elif [[ "${tag}" != v* ]]; then
        tag="v${tag}"
    fi

    echo "${tag}"
}

asset_name() {
    local tag="$1"
    local os="$2"
    local arch="$3"
    local ext="tar.gz"
    if [[ "${os}" == "windows" ]]; then
        ext="zip"
    fi
    echo "${BINARY_NAME}_${tag}_${os}_${arch}.${ext}"
}

release_url() {
    local tag="$1"
    local asset="$2"
    echo "${RELEASE_BASE_URL%/}/${tag}/${asset}"
}

get_installed_version() {
    local binary_path="$1"
    if [[ -x "${binary_path}" ]]; then
        "${binary_path}" version 2>/dev/null | awk '/^nanayam version / {print $3; exit}' || true
    fi
}

download_binary() {
    local os="$1"
    local arch="$2"
    local dest="$3"
    local tag="$4"
    local asset
    local url
    local tmp_dir

    asset=$(asset_name "${tag}" "${os}" "${arch}")
    url=$(release_url "${tag}" "${asset}")

    log_info "Downloading ${BINARY_NAME} ${tag} for ${os}/${arch}..."
    log_info "  → ${url}"

    tmp_dir=$(mktemp -d)

    curl -fsSL "${url}" -o "${tmp_dir}/${asset}"
    tar -xzf "${tmp_dir}/${asset}" -C "${tmp_dir}"
    mv "${tmp_dir}/${BINARY_NAME}" "${dest}"
    chmod +x "${dest}"
    rm -rf "${tmp_dir}"
    log_ok "Binary downloaded to ${dest}"
}

resolve_local_source() {
    local candidate="${SOURCE_PATH}"
    if [[ -z "${candidate}" ]]; then
        candidate="$(pwd)"
    fi

    if [[ -f "${candidate}/cli/go.mod" ]]; then
        echo "${candidate}"
        return 0
    fi

    if [[ "$(basename "${candidate}")" == "cli" && -f "${candidate}/go.mod" ]]; then
        dirname "${candidate}"
        return 0
    fi

    log_err "Could not find a Nanayam repository. Use --source /path/to/nanayam with --dev-local."
    exit 1
}

build_local_binary() {
    local dest="$1"
    local repo_root="$2"
    local build_version
    local build_commit
    local build_date

    if ! command -v go >/dev/null 2>&1; then
        log_err "Go is required for --dev-local installs"
        exit 1
    fi

    build_version=$(git -C "${repo_root}" describe --tags --always --dirty 2>/dev/null || echo "dev-local")
    build_commit=$(git -C "${repo_root}" rev-parse --short HEAD 2>/dev/null || echo "local")
    build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    log_info "Building local Nanayam CLI from ${repo_root}..."
    (
        cd "${repo_root}/cli"
        go build \
          -ldflags "-X github.com/bytamilan/nanayam/cli/cmd.version=${build_version} -X github.com/bytamilan/nanayam/cli/cmd.commit=${build_commit} -X github.com/bytamilan/nanayam/cli/cmd.date=${build_date}" \
          -o "${dest}" .
    )
    chmod +x "${dest}"
    log_ok "Local build installed to ${dest}"
}

add_to_path() {
    local shell_cfg=""
    case "${SHELL}" in
        */zsh) shell_cfg="${HOME}/.zshrc" ;;
        */bash) shell_cfg="${HOME}/.bashrc" ;;
        *) shell_cfg="${HOME}/.profile" ;;
    esac

    touch "${shell_cfg}"
    if ! grep -q "\.nanayam/bin" "${shell_cfg}"; then
        echo "" >> "${shell_cfg}"
        echo "# Nanayam CLI" >> "${shell_cfg}"
        echo 'export PATH="$HOME/.nanayam/bin:$PATH"' >> "${shell_cfg}"
        log_ok "Added ~/.nanayam/bin to PATH in ${shell_cfg}"
        log_warn "Run 'source ${shell_cfg}' or restart your terminal to apply PATH changes"
    else
        log_info "~/.nanayam/bin already in PATH"
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

    read -r platform_os platform_arch <<< "$(detect_platform)"
    log_info "Detected platform: ${platform_os}/${platform_arch}"

    mkdir -p "${INSTALL_DIR}"
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    local target_version=""
    local current_version=""

    if [[ "${DEV_LOCAL}" == "true" ]]; then
        local repo_root
        repo_root=$(resolve_local_source)
        build_local_binary "${binary_path}" "${repo_root}"
    else
        target_version=$(resolve_version)
        current_version=$(get_installed_version "${binary_path}")

        if [[ -n "${current_version}" && "${current_version}" == "${target_version}" && "${REFRESH}" != "true" ]]; then
            log_ok "Nanayam ${target_version} is already installed. Use --refresh to reinstall it."
        else
            download_binary "${platform_os}" "${platform_arch}" "${binary_path}" "${target_version}"
        fi
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
    echo "Upgrade check: nanayam upgrade --check"
    echo "Help: nanayam --help"
    echo ""
}

main "$@"
