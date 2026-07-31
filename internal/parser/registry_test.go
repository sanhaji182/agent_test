package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser"
	golang "github.com/go-go-golems/gotest-agent/internal/parser/go"
	"github.com/go-go-golems/gotest-agent/internal/parser/javascript"
	"github.com/go-go-golems/gotest-agent/internal/parser/php"
	"github.com/go-go-golems/gotest-agent/internal/parser/python"
	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestNewRegistry(t *testing.T) {
	registry := parser.NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
}

func TestRegisterParser(t *testing.T) {
	registry := parser.NewRegistry()
	jsParser := javascript.NewParser()

	registry.Register("javascript", jsParser)

	retrieved, err := registry.GetParser("javascript")
	if err != nil {
		t.Fatalf("Failed to retrieve registered parser: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved parser is nil")
	}
}

func TestGetParserNotExists(t *testing.T) {
	registry := parser.NewRegistry()

	_, err := registry.GetParser("nonexistent")
	if err == nil {
		t.Fatal("Should return error for non-existent parser")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "JavaScript project",
			files: map[string]string{
				"package.json": `{"name": "test-project"}`,
				"index.js":     "console.log('hello')",
			},
			expected: "javascript",
		},
		{
			name: "Go project",
			files: map[string]string{
				"go.mod":  "module test",
				"main.go": "package main",
			},
			expected: "go",
		},
		{
			name: "PHP project",
			files: map[string]string{
				"composer.json": `{"name": "test/project"}`,
				"index.php":     "<?php echo 'hello'; ?>",
			},
			expected: "php",
		},
		{
			name: "Python project",
			files: map[string]string{
				"requirements.txt": "flask==2.0.0",
				"app.py":           "from flask import Flask",
			},
			expected: "python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "parser-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test files
			for filename, content := range tt.files {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", filename, err)
				}
			}

			// Test language detection
			registry := parser.NewRegistry()
			language, err := registry.DetectLanguage(tmpDir)
			if err != nil {
				t.Fatalf("DetectLanguage failed: %v", err)
			}
			if language != tt.expected {
				t.Errorf("Expected language %s, got %s", tt.expected, language)
			}
		})
	}
}

func TestParseCodebase(t *testing.T) {
	// Create temp directory with JavaScript project
	tmpDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"name": "test-project", "dependencies": {"express": "^4.18.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Setup registry
	registry := parser.NewRegistry()
	registry.Register("javascript", javascript.NewParser())

	// Parse codebase
	codebase, err := registry.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if codebase == nil {
		t.Fatal("Parse returned nil")
	}
	if codebase.Language != "javascript" {
		t.Errorf("Expected language javascript, got %s", codebase.Language)
	}
}

func TestAllParsers(t *testing.T) {
	parsers := map[string]parser.Parser{
		"javascript": javascript.NewParser(),
		"go":         golang.NewParser(),
		"php":        php.NewParser(),
		"python":     python.NewParser(),
	}

	for name, p := range parsers {
		t.Run(name, func(t *testing.T) {
			if p == nil {
				t.Fatalf("%s parser is nil", name)
			}
		})
	}
}

func TestSupportedLanguages(t *testing.T) {
	registry := parser.NewRegistry()
	registry.Register("javascript", javascript.NewParser())
	registry.Register("go", golang.NewParser())

	languages := registry.SupportedLanguages()
	if len(languages) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(languages))
	}

	// Check if both languages are present
	hasJS := false
	hasGo := false
	for _, lang := range languages {
		if lang == "javascript" {
			hasJS = true
		}
		if lang == "go" {
			hasGo = true
		}
	}

	if !hasJS {
		t.Error("Expected javascript in supported languages")
	}
	if !hasGo {
		t.Error("Expected go in supported languages")
	}
}

func TestParseWithAutoDetect(t *testing.T) {
	// Create temp directory with Go project
	tmpDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := `module test

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Setup registry with all parsers
	registry := parser.NewRegistry()
	registry.Register("javascript", javascript.NewParser())
	registry.Register("go", golang.NewParser())
	registry.Register("php", php.NewParser())
	registry.Register("python", python.NewParser())

	// Parse codebase (should auto-detect as Go)
	codebase, err := registry.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if codebase.Language != "go" {
		t.Errorf("Expected language go, got %s", codebase.Language)
	}
}

func TestDetectLanguageByFileExtension(t *testing.T) {
	// Create temp directory with only Python files (no requirements.txt)
	tmpDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Python files
	pyFiles := map[string]string{
		"app.py":    "from flask import Flask",
		"models.py": "class User:",
		"utils.py":  "def helper():",
	}

	for filename, content := range pyFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Test language detection
	registry := parser.NewRegistry()
	language, err := registry.DetectLanguage(tmpDir)
	if err != nil {
		t.Fatalf("DetectLanguage failed: %v", err)
	}
	if language != "python" {
		t.Errorf("Expected language python, got %s", language)
	}
}

// TestParseEmptyDirectory tests parsing an empty directory
func TestParseEmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := parser.NewRegistry()
	registry.Register("javascript", javascript.NewParser())

	// Should fail to detect language
	_, err = registry.Parse(context.Background(), tmpDir)
	if err == nil {
		t.Fatal("Expected error when parsing empty directory")
	}
}

// TestCodebaseType ensures the Codebase type is properly defined
func TestCodebaseType(t *testing.T) {
	codebase := &types.Codebase{
		Language:  "javascript",
		Framework: "express",
	}

	if codebase.Language != "javascript" {
		t.Errorf("Expected language javascript, got %s", codebase.Language)
	}
	if codebase.Framework != "express" {
		t.Errorf("Expected framework express, got %s", codebase.Framework)
	}
}
