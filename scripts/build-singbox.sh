#!/bin/bash
#
# Build sing-box binaries for all target platforms
# Usage: ./scripts/build-singbox.sh [path-to-sing-box-source]
#

set -e

# Configuration
SINGBOX_SOURCE="${1:-../sing-box-new}"
OUTPUT_DIR="internal/binary/embedded"
TAGS="with_gvisor,with_quic,with_dhcp,with_wireguard,with_utls,with_acme,with_clash_api,with_tailscale,with_ccm,with_ocm,badlinkname,tfogo_checklinkname0"

# Target platforms (GOOS/GOARCH pairs)
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Check sing-box source exists
if [ ! -d "$SINGBOX_SOURCE" ]; then
    error "sing-box source not found at: $SINGBOX_SOURCE"
fi

if [ ! -f "$SINGBOX_SOURCE/go.mod" ]; then
    error "Invalid sing-box source (no go.mod): $SINGBOX_SOURCE"
fi

# Get version from sing-box
cd "$SINGBOX_SOURCE"
SINGBOX_SOURCE=$(pwd)  # Get absolute path

# Try to get version
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

info "Building sing-box version: $VERSION (commit: $COMMIT)"
info "Source: $SINGBOX_SOURCE"

# Return to nethopper directory
cd - > /dev/null

# Create output directory
mkdir -p "$OUTPUT_DIR"
info "Output directory: $OUTPUT_DIR"

# Build parameters
LDFLAGS="-X 'github.com/sagernet/sing-box/constant.Version=${VERSION}' -s -w -buildid="
BUILD_FLAGS="-trimpath -ldflags \"$LDFLAGS\" -tags \"$TAGS\""

echo ""
info "Starting cross-compilation for ${#PLATFORMS[@]} platforms..."
echo ""

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"

    # Output filename
    OUTPUT_NAME="sing-box-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"

    info "Building for ${GOOS}/${GOARCH}..."

    # Build
    cd "$SINGBOX_SOURCE"

    if GOTOOLCHAIN=local CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
        go build -trimpath \
        -ldflags "-X 'github.com/sagernet/sing-box/constant.Version=${VERSION}' -s -w -buildid= -checklinkname=0" \
        -tags "$TAGS" \
        -o "${OLDPWD}/${OUTPUT_PATH}" \
        ./cmd/sing-box 2>&1; then

        cd - > /dev/null
        SIZE=$(ls -lh "$OUTPUT_PATH" | awk '{print $5}')
        echo -e "  ${GREEN}✓${NC} ${OUTPUT_NAME} (${SIZE})"
    else
        cd - > /dev/null
        echo -e "  ${RED}✗${NC} Failed to build for ${GOOS}/${GOARCH}"
    fi
done

echo ""
info "Build complete! Binaries in: $OUTPUT_DIR"
echo ""

# List results
ls -lh "$OUTPUT_DIR"/ | grep -v ".gitkeep" | tail -n +2

echo ""
info "You can now build nethopper with: go build ./cmd/nethopper"
