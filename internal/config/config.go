// Package config memuat konfigurasi aplikasi dari environment variables
package config

import "os"

// Config menyimpan semua konfigurasi yang dibutuhkan aplikasi
type Config struct {
	AppPort             string // Port HTTP server (default: 8080)
	AppEnv              string // Environment: development, production (default: development)
	APIKey              string // API key untuk autentikasi (kosong = tanpa auth)
	CORSAllowedOrigins  string // Daftar origin CORS dipisah koma (kosong/"*" = wildcard, development)
	GitHubWebhookSecret string // Secret verifikasi HMAC webhook GitHub
	JWTSecret           string // Secret untuk JWT cookie auth dashboard
	DatabaseURL         string // URL koneksi PostgreSQL
	RedisURL            string // Alamat Redis untuk job queue
	QueueEnabled        bool   // Aktifkan durable job queue Redis/Asynq (default: false, in-process goroutines)
	AnthropicAPIKey     string // API key Anthropic untuk LLM (fallback ketika LLM_API_KEY kosong)
	LLMModel            string // Model LLM yang digunakan (default: claude-sonnet-4-5)
	LLMProvider         string // Provider LLM: anthropic, openai, google, deepseek, etc.
	LLMBaseURL          string // Base URL untuk OpenAI-compatible endpoint
	SteelAPIURL         string // URL Steel Browser API
	SteelAPIKey         string // API key Steel Browser (opsional)
	SteelMaxSessions    int    // Maksimal sesi browser bersamaan
	MaxConcurrentRuns   int    // Maksimal concurrent test run goroutines (default: 10)
	LangGraphURL        string // URL LangGraph sidecar
	MaxFixAttempts      int    // Maksimal percobaan fix test gagal
	TimeoutSeconds      int    // Timeout eksekusi test (detik)
	ScreenshotsPath     string // Direktori penyimpanan screenshot
	ReportsPath         string // Direktori penyimpanan laporan HTML
	TracingEnabled      bool   // Enable distributed tracing (OTLP)
	TracingEndpoint     string // OTLP gRPC endpoint (default: localhost:4317)
	ServiceVersion      string // Service version for tracing (default: 1.0.0)
}

// Load membaca konfigurasi dari environment variables dengan nilai default
func Load() *Config {
	return &Config{
		AppPort:             getEnv("APP_PORT", "8080"),
		AppEnv:              getEnv("APP_ENV", "development"),
		APIKey:              getEnv("API_KEY", ""),
		CORSAllowedOrigins:  getEnv("CORS_ALLOWED_ORIGINS", ""),
		GitHubWebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gotest_agent?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		QueueEnabled:        getEnv("QUEUE_ENABLED", "") == "true",
		AnthropicAPIKey:     getEnv("ANTHROPIC_API_KEY", ""),
		LLMModel:            getEnv("LLM_MODEL", "claude-sonnet-4-5"),
		LLMProvider:         getEnv("LLM_PROVIDER", "anthropic"),
		LLMBaseURL:          getEnv("LLM_BASE_URL", ""),
		SteelAPIURL:         getEnv("STEEL_API_URL", "http://localhost:3000"),
		SteelAPIKey:         getEnv("STEEL_API_KEY", ""),
		SteelMaxSessions:    getEnvInt("STEEL_MAX_SESSIONS", 10),
		MaxConcurrentRuns:   getEnvInt("MAX_CONCURRENT_RUNS", 10),
		LangGraphURL:        getEnv("LANGGRAPH_URL", "http://localhost:8000"),
		MaxFixAttempts:      getEnvInt("MAX_FIX_ATTEMPTS", 3),
		TimeoutSeconds:      getEnvInt("DEFAULT_TIMEOUT_SECONDS", 300),
		ScreenshotsPath:     getEnv("SCREENSHOTS_PATH", "./data/screenshots"),
		ReportsPath:         getEnv("REPORTS_PATH", "./data/reports"),
		TracingEnabled:      getEnv("TRACING_ENABLED", "false") == "true",
		TracingEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceVersion:      getEnv("SERVICE_VERSION", "1.0.0"),
	}
}

// getEnv membaca env var, gunakan fallback jika kosong
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt reads an integer environment variable, returning fallback if unset or invalid.
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return fallback
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 {
		return fallback
	}
	return n
}
