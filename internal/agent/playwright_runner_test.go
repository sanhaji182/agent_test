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
