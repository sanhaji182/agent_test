package agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/gotest-agent/internal/steel"
)

// SteelScreenshotter mengambil screenshot melalui Steel Browser API
type SteelScreenshotter struct {
	client    *steel.Client
	outputDir string // Direktori untuk menyimpan file screenshot
	sessionID string // ID sesi Steel yang aktif
}

// NewSteelScreenshotter membuat screenshotter baru dengan Steel client
func NewSteelScreenshotter(client *steel.Client, outputDir string) *SteelScreenshotter {
	return &SteelScreenshotter{client: client, outputDir: outputDir}
}

// SetSession mengatur sesi Steel yang aktif untuk capture screenshot
func (s *SteelScreenshotter) SetSession(sessionID string) {
	s.sessionID = sessionID
}

// Capture mengambil screenshot dari sesi Steel dan menyimpannya ke disk
// Mengembalikan URL relatif ke file screenshot
func (s *SteelScreenshotter) Capture(ctx context.Context, runID string, label string) (string, error) {
	if s.sessionID == "" {
		return "", fmt.Errorf("no active steel session")
	}

	// Ambil screenshot dari Steel Browser
	data, err := s.client.Screenshot(ctx, s.sessionID, true)
	if err != nil {
		return "", fmt.Errorf("capture screenshot: %w", err)
	}

	// Buat direktori output per run
	dir := filepath.Join(s.outputDir, runID)
	os.MkdirAll(dir, 0755)

	filename := label + ".png"
	path := filepath.Join(dir, filename)

	// Jika data dalam format base64, decode dulu
	if decoded, err := base64.StdEncoding.DecodeString(string(data)); err == nil && len(decoded) > 0 {
		data = decoded
	}

	// Simpan ke file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write screenshot: %w", err)
	}

	return fmt.Sprintf("/screenshots/%s/%s", runID, filename), nil
}
