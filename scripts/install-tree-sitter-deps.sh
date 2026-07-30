#!/bin/bash
# Install tree-sitter dependencies for all parsers

set -e

echo "Installing tree-sitter dependencies..."

# Core tree-sitter library
echo "Installing github.com/smacker/go-tree-sitter..."
go get github.com/smacker/go-tree-sitter@latest

# Language-specific parsers
echo "Installing JavaScript parser..."
go get github.com/smacker/go-tree-sitter/javascript@latest

echo "Installing Go parser..."
go get github.com/smacker/go-tree-sitter/golang@latest

echo "Installing PHP parser..."
go get github.com/smacker/go-tree-sitter/php@latest

echo "Installing Python parser..."
go get github.com/smacker/go-tree-sitter/python@latest

# Clean up go.mod
echo "Running go mod tidy..."
go mod tidy

echo "✓ All tree-sitter dependencies installed successfully!"
echo ""
echo "You can now run parser tests:"
echo "  go test ./internal/parser/... -v"
