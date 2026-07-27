package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	// Ensure no interfering env vars.
	for _, k := range []string{"MAX_FIX_ATTEMPTS", "DEFAULT_TIMEOUT_SECONDS", "STEEL_MAX_SESSIONS", "MAX_CONCURRENT_RUNS"} {
		t.Setenv(k, "")
	}
	cfg := Load()
	if cfg.MaxFixAttempts != 3 {
		t.Fatalf("expected MaxFixAttempts default 3, got %d", cfg.MaxFixAttempts)
	}
	if cfg.TimeoutSeconds != 300 {
		t.Fatalf("expected TimeoutSeconds default 300, got %d", cfg.TimeoutSeconds)
	}
	if cfg.SteelMaxSessions != 10 {
		t.Fatalf("expected SteelMaxSessions default 10, got %d", cfg.SteelMaxSessions)
	}
	if cfg.MaxConcurrentRuns != 10 {
		t.Fatalf("expected MaxConcurrentRuns default 10, got %d", cfg.MaxConcurrentRuns)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("MAX_FIX_ATTEMPTS", "5")
	t.Setenv("DEFAULT_TIMEOUT_SECONDS", "120")
	t.Setenv("STEEL_MAX_SESSIONS", "4")
	cfg := Load()
	if cfg.MaxFixAttempts != 5 {
		t.Fatalf("expected MaxFixAttempts 5, got %d", cfg.MaxFixAttempts)
	}
	if cfg.TimeoutSeconds != 120 {
		t.Fatalf("expected TimeoutSeconds 120, got %d", cfg.TimeoutSeconds)
	}
	if cfg.SteelMaxSessions != 4 {
		t.Fatalf("expected SteelMaxSessions 4, got %d", cfg.SteelMaxSessions)
	}
}

func TestLoad_InvalidIntFallsBack(t *testing.T) {
	t.Setenv("MAX_FIX_ATTEMPTS", "not-a-number")
	t.Setenv("DEFAULT_TIMEOUT_SECONDS", "-1")
	t.Setenv("STEEL_MAX_SESSIONS", "0")
	cfg := Load()
	if cfg.MaxFixAttempts != 3 {
		t.Fatalf("expected fallback 3 for invalid MAX_FIX_ATTEMPTS, got %d", cfg.MaxFixAttempts)
	}
	if cfg.TimeoutSeconds != 300 {
		t.Fatalf("expected fallback 300 for invalid DEFAULT_TIMEOUT_SECONDS, got %d", cfg.TimeoutSeconds)
	}
	if cfg.SteelMaxSessions != 10 {
		t.Fatalf("expected fallback 10 for zero STEEL_MAX_SESSIONS, got %d", cfg.SteelMaxSessions)
	}
}
