package golang

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser implements the Parser interface for Go
type Parser struct{}

// NewParser creates a new Go parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse analyzes a Go codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "go",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// TODO: Implement full parsing logic with tree-sitter
	// For now, return empty codebase structure
	// This will be implemented in Sprint 5

	return codebase, nil
}

// SupportedLanguages returns the languages this parser supports
func (p *Parser) SupportedLanguages() []string {
	return []string{"go"}
}

// DetectFramework detects the Go framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check go.mod
	modPath := filepath.Join(rootDir, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	contentStr := string(content)

	// Check for common frameworks
	if strings.Contains(contentStr, "github.com/gin-gonic/gin") {
		return "gin", nil
	}
	if strings.Contains(contentStr, "github.com/labstack/echo") {
		return "echo", nil
	}
	if strings.Contains(contentStr, "github.com/gofiber/fiber") {
		return "fiber", nil
	}
	if strings.Contains(contentStr, "github.com/go-chi/chi") {
		return "chi", nil
	}
	if strings.Contains(contentStr, "github.com/gorilla/mux") {
		return "gorilla", nil
	}

	return "stdlib", nil
}
