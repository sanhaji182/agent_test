package agent

import "testing"

func TestIsSafeBrowserURL(t *testing.T) {
	tests := []struct {
		url  string
		safe bool
	}{
		// Safe public URLs
		{"https://example.com", true},
		{"https://github.com/go-go-golems", true},
		{"https://api.anthropic.com/v1", true},
		{"http://myapp.local:3000", true},

		// Loopback
		{"http://127.0.0.1:8080", false},
		{"http://localhost", false},
		{"http://[::1]", false},

		// Private RFC1918
		{"http://192.168.1.1", false},
		{"http://10.0.0.1", false},
		{"http://172.16.0.1", false},

		// Link-local
		{"http://169.254.1.1", false},

		// Cloud metadata
		{"http://169.254.169.254", false},
		{"http://metadata.google.internal", false},
		{"http://metadata", false},

		// Invalid
		{"", false},
		{"not-a-url", false},
		{"://missing-scheme", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isSafeBrowserURL(tt.url)
			if got != tt.safe {
				t.Fatalf("isSafeBrowserURL(%q) = %v, want %v", tt.url, got, tt.safe)
			}
		})
	}
}

func TestExpandString_ReplacesAllKeys(t *testing.T) {
	data := map[string]string{
		"username": "admin",
		"password": "secret123",
		"domain":   "example.com",
	}
	cases := []struct {
		input string
		want  string
	}{
		{"https://{{domain}}/login", "https://example.com/login"},
		{"{{username}}", "admin"},
		{"user={{username}}&pass={{password}}", "user=admin&pass=secret123"},
		{"no templates here", "no templates here"},
		{"", ""},
		{"{{unknown}}", "{{unknown}}"},
	}
	for _, tc := range cases {
		got := expandString(tc.input, data)
		if got != tc.want {
			t.Errorf("expandString(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExpandTemplate_AllFields(t *testing.T) {
	r := &PlaywrightRunner{
		TestData: map[string]string{"url": "https://app.io", "email": "test@test.com"},
	}
	a := &BrowserAction{
		Action:     "fill",
		URL:        "{{url}}/login",
		Selector:   "input[name='email']",
		Value:      "{{email}}",
		Text:       "Welcome to {{url}}",
		NetworkURL: "{{url}}/api",
	}
	r.expandTemplate(a)

	if a.URL != "https://app.io/login" {
		t.Errorf("URL = %q", a.URL)
	}
	if a.Value != "test@test.com" {
		t.Errorf("Value = %q", a.Value)
	}
	if a.Text != "Welcome to https://app.io" {
		t.Errorf("Text = %q", a.Text)
	}
	if a.NetworkURL != "https://app.io/api" {
		t.Errorf("NetworkURL = %q", a.NetworkURL)
	}
	if a.Selector != "input[name='email']" {
		t.Errorf("Selector should be unchanged, got %q", a.Selector)
	}
}

func TestExpandTemplate_NoData_NoOp(t *testing.T) {
	r := &PlaywrightRunner{}
	a := &BrowserAction{URL: "{{url}}/test", Value: "{{val}}"}
	r.expandTemplate(a)
	if a.URL != "{{url}}/test" {
		t.Errorf("expected no-op, got URL = %q", a.URL)
	}
}

func TestViewportPresets_AllValid(t *testing.T) {
	for name, vp := range ViewportPresets {
		if vp.Width <= 0 || vp.Height <= 0 {
			t.Errorf("preset %q has invalid dimensions: %dx%d", name, vp.Width, vp.Height)
		}
		if vp.Name != name {
			t.Errorf("preset %q has mismatched Name field: %q", name, vp.Name)
		}
	}
	expected := []string{"desktop", "desktop-hd", "ipad", "iphone-14", "pixel-7", "galaxy-s23"}
	for _, name := range expected {
		if _, ok := ViewportPresets[name]; !ok {
			t.Errorf("expected preset %q not found", name)
		}
	}
}

func TestWithBrowser_SetsType(t *testing.T) {
	r := NewPlaywrightRunner("/tmp", nil)
	if r.BrowserType != "chromium" {
		t.Errorf("default browser = %q, want chromium", r.BrowserType)
	}
	r.WithBrowser("firefox")
	if r.BrowserType != "firefox" {
		t.Errorf("browser = %q, want firefox", r.BrowserType)
	}
	r.WithBrowser("webkit")
	if r.BrowserType != "webkit" {
		t.Errorf("browser = %q, want webkit", r.BrowserType)
	}
}

func TestWithViewport_SetsPreset(t *testing.T) {
	r := NewPlaywrightRunner("/tmp", nil)
	r.WithViewport("iphone-14")
	if r.Viewport == nil {
		t.Fatal("viewport not set")
	}
	if r.Viewport.Width != 390 || r.Viewport.Height != 844 {
		t.Errorf("iphone-14 viewport = %dx%d", r.Viewport.Width, r.Viewport.Height)
	}
	if !r.Viewport.IsMobile {
		t.Error("iphone-14 should be mobile")
	}

	r2 := NewPlaywrightRunner("/tmp", nil)
	r2.WithViewport("nonexistent")
	if r2.Viewport != nil {
		t.Error("unknown preset should not set viewport")
	}
}

func TestWithParallel(t *testing.T) {
	r := NewPlaywrightRunner("/tmp", nil)
	if r.Parallel {
		t.Error("parallel should default to false")
	}
	r.WithParallel(true)
	if !r.Parallel {
		t.Error("parallel should be true after WithParallel(true)")
	}
}
