package config

import "os"

type Config struct {
	AppPort          string
	APIKey           string
	DatabaseURL      string
	RedisURL         string
	AnthropicAPIKey  string
	LLMModel         string
	SteelAPIURL      string
	SteelAPIKey      string
	SteelMaxSessions int
	LangGraphURL     string
	MaxFixAttempts   int
	TimeoutSeconds   int
	ScreenshotsPath  string
	ReportsPath      string
}

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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
