package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	golang "github.com/go-go-golems/gotest-agent/internal/parser/go"
	"github.com/go-go-golems/gotest-agent/internal/parser/javascript"
	"github.com/go-go-golems/gotest-agent/internal/parser/php"
	"github.com/go-go-golems/gotest-agent/internal/parser/python"
	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Registry mengelola semua parser yang tersedia
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry membuat registry kosong; gunakan Register untuk menambahkan parser
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

// NewDefaultRegistry membuat registry dengan semua parser default terdaftar
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("javascript", javascript.NewParser())
	r.Register("go", golang.NewParser())
	r.Register("php", php.NewParser())
	r.Register("python", python.NewParser())
	return r
}

// Register menambahkan parser ke registry
func (r *Registry) Register(language string, parser Parser) {
	r.parsers[language] = parser
}

// GetParser mengambil parser untuk bahasa tertentu
func (r *Registry) GetParser(language string) (Parser, error) {
	parser, ok := r.parsers[language]
	if !ok {
		return nil, fmt.Errorf("no parser available for language: %s", language)
	}
	return parser, nil
}

// DetectLanguage mendeteksi bahasa pemrograman dari struktur project
func (r *Registry) DetectLanguage(rootDir string) (string, error) {
	// Check package.json -> JavaScript/Node.js
	if _, err := os.Stat(filepath.Join(rootDir, "package.json")); err == nil {
		return "javascript", nil
	}

	// Check go.mod -> Go
	if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
		return "go", nil
	}

	// Check composer.json -> PHP
	if _, err := os.Stat(filepath.Join(rootDir, "composer.json")); err == nil {
		return "php", nil
	}

	// Check requirements.txt -> Python
	if _, err := os.Stat(filepath.Join(rootDir, "requirements.txt")); err == nil {
		return "python", nil
	}

	// Check pyproject.toml -> Python
	if _, err := os.Stat(filepath.Join(rootDir, "pyproject.toml")); err == nil {
		return "python", nil
	}

	// Check setup.py -> Python
	if _, err := os.Stat(filepath.Join(rootDir, "setup.py")); err == nil {
		return "python", nil
	}

	// Fallback: scan file extensions
	files, err := os.ReadDir(rootDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	langCount := make(map[string]int)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		switch ext {
		case ".js", ".jsx", ".ts", ".tsx":
			langCount["javascript"]++
		case ".go":
			langCount["go"]++
		case ".php":
			langCount["php"]++
		case ".py":
			langCount["python"]++
		}
	}

	// Return language with most files
	var maxLang string
	var maxCount int
	for lang, count := range langCount {
		if count > maxCount {
			maxLang = lang
			maxCount = count
		}
	}

	if maxLang != "" {
		return maxLang, nil
	}

	return "", fmt.Errorf("could not detect programming language")
}

// Parse menganalisis codebase dengan auto-detect bahasa
func (r *Registry) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	// Detect language
	language, err := r.DetectLanguage(rootDir)
	if err != nil {
		return nil, fmt.Errorf("language detection failed: %w", err)
	}

	// Get parser
	parser, err := r.GetParser(language)
	if err != nil {
		return nil, err
	}

	// Parse codebase
	return parser.Parse(ctx, rootDir)
}

// SupportedLanguages mengembalikan list semua bahasa yang di-support
func (r *Registry) SupportedLanguages() []string {
	languages := make([]string, 0, len(r.parsers))
	for lang := range r.parsers {
		languages = append(languages, lang)
	}
	return languages
}
