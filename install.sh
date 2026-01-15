#!/bin/bash

# Installation script for AI Commit Message Generator
# Downloads latest release from GitHub and installs
# Supports Mac (ARM64/Intel) and Linux (64-bit/ARM64)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repository
GITHUB_REPO="AllaySahoo222/AI-Commit-Message-Generator"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"

echo "AI Commit Message Generator - Installation Script"
echo "=================================================="
echo ""

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

echo "Detected OS: $OS"
echo "Detected Architecture: $ARCH"
echo ""

# Determine zip file name based on GoReleaser naming convention
ZIP_NAME=""
if [[ "$OS" == "Darwin" ]]; then
    if [[ "$ARCH" == "arm64" ]]; then
        ZIP_NAME="generate-commit-mac-arm64.zip"
        echo "Using Mac ARM64 binary"
    else
        ZIP_NAME="generate-commit-mac-intel.zip"
        echo "Using Mac Intel binary"
    fi
elif [[ "$OS" == "Linux" ]]; then
    if [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
        ZIP_NAME="generate-commit-linux-arm64.zip"
        echo "Using Linux ARM64 binary"
    else
        ZIP_NAME="generate-commit-linux.zip"
        echo "Using Linux binary"
    fi
else
    echo -e "${RED}Error: Unsupported OS: $OS${NC}"
    exit 1
fi

# Get latest release tag
echo -e "${BLUE}Fetching latest release...${NC}"
LATEST_TAG=$(curl -s "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "$LATEST_TAG" ]]; then
    echo -e "${RED}Error: Failed to fetch latest release tag${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: ${LATEST_TAG}${NC}"
echo ""

# Construct download URL
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_TAG}/${ZIP_NAME}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download the release
echo -e "${BLUE}Downloading ${ZIP_NAME}...${NC}"
if ! curl -L -o "${TEMP_DIR}/${ZIP_NAME}" "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download ${ZIP_NAME}${NC}"
    echo "URL: $DOWNLOAD_URL"
    exit 1
fi

# Extract the zip
echo -e "${BLUE}Extracting...${NC}"
cd "$TEMP_DIR"
if ! unzip -q "${ZIP_NAME}"; then
    echo -e "${RED}Error: Failed to extract ${ZIP_NAME}${NC}"
    exit 1
fi

# Find the binary (should be generate-commit)
BINARY_NAME="generate-commit"
if [[ ! -f "$BINARY_NAME" ]]; then
    echo -e "${RED}Error: Binary not found in archive${NC}"
    exit 1
fi

chmod +x "$BINARY_NAME"

# Installation directory
INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="$INSTALL_DIR/generate-commit"

# Check if already installed
if [[ -f "$INSTALL_PATH" ]]; then
    echo -e "${YELLOW}generate-commit is already installed at $INSTALL_PATH${NC}"
    read -p "Overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Install binary
echo -e "${BLUE}Installing to $INSTALL_PATH...${NC}"
sudo cp "$BINARY_NAME" "$INSTALL_PATH"
sudo chmod +x "$INSTALL_PATH"
echo -e "${GREEN}✓ Installed successfully!${NC}"

# Verify installation
if command -v generate-commit &> /dev/null; then
    echo ""
    echo -e "${GREEN}Installation complete!${NC}"
    echo ""
    echo "You can now use 'generate-commit' from anywhere."
    echo ""
    echo "Next steps:"
    echo "1. Set your OLLAMA_API_KEY:"
    echo "   export OLLAMA_API_KEY=your_api_key"
    echo "   (Add this to your ~/.bashrc or ~/.zshrc for persistence)"
    echo ""
    echo "2. Navigate to a git repository and run:"
    echo "   generate-commit init"
    echo ""
    echo "3. Start committing with AI-generated messages!"
else
    echo -e "${YELLOW}Warning: Installation may have failed. Please check manually.${NC}"
    exit 1
fi
