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
