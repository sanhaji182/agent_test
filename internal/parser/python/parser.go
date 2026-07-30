package python

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser implements the Parser interface for Python
type Parser struct{}

// NewParser creates a new Python parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse analyzes a Python codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "python",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// TODO: Implement full parsing logic with tree-sitter
	// For now, return empty codebase structure
	// This will be implemented in Sprint 7

	return codebase, nil
}

// SupportedLanguages returns the languages this parser supports
func (p *Parser) SupportedLanguages() []string {
	return []string{"python"}
}

// DetectFramework detects the Python framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check requirements.txt
	reqPath := filepath.Join(rootDir, "requirements.txt")
	if content, err := os.ReadFile(reqPath); err == nil {
		framework := p.detectFrameworkFromContent(string(content))
		if framework != "" {
			return framework, nil
		}
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(rootDir, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		framework := p.detectFrameworkFromContent(string(content))
		if framework != "" {
			return framework, nil
		}
	}

	return "python", nil
}

func (p *Parser) detectFrameworkFromContent(content string) string {
	if strings.Contains(content, "fastapi") {
		return "fastapi"
	}
	if strings.Contains(content, "flask") {
		return "flask"
	}
	if strings.Contains(content, "django") {
		return "django"
	}
	if strings.Contains(content, "tornado") {
		return "tornado"
	}
	return ""
}
