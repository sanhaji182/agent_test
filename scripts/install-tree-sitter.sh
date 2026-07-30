#!/bin/bash

# Install tree-sitter dependencies for all parsers

echo "Installing tree-sitter dependencies..."

# Core tree-sitter library
go get github.com/smacker/go-tree-sitter

# Language grammars
go get github.com/smacker/go-tree-sitter/javascript
go get github.com/smacker/go-tree-sitter/golang
go get github.com/smacker/go-tree-sitter/php
go get github.com/smacker/go-tree-sitter/python

# Tidy up
go mod tidy

echo "✅ Tree-sitter dependencies installed successfully!"
echo ""
echo "You can now run parser tests:"
echo "  go test ./internal/parser/... -v"
