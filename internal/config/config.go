// Package config memuat konfigurasi aplikasi dari environment variables
package config

import "os"

// Config menyimpan semua konfigurasi yang dibutuhkan aplikasi
type Config struct {
	AppPort          string // Port HTTP server (default: 8080)
	APIKey           string // API key untuk autentikasi (kosong = tanpa auth)
	DatabaseURL      string // URL koneksi PostgreSQL
	RedisURL         string // Alamat Redis untuk job queue
	AnthropicAPIKey  string // API key Anthropic untuk LLM
	LLMModel         string // Model LLM yang digunakan (default: claude-sonnet-4-5)
	SteelAPIURL      string // URL Steel Browser API
	SteelAPIKey      string // API key Steel Browser (opsional)
	SteelMaxSessions int    // Maksimal sesi browser bersamaan
	LangGraphURL     string // URL LangGraph sidecar
	MaxFixAttempts   int    // Maksimal percobaan fix test gagal
	TimeoutSeconds   int    // Timeout eksekusi test (detik)
	ScreenshotsPath  string // Direktori penyimpanan screenshot
	ReportsPath      string // Direktori penyimpanan laporan HTML
}

// Load membaca konfigurasi dari environment variables dengan nilai default
func Load() *Config {
	return &Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		APIKey:           getEnv("API_KEY", ""),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gotest_agent?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		LLMModel:         getEnv("LLM_MODEL", "claude-sonnet-4-5"),
		SteelAPIURL:      getEnv("STEEL_API_URL", "http://localhost:3000"),
		SteelAPIKey:      getEnv("STEEL_API_KEY", ""),
		SteelMaxSessions: 10,
		LangGraphURL:     getEnv("LANGGRAPH_URL", "http://localhost:8000"),
		MaxFixAttempts:   3,
		TimeoutSeconds:   300,
		ScreenshotsPath:  getEnv("SCREENSHOTS_PATH", "./data/screenshots"),
		ReportsPath:      getEnv("REPORTS_PATH", "./data/reports"),
	}
}

// getEnv membaca env var, gunakan fallback jika kosong
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
