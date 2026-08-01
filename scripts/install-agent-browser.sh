#!/usr/bin/env bash
# install-agent-browser.sh — Install agent-browser CLI for GoTest Agent
#
# agent-browser is an optional dependency that provides compact accessibility
# tree snapshots for LLM-efficient page analysis. GoTest Agent works without
# it (falling back to Playwright), but having it improves test generation
# token efficiency by ~10x.
#
# Usage:
#   ./scripts/install-agent-browser.sh          # install latest
#   ./scripts/install-agent-browser.sh 1.2.3    # pin specific version
#
set -euo pipefail

VERSION="${1:-latest}"

echo "Installing agent-browser (v${VERSION})..."

# macOS via Homebrew
if command -v brew &>/dev/null; then
    echo "Detected macOS with Homebrew"
    if [ "$VERSION" = "latest" ]; then
        brew install agent-browser 2>/dev/null || brew upgrade agent-browser 2>/dev/null
    else
        brew install agent-browser@$VERSION 2>/dev/null || true
    fi
# npm/npx fallback (cross-platform)
elif command -v npm &>/dev/null || command -v npx &>/dev/null; then
    echo "Using npm install"
    if [ "$VERSION" = "latest" ]; then
        npm install -g agent-browser
    else
        npm install -g "agent-browser@${VERSION}"
    fi
else
    echo "ERROR: Neither brew nor npm found. Install one of them first."
    echo ""
    echo "Alternative: download binary from https://github.com/vercel-labs/agent-browser/releases"
    exit 1
fi

# Download Chrome for agent-browser (first time only)
echo ""
echo "Running agent-browser install (download Chrome)..."
agent-browser install || echo "Warning: agent-browser install failed — Chrome may already be available"

# Verify
echo ""
if command -v agent-browser &>/dev/null; then
    echo "✅ agent-browser installed successfully"
    agent-browser --version
else
    echo "❌ agent-browser not found in PATH after install"
    echo "   You may need to restart your shell or add npm global bin to PATH"
    exit 1
fi
