#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${PROJECT_ROOT}/dist"
BUNDLE_DIR="${DIST_DIR}/tally-mcp-bundle"
SERVER_DIR="${BUNDLE_DIR}/server"
OUTPUT_FILE="${DIST_DIR}/tally-mcp-0.1.0.mcpb"
INTEGRATIONS_DIR="${PROJECT_ROOT}/integrations/claude/extension"

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Main build function
main() {
    log_info "Building Tally MCP Claude Extension..."

    # Step 1: Create directory structure
    log_info "Creating directory structure..."
    mkdir -p "${SERVER_DIR}"
    log_success "Directory structure created"

    # Step 2: Build binaries for all platforms
    log_info "Building binaries for all platforms..."
    cd "${PROJECT_ROOT}"

    log_info "  Building macOS ARM64..."
    GOOS=darwin GOARCH=arm64 go build -o "${SERVER_DIR}/tally-mcp-mac" .
    log_success "macOS ARM64 built"

    log_info "  Building macOS Intel..."
    GOOS=darwin GOARCH=amd64 go build -o "${SERVER_DIR}/tally-mcp-mac-x86" .
    log_success "macOS Intel built"

    log_info "  Building Linux..."
    GOOS=linux GOARCH=amd64 go build -o "${SERVER_DIR}/tally-mcp-linux" .
    log_success "Linux built"

    log_info "  Building Windows..."
    GOOS=windows GOARCH=amd64 go build -o "${SERVER_DIR}/tally-mcp.exe" .
    log_success "Windows built"

    # Step 3: Copy manifest and icon
    log_info "Copying extension metadata..."
    if [ ! -f "${INTEGRATIONS_DIR}/manifest.json" ]; then
        log_error "manifest.json not found at ${INTEGRATIONS_DIR}/manifest.json"
        exit 1
    fi
    if [ ! -f "${INTEGRATIONS_DIR}/icon.png" ]; then
        log_error "icon.png not found at ${INTEGRATIONS_DIR}/icon.png"
        exit 1
    fi

    cp "${INTEGRATIONS_DIR}/manifest.json" "${BUNDLE_DIR}/"
    cp "${INTEGRATIONS_DIR}/icon.png" "${BUNDLE_DIR}/"
    log_success "Manifest and icon copied"

    # Step 4: Copy wrapper script
    log_info "Copying wrapper script..."
    if [ ! -f "${INTEGRATIONS_DIR}/tally-mcp.sh" ]; then
        log_error "Wrapper script not found at ${INTEGRATIONS_DIR}/tally-mcp.sh"
        exit 1
    fi

    cp "${INTEGRATIONS_DIR}/tally-mcp.sh" "${SERVER_DIR}/tally-mcp"
    chmod +x "${SERVER_DIR}/tally-mcp"
    log_success "Wrapper script copied and made executable"

    # Step 5: Copy tools directory
    log_info "Copying tools directory..."
    if [ ! -d "${PROJECT_ROOT}/tools" ]; then
        log_error "Tools directory not found at ${PROJECT_ROOT}/tools"
        exit 1
    fi

    rm -rf "${SERVER_DIR}/tools"
    cp -r "${PROJECT_ROOT}/tools" "${SERVER_DIR}/tools"
    log_success "Tools directory copied"

    # Step 7: Package as .mcpb
    log_info "Packaging as Claude extension..."
    npx @anthropic-ai/mcpb pack "${BUNDLE_DIR}" "${OUTPUT_FILE}"

    # Step 8: Verify output
    if [ ! -f "${OUTPUT_FILE}" ]; then
        log_error "Failed to create ${OUTPUT_FILE}"
        exit 1
    fi

    local size=$(ls -lh "${OUTPUT_FILE}" | awk '{print $5}')
    local shasum=$(shasum -a 256 "${OUTPUT_FILE}" | awk '{print $1}')

    log_success "Extension packaged successfully"

    # Summary
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ BUILD COMPLETE${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Extension Details:"
    echo "  File: ${OUTPUT_FILE}"
    echo "  Size: ${size}"
    echo "  SHA256: ${shasum}"
    echo ""
    echo "Installation Instructions:"
    echo "  1. Open Claude Desktop"
    echo "  2. Settings → Customization → Connectors"
    echo "  3. Click 'Add Custom Connector'"
    echo "  4. Select the .mcpb file"
    echo "  5. Configure Tally connection details"
    echo "  6. Restart Claude Desktop"
    echo ""
    echo "Configuration Required:"
    echo "  • Tally Server Host (IP or hostname)"
    echo "  • Tally Server Port (default: 9000)"
    echo "  • Company Name (as it appears in Tally)"
    echo ""
    echo "Before installing, ensure:"
    echo "  • Tally's XML API is enabled"
    echo "  • Sales and Journal vouchers are set to 'Automatic' numbering"
    echo "  • See README.md for detailed Tally setup instructions"
    echo ""
}

# Run main
main "$@"
