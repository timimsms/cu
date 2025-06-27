#!/usr/bin/env bash
# ClickUp CLI Installation Script
# 
# This script installs the ClickUp CLI (cu) on Unix-like systems (macOS, Linux)
# Usage: curl -sSL https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.sh | bash
#
# You can also specify a version:
# curl -sSL https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.sh | bash -s -- --version v1.0.0
#
# Or install to a custom directory:
# curl -sSL https://raw.githubusercontent.com/timimsms/cu/main/scripts/install.sh | bash -s -- --dir /usr/local/bin

set -euo pipefail

# Configuration
REPO_OWNER="timimsms"
REPO_NAME="cu"
BINARY_NAME="cu"
INSTALL_DIR="${HOME}/.local/bin"
VERSION="latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

success() {
    echo -e "${GREEN}$1${NC}"
}

info() {
    echo -e "${BLUE}$1${NC}"
}

warning() {
    echo -e "${YELLOW}$1${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "ClickUp CLI Installation Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION    Install specific version (default: latest)"
            echo "  --dir DIRECTORY      Install to specific directory (default: ~/.local/bin)"
            echo "  --help               Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Darwin)
            os="darwin"
            ;;
        Linux)
            os="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            error "Windows detected. Please use install.ps1 instead."
            exit 1
            ;;
        *)
            error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="x86_64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        i386|i686)
            arch="i386"
            ;;
        armv7l|armv7)
            arch="armv7"
            ;;
        *)
            error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

# Get the latest version from GitHub
get_latest_version() {
    local latest_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Download file
download_file() {
    local url="$1"
    local output="$2"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksums_url="$2"
    local expected_checksum
    
    info "Verifying checksum..."
    
    # Download checksums file
    download_file "$checksums_url" "checksums.txt"
    
    # Extract expected checksum for our file
    expected_checksum=$(grep "$(basename "$file")" checksums.txt | cut -d' ' -f1)
    
    if [ -z "$expected_checksum" ]; then
        warning "Could not find checksum for $(basename "$file"), skipping verification"
        rm -f checksums.txt
        return 0
    fi
    
    # Calculate actual checksum
    local actual_checksum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        actual_checksum=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        warning "No SHA256 tool found, skipping checksum verification"
        rm -f checksums.txt
        return 0
    fi
    
    # Compare checksums
    if [ "$expected_checksum" != "$actual_checksum" ]; then
        error "Checksum verification failed!"
        error "Expected: $expected_checksum"
        error "Actual:   $actual_checksum"
        rm -f checksums.txt
        exit 1
    fi
    
    success "Checksum verified ✓"
    rm -f checksums.txt
}

# Main installation function
main() {
    info "ClickUp CLI Installer"
    info "===================="
    echo ""
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    info "Detected platform: $platform"
    
    # Get version to install
    if [ "$VERSION" == "latest" ]; then
        info "Fetching latest version..."
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            error "Failed to get latest version"
            exit 1
        fi
    fi
    info "Installing version: $VERSION"
    
    # Construct download URL
    local archive_name="${BINARY_NAME}_${platform}.tar.gz"
    if [[ "$platform" == *"windows"* ]]; then
        archive_name="${BINARY_NAME}_${platform}.zip"
    fi
    
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${archive_name}"
    local checksums_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/checksums.txt"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    cd "$temp_dir"
    
    # Download archive
    info "Downloading $BINARY_NAME..."
    download_file "$download_url" "$archive_name"
    
    # Verify checksum
    verify_checksum "$archive_name" "$checksums_url"
    
    # Extract archive
    info "Extracting archive..."
    if [[ "$archive_name" == *.tar.gz ]]; then
        tar -xzf "$archive_name"
    elif [[ "$archive_name" == *.zip ]]; then
        unzip -q "$archive_name"
    fi
    
    # Find the binary
    if [ ! -f "$BINARY_NAME" ]; then
        error "Binary $BINARY_NAME not found in archive"
        exit 1
    fi
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        info "Creating directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    # Install binary
    info "Installing $BINARY_NAME to $INSTALL_DIR..."
    chmod +x "$BINARY_NAME"
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    
    # Verify installation
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        success "Installation successful! ✓"
        echo ""
        
        # Check if install directory is in PATH
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            warning "Note: $INSTALL_DIR is not in your PATH"
            echo ""
            echo "Add the following to your shell configuration file:"
            echo ""
            case "$SHELL" in
                */bash)
                    echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
                    echo "  source ~/.bashrc"
                    ;;
                */zsh)
                    echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
                    echo "  source ~/.zshrc"
                    ;;
                */fish)
                    echo "  echo 'set -gx PATH \$PATH $INSTALL_DIR' >> ~/.config/fish/config.fish"
                    echo "  source ~/.config/fish/config.fish"
                    ;;
                *)
                    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
                    ;;
            esac
            echo ""
        fi
        
        # Show version
        if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]] || [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
            info "Installed version:"
            "$INSTALL_DIR/$BINARY_NAME" --version || true
        fi
        
        echo ""
        info "Get started with:"
        echo "  $BINARY_NAME --help"
        echo "  $BINARY_NAME auth login"
        
    else
        error "Installation failed"
        exit 1
    fi
}

# Run main function
main