package parser

import (
	"context"
	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser adalah interface untuk parser bahasa pemrograman
// Setiap bahasa (JavaScript, Go, PHP, Python) harus implement interface ini
type Parser interface {
	// Parse menganalisis codebase dari root directory
	// Mengembalikan Codebase yang berisi routes, models, dan handlers
	Parse(ctx context.Context, rootDir string) (*types.Codebase, error)

	// SupportedLanguages mengembalikan list bahasa yang di-support parser ini
	SupportedLanguages() []string

	// DetectFramework mendeteksi framework dari struktur project
	DetectFramework(rootDir string) (string, error)
}
