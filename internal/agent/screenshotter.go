package agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/gotest-agent/internal/steel"
)

// SteelScreenshotter captures screenshots via Steel Browser sessions.
type SteelScreenshotter struct {
	client    *steel.Client
	outputDir string
	sessionID string
}

func NewSteelScreenshotter(client *steel.Client, outputDir string) *SteelScreenshotter {
	return &SteelScreenshotter{client: client, outputDir: outputDir}
}

// SetSession sets the active Steel session to capture from.
func (s *SteelScreenshotter) SetSession(sessionID string) {
	s.sessionID = sessionID
}

func (s *SteelScreenshotter) Capture(ctx context.Context, runID string, label string) (string, error) {
	if s.sessionID == "" {
		return "", fmt.Errorf("no active steel session")
	}

	data, err := s.client.Screenshot(ctx, s.sessionID, true)
	if err != nil {
		return "", fmt.Errorf("capture screenshot: %w", err)
	}

	// Save to disk
	dir := filepath.Join(s.outputDir, runID)
	os.MkdirAll(dir, 0755)

	filename := label + ".png"
	path := filepath.Join(dir, filename)

	// If data is base64-encoded, decode it; otherwise write raw
	if decoded, err := base64.StdEncoding.DecodeString(string(data)); err == nil && len(decoded) > 0 {
		data = decoded
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write screenshot: %w", err)
	}

	return fmt.Sprintf("/screenshots/%s/%s", runID, filename), nil
}
