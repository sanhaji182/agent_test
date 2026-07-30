package javascript

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser implements the Parser interface for JavaScript/Node.js
type Parser struct{}

// NewParser creates a new JavaScript parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse analyzes a JavaScript codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "javascript",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// TODO: Implement full parsing logic with tree-sitter
	// For now, return empty codebase structure
	// This will be implemented in Sprint 3-4

	return codebase, nil
}

// SupportedLanguages returns the languages this parser supports
func (p *Parser) SupportedLanguages() []string {
	return []string{"javascript", "typescript"}
}

// DetectFramework detects the JavaScript framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check package.json
	packagePath := filepath.Join(rootDir, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read package.json: %w", err)
	}

	contentStr := string(content)

	// Check for common frameworks
	if contains(contentStr, "express") {
		return "express", nil
	}
	if contains(contentStr, "next") {
		return "nextjs", nil
	}
	if contains(contentStr, "nuxt") {
		return "nuxtjs", nil
	}
	if contains(contentStr, "react") {
		return "react", nil
	}
	if contains(contentStr, "vue") {
		return "vue", nil
	}
	if contains(contentStr, "angular") {
		return "angular", nil
	}
	if contains(contentStr, "fastify") {
		return "fastify", nil
	}
	if contains(contentStr, "hapi") {
		return "hapi", nil
	}

	return "nodejs", nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
