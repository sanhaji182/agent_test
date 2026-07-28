#!/bin/bash
# scripts/install-playwright-driver.sh
# Installs Playwright driver manually when CDN is unavailable.
#
# Usage:
#   1. Download playwright-1.57.0-mac-arm64.zip from a working source
#   2. Place it in the current directory or specify path as $1
#   3. Run: bash scripts/install-playwright-driver.sh [path-to-zip]
#
# Sources to try:
#   - https://www.npmjs.com/package/playwright-core/v/1.57.0 (npm package)
#   - https://github.com/Automattic/playwright-builds (check releases/assets)
#   - Mirror your own: host the zip on S3/Cloudflare/GitHub releases

set -euo pipefail

VERSION="1.57.0"
PLATFORM=""
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    if [ "$ARCH" = "arm64" ]; then
      PLATFORM="mac-arm64"
    else
      PLATFORM="mac"
    fi
    CACHE_DIR="$HOME/Library/Caches"
    ;;
  Linux)
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
      PLATFORM="linux-arm64"
    else
      PLATFORM="linux"
    fi
    CACHE_DIR="$HOME/.cache"
    ;;
  MINGW*|MSYS*|CYGWIN*)
    PLATFORM="win32_x64"
    CACHE_DIR="$HOME/AppData/Local"
    ;;
  *)
    echo "ERROR: Unsupported OS: $OS"
    exit 1
    ;;
esac

FILENAME="playwright-${VERSION}-${PLATFORM}.zip"
TARGET_DIR="${CACHE_DIR}/ms-playwright-go/${VERSION}"

echo "Platform:  $PLATFORM"
echo "Target:    $TARGET_DIR"

ZIP_PATH="${1:-}"

if [ -z "$ZIP_PATH" ]; then
  echo ""
  echo "No zip path provided. Looking for ./$FILENAME ..."
  if [ -f "./$FILENAME" ]; then
    ZIP_PATH="./$FILENAME"
  else
    echo ""
    echo "ERROR: $FILENAME not found in current directory."
    echo ""
    echo "Download the driver from one of these sources:"
    echo "  1. npm: npm pack playwright-core@${VERSION}"
    echo "     (then extract .local-browsers/ from the .tgz)"
    echo "  2. GitHub: check https://github.com/Automattic/playwright-builds"
    echo "  3. Self-host: upload the zip to your CDN and set PLAYWRIGHT_DOWNLOAD_HOST"
    echo ""
    echo "Then re-run: bash $0 /path/to/$FILENAME"
    exit 1
  fi
fi

if [ ! -f "$ZIP_PATH" ]; then
  echo "ERROR: File not found: $ZIP_PATH"
  exit 1
fi

echo "Source:    $ZIP_PATH"
echo ""

# Create target directory
mkdir -p "$TARGET_DIR"

# Extract
echo "Extracting to $TARGET_DIR ..."
unzip -o "$ZIP_PATH" -d "$TARGET_DIR"

# Make node executable on Unix
if [ "$OS" != "MINGW" ] && [ -f "$TARGET_DIR/node" ]; then
  chmod +x "$TARGET_DIR/node"
fi

echo ""
echo "Done. Verify with:"
echo "  ls -la $TARGET_DIR"
echo ""
echo "The driver should contain: node (or node.exe), package/cli.js"

# Verify
if [ -f "$TARGET_DIR/package/cli.js" ]; then
  echo ""
  echo "✅ Driver installed successfully"
  echo "   Version: $($TARGET_DIR/node $TARGET_DIR/package/cli.js --version 2>/dev/null || echo 'unknown')"
else
  echo ""
  echo "⚠️  Warning: package/cli.js not found — zip structure may be wrong"
  echo "   Expected structure: node, package/cli.js"
  echo "   Got:"
  ls -la "$TARGET_DIR" | head -10
fi
