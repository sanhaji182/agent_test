package chrome_extension

import (
	"encoding/json"
	"image/png"
	"os"
	"strings"
	"testing"
)

type manifest struct {
	ManifestVersion int      `json:"manifest_version"`
	Name            string   `json:"name"`
	Permissions     []string `json:"permissions"`
	HostPermissions []string `json:"host_permissions"`
	Action          struct {
		DefaultPopup string            `json:"default_popup"`
		DefaultIcon  map[string]string `json:"default_icon"`
	} `json:"action"`
	Background struct {
		ServiceWorker string `json:"service_worker"`
	} `json:"background"`
	ContentScripts []struct {
		Matches []string `json:"matches"`
		JS      []string `json:"js"`
		RunAt   string   `json:"run_at"`
	} `json:"content_scripts"`
}

func TestManifestReferencesExistingExtensionAssets(t *testing.T) {
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	if m.ManifestVersion != 3 {
		t.Fatalf("expected manifest v3, got %d", m.ManifestVersion)
	}
	if m.Name == "" {
		t.Fatal("manifest name is required")
	}
	assertFileExists(t, m.Action.DefaultPopup)
	assertFileExists(t, m.Background.ServiceWorker)
	if len(m.ContentScripts) == 0 {
		t.Fatal("expected at least one content script")
	}
	for _, script := range m.ContentScripts {
		for _, path := range script.JS {
			assertFileExists(t, path)
		}
	}
	for size, path := range m.Action.DefaultIcon {
		assertPNGExists(t, path, size)
	}
}

func TestExtensionUsesBackendAPIKeyHeader(t *testing.T) {
	data, err := os.ReadFile("background.js")
	if err != nil {
		t.Fatalf("read background.js: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "'X-Api-Key': apiKey") {
		t.Fatal("background.js should send API keys with X-Api-Key to match backend auth middleware")
	}
	if strings.Contains(content, "Authorization: `Bearer ${apiKey}`") {
		t.Fatal("background.js must not treat the raw API key as a JWT bearer token")
	}
	if !strings.Contains(content, "normalizeHTTPURL") {
		t.Fatal("background.js should validate backend and target URLs before recording")
	}
}

func TestPopupHandlesAsyncErrorResponses(t *testing.T) {
	data, err := os.ReadFile("popup.js")
	if err != nil {
		t.Fatalf("read popup.js: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "session.ok === false") {
		t.Fatal("popup.js should surface START_RECORDING errors returned by the service worker")
	}
	if !strings.Contains(content, "result.ok === false") {
		t.Fatal("popup.js should surface STOP_RECORDING and settings errors returned by the service worker")
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		t.Fatal("manifest contains empty asset path")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("manifest asset %s is missing: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("manifest asset %s is a directory", path)
	}
}

func assertPNGExists(t *testing.T, path string, size string) {
	t.Helper()
	assertFileExists(t, path)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open icon %s: %v", path, err)
	}
	defer file.Close()
	img, err := png.DecodeConfig(file)
	if err != nil {
		t.Fatalf("icon %s is not a valid PNG: %v", path, err)
	}
	expected := map[string]int{"16": 16, "48": 48, "128": 128}[size]
	if expected == 0 {
		t.Fatalf("unexpected icon size key %s for %s", size, path)
	}
	if img.Width != expected || img.Height != expected {
		t.Fatalf("icon %s has size %dx%d, expected %dx%d", path, img.Width, img.Height, expected, expected)
	}
}
