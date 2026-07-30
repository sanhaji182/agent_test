package php

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser implements the Parser interface for PHP
type Parser struct{}

// NewParser creates a new PHP parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse analyzes a PHP codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "php",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// TODO: Implement full parsing logic with tree-sitter
	// For now, return empty codebase structure
	// This will be implemented in Sprint 6

	return codebase, nil
}

// SupportedLanguages returns the languages this parser supports
func (p *Parser) SupportedLanguages() []string {
	return []string{"php"}
}

// DetectFramework detects the PHP framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check composer.json
	composerPath := filepath.Join(rootDir, "composer.json")
	content, err := os.ReadFile(composerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read composer.json: %w", err)
	}

	contentStr := string(content)

	// Check for common frameworks
	if strings.Contains(contentStr, "laravel/framework") {
		return "laravel", nil
	}
	if strings.Contains(contentStr, "symfony/symfony") || strings.Contains(contentStr, "symfony/framework-bundle") {
		return "symfony", nil
	}
	if strings.Contains(contentStr, "slim/slim") {
		return "slim", nil
	}
	if strings.Contains(contentStr, "cakephp/cakephp") {
		return "cakephp", nil
	}

	return "php", nil
}
