#!/usr/bin/env bash
#
# JikiME-ADK Installer
#
# Usage:
#   curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash
#   curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --global
#   curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --version 2.1.0
#
# Alternative:
#   go install github.com/jikime/jikime-adk@latest

set -euo pipefail

# Configuration
readonly REPO="jikime/jikime-adk"
readonly BINARY_NAME="jikime-adk"
readonly WT_BINARY_NAME="jikime-wt"
readonly SYMLINK_NAME="jikime"
readonly DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
readonly GLOBAL_INSTALL_DIR="/usr/local/bin"
readonly RELEASES_URL="https://api.github.com/repos/${REPO}/releases"
readonly DOWNLOAD_BASE="https://github.com/${REPO}/releases/download"
readonly MAX_RETRIES=3
readonly RETRY_DELAY=2

# Colors (Cyberpunk theme)
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NEON_GREEN='\033[1;32m'
RED='\033[0;31m'
DIM='\033[2m'
BOLD='\033[1m'
NC='\033[0m'

# Options
INSTALL_DIR="${DEFAULT_INSTALL_DIR}"
TARGET_VERSION=""
USE_SUDO=false
NO_COLOR=false

# Cleanup handler
TMP_DIR=""
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT

# ─── Output Helpers ─────────────────────────────────────────────────

info() {
    if [ "$NO_COLOR" = true ]; then
        echo "  $1"
    else
        echo -e "  ${CYAN}$1${NC}"
    fi
}

success() {
    if [ "$NO_COLOR" = true ]; then
        echo "  $1"
    else
        echo -e "  ${NEON_GREEN}$1${NC}"
    fi
}

error() {
    if [ "$NO_COLOR" = true ]; then
        echo "ERROR: $1" >&2
    else
        echo -e "  ${RED}ERROR: $1${NC}" >&2
    fi
}

dim() {
    if [ "$NO_COLOR" = true ]; then
        echo "  $1"
    else
        echo -e "  ${DIM}$1${NC}"
    fi
}

# ─── Core Functions ─────────────────────────────────────────────────

print_banner() {
    if [ "$NO_COLOR" = true ]; then
        echo ""
        echo "  ╔════════════════════════════════════════╗"
        echo "  ║       JikiME-ADK Installer             ║"
        echo "  ╚════════════════════════════════════════╝"
        echo ""
    else
        echo ""
        echo -e "  ${MAGENTA}╔════════════════════════════════════════╗${NC}"
        echo -e "  ${MAGENTA}║${NC}  ${CYAN}${BOLD}JikiME-ADK${NC} ${DIM}Installer${NC}               ${MAGENTA}║${NC}"
        echo -e "  ${MAGENTA}╚════════════════════════════════════════╝${NC}"
        echo ""
    fi
}

check_requirements() {
    local missing=()

    if ! command -v curl &>/dev/null; then
        missing+=("curl")
    fi

    if ! command -v sha256sum &>/dev/null && ! command -v shasum &>/dev/null; then
        missing+=("sha256sum or shasum")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        error "Missing required tools: ${missing[*]}"
        echo ""
        echo "  Please install the missing tools and try again."
        exit 1
    fi
}

detect_platform() {
    local os
    os="$(uname -s)"

    case "$os" in
        Darwin) echo "darwin" ;;
        Linux)  echo "linux" ;;
        *)
            error "Unsupported platform: $os"
            echo "  Supported: macOS (darwin), Linux"
            exit 1
            ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m)"

    case "$arch" in
        x86_64|amd64)       echo "amd64" ;;
        aarch64|arm64)      echo "arm64" ;;
        *)
            error "Unsupported architecture: $arch"
            echo "  Supported: x86_64 (amd64), aarch64 (arm64)"
            exit 1
            ;;
    esac
}

get_latest_version() {
    local url="${RELEASES_URL}/latest"
    local response
    local attempt=0

    # Build authorization header if token is available
    # GITHUB_TOKEN or GH_TOKEN increases rate limit from 60/hour to 5000/hour
    local auth_header=""
    local token="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
    if [ -n "$token" ]; then
        auth_header="-H \"Authorization: Bearer $token\""
        dim "Using GitHub token for API request"
    fi

    while [ $attempt -lt $MAX_RETRIES ]; do
        attempt=$((attempt + 1))

        # Use eval to properly expand auth_header
        response=$(eval curl -fsSL -H \"Accept: application/vnd.github.v3+json\" \
            $auth_header \"$url\" 2>/dev/null) && break

        if [ $attempt -lt $MAX_RETRIES ]; then
            dim "Retry $attempt/$MAX_RETRIES in ${RETRY_DELAY}s..."
            sleep $RETRY_DELAY
        fi
    done

    if [ -z "${response:-}" ]; then
        error "Failed to fetch latest version from GitHub API"
        dim "Check your network connection and try again."
        dim "Tip: Set GITHUB_TOKEN to increase rate limit (60 → 5000 requests/hour)"
        exit 1
    fi

    # Extract tag_name
    local tag
    tag=$(echo "$response" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

    if [ -z "$tag" ]; then
        error "Could not parse version from GitHub API response"
        exit 1
    fi

    # Remove 'v' prefix if present
    echo "${tag#v}"
}

download_binary() {
    local version="$1"
    local platform="$2"
    local arch="$3"
    local dest="$4"
    local binary_name="${5:-$BINARY_NAME}"

    local asset_name="${binary_name}-${platform}-${arch}"
    local download_url="${DOWNLOAD_BASE}/v${version}/${asset_name}"

    local attempt=0
    while [ $attempt -lt $MAX_RETRIES ]; do
        attempt=$((attempt + 1))

        if curl -fsSL -o "$dest" "$download_url" 2>/dev/null; then
            return 0
        fi

        if [ $attempt -lt $MAX_RETRIES ]; then
            dim "Download retry $attempt/$MAX_RETRIES in ${RETRY_DELAY}s..."
            sleep $RETRY_DELAY
        fi
    done

    error "Failed to download binary from: $download_url"
    return 1
}

verify_checksum() {
    local file="$1"
    local version="$2"
    local platform="$3"
    local arch="$4"
    local binary_name="${5:-$BINARY_NAME}"

    local asset_name="${binary_name}-${platform}-${arch}"
    local checksums_url="${DOWNLOAD_BASE}/v${version}/checksums.txt"

    # Download checksums.txt
    local checksums_file="${TMP_DIR}/checksums.txt"
    if ! curl -fsSL -o "$checksums_file" "$checksums_url" 2>/dev/null; then
        dim "Warning: Could not download checksums.txt, skipping verification"
        return 0
    fi

    # Find expected checksum
    local expected_hash
    expected_hash=$(grep "$asset_name" "$checksums_file" | awk '{print $1}')

    if [ -z "$expected_hash" ]; then
        dim "Warning: Checksum not found for $asset_name, skipping verification"
        return 0
    fi

    # Calculate actual checksum
    local actual_hash
    if command -v sha256sum &>/dev/null; then
        actual_hash=$(sha256sum "$file" | awk '{print $1}')
    else
        actual_hash=$(shasum -a 256 "$file" | awk '{print $1}')
    fi

    if [ "$actual_hash" != "$expected_hash" ]; then
        error "Checksum mismatch!"
        echo "  Expected: $expected_hash"
        echo "  Got:      $actual_hash"
        return 1
    fi

    return 0
}

install_binary() {
    local src="$1"
    local install_dir="$2"
    local target_name="${3:-$BINARY_NAME}"

    if [ "$USE_SUDO" = true ]; then
        sudo mkdir -p "$install_dir"
        sudo cp "$src" "${install_dir}/${target_name}"
        sudo chmod 755 "${install_dir}/${target_name}"
    else
        mkdir -p "$install_dir"
        cp "$src" "${install_dir}/${target_name}"
        chmod 755 "${install_dir}/${target_name}"
    fi
}

create_symlink() {
    local install_dir="$1"
    local target="${install_dir}/${BINARY_NAME}"
    local link="${install_dir}/${SYMLINK_NAME}"

    # Remove existing symlink or file
    if [ "$USE_SUDO" = true ]; then
        sudo rm -f "$link"
        sudo ln -s "$target" "$link"
    else
        rm -f "$link"
        ln -s "$target" "$link"
    fi
}

verify_path() {
    local install_dir="$1"

    case ":$PATH:" in
        *":${install_dir}:"*) return 0 ;;
    esac

    echo ""
    dim "NOTE: ${install_dir} is not in your PATH."
    dim "Add it to your shell profile:"
    echo ""

    local shell_name
    shell_name="$(basename "${SHELL:-/bin/bash}")"

    case "$shell_name" in
        zsh)
            dim "  echo 'export PATH=\"${install_dir}:\$PATH\"' >> ~/.zshrc"
            dim "  source ~/.zshrc"
            ;;
        bash)
            dim "  echo 'export PATH=\"${install_dir}:\$PATH\"' >> ~/.bashrc"
            dim "  source ~/.bashrc"
            ;;
        fish)
            dim "  fish_add_path ${install_dir}"
            ;;
        *)
            dim "  export PATH=\"${install_dir}:\$PATH\""
            ;;
    esac
}

verify_installation() {
    local install_dir="$1"
    local binary="${install_dir}/${BINARY_NAME}"

    if [ ! -x "$binary" ]; then
        error "Installation verification failed: binary not found at $binary"
        return 1
    fi

    local installed_version
    installed_version=$("$binary" --version 2>/dev/null || true)

    if [ -n "$installed_version" ]; then
        success "Installed: ${BINARY_NAME} ${installed_version}"
    else
        success "Installed: ${BINARY_NAME} at ${binary}"
    fi
}

print_success() {
    local install_dir="$1"
    echo ""
    if [ "$NO_COLOR" = true ]; then
        echo "  Installation complete!"
    else
        echo -e "  ${NEON_GREEN}${BOLD}Installation complete!${NC}"
    fi
    echo ""
    dim "Installed:"
    dim "  ${install_dir}/${BINARY_NAME}"
    dim "  ${install_dir}/${WT_BINARY_NAME}"
    dim "  ${install_dir}/${SYMLINK_NAME} -> ${BINARY_NAME}"
    echo ""
    dim "Run 'jikime --help' to get started."
    dim "Run 'jikime update --check' to check for updates."
    echo ""
}

usage() {
    echo "JikiME-ADK Installer"
    echo ""
    echo "Usage:"
    echo "  curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash"
    echo "  curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- [options]"
    echo ""
    echo "Options:"
    echo "  --global      Install to /usr/local/bin (requires sudo)"
    echo "  --version X   Install specific version (e.g., 2.1.0)"
    echo "  --no-color    Disable colored output"
    echo "  --help        Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  GITHUB_TOKEN  GitHub personal access token (increases API rate limit)"
    echo "  GH_TOKEN      Alternative token variable (used by gh CLI)"
    echo ""
    echo "Alternative installation:"
    echo "  go install github.com/jikime/jikime-adk@latest"
    exit 0
}

# ─── Main ───────────────────────────────────────────────────────────

main() {
    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --global)
                INSTALL_DIR="${GLOBAL_INSTALL_DIR}"
                USE_SUDO=true
                shift
                ;;
            --version)
                if [ $# -lt 2 ]; then
                    error "--version requires a value"
                    exit 1
                fi
                TARGET_VERSION="$2"
                shift 2
                ;;
            --no-color)
                NO_COLOR=true
                shift
                ;;
            --help|-h)
                usage
                ;;
            *)
                error "Unknown option: $1"
                echo "  Run with --help for usage information."
                exit 1
                ;;
        esac
    done

    print_banner

    # Step 1: Check requirements
    info "Checking requirements..."
    check_requirements
    success "Requirements satisfied"

    # Step 2: Detect platform
    local platform arch
    platform=$(detect_platform)
    arch=$(detect_arch)
    info "Platform: ${platform}/${arch}"

    # Step 3: Get version
    local version
    if [ -n "$TARGET_VERSION" ]; then
        version="$TARGET_VERSION"
        info "Target version: ${version}"
    else
        info "Fetching latest version..."
        version=$(get_latest_version)
        success "Latest version: ${version}"
    fi

    # Step 4: Create temp directory
    TMP_DIR=$(mktemp -d)
    local tmp_adk="${TMP_DIR}/${BINARY_NAME}-${platform}-${arch}"
    local tmp_wt="${TMP_DIR}/${WT_BINARY_NAME}-${platform}-${arch}"

    # Step 5: Download binaries
    info "Downloading ${BINARY_NAME} v${version}..."
    if ! download_binary "$version" "$platform" "$arch" "$tmp_adk" "$BINARY_NAME"; then
        exit 1
    fi
    success "${BINARY_NAME} downloaded"

    info "Downloading ${WT_BINARY_NAME} v${version}..."
    if ! download_binary "$version" "$platform" "$arch" "$tmp_wt" "$WT_BINARY_NAME"; then
        exit 1
    fi
    success "${WT_BINARY_NAME} downloaded"

    # Step 6: Verify checksums
    info "Verifying checksums..."
    if ! verify_checksum "$tmp_adk" "$version" "$platform" "$arch" "$BINARY_NAME"; then
        error "Checksum verification failed for ${BINARY_NAME}. Aborting."
        exit 1
    fi
    if ! verify_checksum "$tmp_wt" "$version" "$platform" "$arch" "$WT_BINARY_NAME"; then
        error "Checksum verification failed for ${WT_BINARY_NAME}. Aborting."
        exit 1
    fi
    success "Checksums verified"

    # Step 7: Install binaries
    info "Installing to ${INSTALL_DIR}..."
    if [ "$USE_SUDO" = true ]; then
        dim "Sudo access required for /usr/local/bin installation"
    fi
    install_binary "$tmp_adk" "$INSTALL_DIR" "$BINARY_NAME"
    install_binary "$tmp_wt" "$INSTALL_DIR" "$WT_BINARY_NAME"
    success "Binaries installed"

    # Step 8: Create symlink (jikime -> jikime-adk)
    info "Creating symlink: ${SYMLINK_NAME} -> ${BINARY_NAME}..."
    create_symlink "$INSTALL_DIR"
    success "Symlink created"

    # Step 9: Verify PATH
    verify_path "$INSTALL_DIR"

    # Step 10: Verify installation
    verify_installation "$INSTALL_DIR"

    # Done
    print_success "$INSTALL_DIR"
}

main "$@"
