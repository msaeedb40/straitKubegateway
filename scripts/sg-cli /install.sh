#!/usr/bin/env bash
# ============================================================================
# straitKubegateway — sg-cli: Install (amd64, arm64)
# ============================================================================
# Installs the sg-cli management command-line binary.
# Supports Linux and macOS (Darwin) for amd64 and arm64 architectures.
#
# Usage:
#   ./install.sh                     # Interactive or direct local install
#   VERSION=v1.0.1 ./install.sh      # Specify a release version
#   INSTALL_DIR=/usr/local/bin ./install.sh
#   curl -fsSL https://raw.githubusercontent.com/msaeedb40/straitKubegateway/developer/scripts/sg-cli/install.sh | bash
# ============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Colors & Formatting
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info()    { echo -e "${CYAN}[INFO]${NC}  $*"; }
log_success() { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }
die()         { log_error "$@"; exit 1; }

echo -e "${BOLD}${CYAN}======================================================${NC}"
echo -e "${BOLD}${CYAN}   straitKubegateway — sg-cli Installer              ${NC}"
echo -e "${BOLD}${CYAN}======================================================${NC}"

# ---------------------------------------------------------------------------
# Architecture & OS Detection
# ---------------------------------------------------------------------------
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  *)       die "Unsupported operating system: ${OS}. sg-cli supports Linux and macOS." ;;
esac

ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             die "Unsupported processor architecture: ${ARCH}. sg-cli supports amd64 and arm64." ;;
esac

log_info "Detected Platform: ${OS}/${ARCH}"

# ---------------------------------------------------------------------------
# Configuration Defaults
# ---------------------------------------------------------------------------
VERSION="${VERSION:-latest}"
REPO="${REPO:-msaeedb40/straitKubegateway}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="sg-cli"
TMP_DIR="$(mktemp -d /tmp/sg-cli-install.XXXXXX)"
cleanup() { rm -rf "${TMP_DIR}"; }
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Check Sudo Requirement for Target Directory
# ---------------------------------------------------------------------------
SUDO=""
mkdir -p "${INSTALL_DIR}" 2>/dev/null || true

if [ ! -w "${INSTALL_DIR}" ]; then
  if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo &>/dev/null; then
      log_warn "${INSTALL_DIR} requires administrative privileges. Using sudo."
      SUDO="sudo"
    else
      FALLBACK_DIR="${HOME}/.local/bin"
      log_warn "Cannot write to ${INSTALL_DIR} and sudo is not available."
      log_info "Falling back to user directory: ${FALLBACK_DIR}"
      INSTALL_DIR="${FALLBACK_DIR}"
      mkdir -p "${INSTALL_DIR}"
    fi
  fi
fi

# ---------------------------------------------------------------------------
# Install Strategy: Local Source Build (if in repo) vs Remote Download
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || echo "")"
REPO_ROOT=""
if [ -n "${SCRIPT_DIR}" ] && [ -d "${SCRIPT_DIR}/../../cmd/sg-cli" ]; then
  REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
elif [ -d "./cmd/sg-cli" ]; then
  REPO_ROOT="$(pwd)"
fi

INSTALLED=false

if [ -n "${REPO_ROOT}" ] && [ -d "${REPO_ROOT}/cmd/sg-cli" ] && command -v go &>/dev/null; then
  log_info "Found local source repository at: ${REPO_ROOT}"
  log_info "Building ${BINARY_NAME} from source using Go $(go env GOVERSION)..."
  
  TARGET_PATH="${TMP_DIR}/${BINARY_NAME}"
  (
    cd "${REPO_ROOT}"
    CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" go build \
      -ldflags="-s -w -X 'github.com/straitkubegateway/straitkubegateway/cmd/sg-cli.Version=${VERSION}'" \
      -o "${TARGET_PATH}" \
      ./cmd/sg-cli
  )
  
  ${SUDO} mkdir -p "${INSTALL_DIR}"
  ${SUDO} install -m 755 "${TARGET_PATH}" "${INSTALL_DIR}/${BINARY_NAME}"
  INSTALLED=true
  log_success "Built and installed ${BINARY_NAME} from local source to ${INSTALL_DIR}/${BINARY_NAME}"
fi

if [ "${INSTALLED}" = false ]; then
  # ---------------------------------------------------------------------------
  # Remote Binary Download
  # ---------------------------------------------------------------------------
  if ! command -v curl &>/dev/null && ! command -v wget &>/dev/null; then
    die "curl or wget is required to download pre-built binaries."
  fi

  BINARY_URL=""
  if [ "${VERSION}" = "latest" ]; then
    BINARY_URL="https://github.com/${REPO}/releases/latest/download/sg-cli-${OS}-${ARCH}"
  else
    BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/sg-cli-${OS}-${ARCH}"
  fi

  log_info "Downloading ${BINARY_NAME} (${OS}/${ARCH}) from:"
  log_info "  ${BINARY_URL}"

  TARGET_PATH="${TMP_DIR}/${BINARY_NAME}"
  DOWNLOAD_OK=false

  if command -v curl &>/dev/null; then
    if curl -fsSL "${BINARY_URL}" -o "${TARGET_PATH}" 2>/dev/null; then
      DOWNLOAD_OK=true
    fi
  elif command -v wget &>/dev/null; then
    if wget -q "${BINARY_URL}" -O "${TARGET_PATH}" 2>/dev/null; then
      DOWNLOAD_OK=true
    fi
  fi

  if [ "${DOWNLOAD_OK}" = true ] && [ -s "${TARGET_PATH}" ]; then
    chmod +x "${TARGET_PATH}"
    ${SUDO} mkdir -p "${INSTALL_DIR}"
    ${SUDO} mv "${TARGET_PATH}" "${INSTALL_DIR}/${BINARY_NAME}"
    INSTALLED=true
    log_success "Downloaded and installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
  else
    # Fallback to Go install if download fails
    if command -v go &>/dev/null; then
      log_warn "Release binary not accessible. Falling back to 'go install'..."
      GOBIN="${INSTALL_DIR}" go install "github.com/${REPO}/cmd/sg-cli@${VERSION}" || \
        go install "github.com/${REPO}/cmd/sg-cli@latest"
      INSTALLED=true
      log_success "Installed ${BINARY_NAME} via go install to ${INSTALL_DIR}/${BINARY_NAME}"
    else
      die "Failed to download binary from ${BINARY_URL}. Please ensure version '${VERSION}' exists or install Go to compile from source."
    fi
  fi
fi

# ---------------------------------------------------------------------------
# Verification
# ---------------------------------------------------------------------------
if command -v "${BINARY_NAME}" &>/dev/null; then
  INSTALLED_BIN="$(command -v "${BINARY_NAME}")"
else
  INSTALLED_BIN="${INSTALL_DIR}/${BINARY_NAME}"
fi

if [ -x "${INSTALLED_BIN}" ]; then
  echo ""
  log_success "Verification successful! Binary ready at: ${INSTALLED_BIN}"
  echo -e "${CYAN}------------------------------------------------------${NC}"
  "${INSTALLED_BIN}" version 2>/dev/null || "${INSTALLED_BIN}" --help | head -n 8
  echo -e "${CYAN}------------------------------------------------------${NC}"
  echo ""
  log_info "Run '${BINARY_NAME} status' or '${BINARY_NAME} --help' to get started."
else
  die "Installation finished but binary at ${INSTALLED_BIN} is not executable."
fi