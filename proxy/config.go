package proxy

import (
	"fmt"
	"os"
)

type Config struct {
	DeepSeekAPIKey string
	TargetModel    string
	ProxyPort      string
	TargetBaseURL  string
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
	}

	model := os.Getenv("TARGET_MODEL")
	if model == "" {
		model = "deepseek-v4-flash"
	}

	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("TARGET_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/anthropic"
	}

	return &Config{
		DeepSeekAPIKey: apiKey,
		TargetModel:    model,
		ProxyPort:      port,
		TargetBaseURL:  baseURL,
	}, nil
}

// SetAPIKey updates the API key at runtime (e.g., from admin UI).
func (c *Config) SetAPIKey(key string) {
	c.DeepSeekAPIKey = key
}

// LoadConfigWithKeyFile tries env vars first, then falls back to the key file.
// Unlike LoadConfig, it does not return an error if DEEPSEEK_API_KEY is missing.
func LoadConfigWithKeyFile() (*Config, error) {
	cfg, err := LoadConfig()
	if err == nil {
		return cfg, nil
	}

	// Env var is missing — try key file
	cfg = &Config{
		TargetModel:   getEnvDefault("TARGET_MODEL", "deepseek-v4-flash"),
		ProxyPort:     getEnvDefault("PROXY_PORT", "8080"),
		TargetBaseURL: getEnvDefault("TARGET_BASE_URL", "https://api.deepseek.com/anthropic"),
	}

	key, kfErr := LoadKeyFile()
	if kfErr == nil {
		cfg.DeepSeekAPIKey = key
	}
	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
