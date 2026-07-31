#!/bin/bash
#
# GoTest Agent - Tree-sitter Dependencies Installer
# This script automatically downloads and installs all tree-sitter language parsers
# required by GoTest Agent's multi-language parser system.
#
# Usage:
#   ./scripts/install-tree-sitter-deps.sh
#
# Supported Languages:
#   - JavaScript (JavaScript + TypeScript)
#   - Go
#   - Python
#   - PHP
#

set -e  # Exit on error

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  GoTest Agent - Tree-sitter Dependencies Installer            ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✓${NC} Go is installed: $(go version)"
echo ""

# Install tree-sitter core
echo "Installing tree-sitter core..."
go get github.com/smacker/go-tree-sitter@v0.0.0-20240827094217-dd81d9e9be82
echo -e "${GREEN}✓${NC} tree-sitter core installed"
echo ""

# Install language parsers
echo "Installing language parsers..."
echo ""

# JavaScript/TypeScript
echo "  Installing JavaScript/TypeScript parser..."
go get github.com/smacker/go-tree-sitter/javascript@v0.0.0-20240827094217-dd81d9e9be82
echo -e "${GREEN}  ✓${NC} JavaScript/TypeScript parser installed"

# Go
echo "  Installing Go parser..."
go get github.com/smacker/go-tree-sitter/golang@v0.0.0-20240827094217-dd81d9e9be82
echo -e "${GREEN}  ✓${NC} Go parser installed"

# Python
echo "  Installing Python parser..."
go get github.com/smacker/go-tree-sitter/python@v0.0.0-20240827094217-dd81d9e9be82
echo -e "${GREEN}  ✓${NC} Python parser installed"

# PHP
echo "  Installing PHP parser..."
go get github.com/smacker/go-tree-sitter/php@v0.0.0-20240827094217-dd81d9e9be82
echo -e "${GREEN}  ✓${NC} PHP parser installed"

echo ""

# Run go mod tidy to clean up
echo "Running go mod tidy..."
go mod tidy
echo -e "${GREEN}✓${NC} go.mod and go.sum updated"
echo ""

# Verify installation
echo "Verifying installation..."
echo ""

# Check if parsers are available
echo "Checking parser availability..."
echo ""

# Create a simple test program
cat > /tmp/test_parsers.go << 'EOF'
package main

import (
	"fmt"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/php"
)

func main() {
	parsers := map[string]*sitter.Language{
		"JavaScript": javascript.GetLanguage(),
		"Go":         golang.GetLanguage(),
		"Python":     python.GetLanguage(),
		"PHP":        php.GetLanguage(),
	}

	for name, lang := range parsers {
		if lang != nil {
			fmt.Printf("  ✓ %s parser available\n", name)
		} else {
			fmt.Printf("  ✗ %s parser NOT available\n", name)
		}
	}
}
EOF

# Run the test
go run /tmp/test_parsers.go

# Clean up
rm /tmp/test_parsers.go

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  Installation Complete!                                        ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  1. Run tests: go test ./internal/parser/..."
echo "  2. Start the server: go run ./cmd/server"
echo ""
echo "For more information, see:"
echo "  - PARSERS.md - Parser implementation details"
echo "  - PHASE-2-IMPLEMENTATION.md - Phase 2 implementation guide"
echo "  - PHASE-3-IMPLEMENTATION.md - Phase 3 implementation guide"
echo "  - PHASE-4-IMPLEMENTATION.md - Phase 4 implementation guide"
echo ""
