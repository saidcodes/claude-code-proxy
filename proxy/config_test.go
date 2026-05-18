package proxy

import (
	"os"
	"testing"
)

func TestLoadConfig_MissingAPIKey(t *testing.T) {
	os.Unsetenv("DEEPSEEK_API_KEY")
	os.Unsetenv("TARGET_MODEL")
	os.Unsetenv("PROXY_PORT")
	os.Unsetenv("TARGET_BASE_URL")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing DEEPSEEK_API_KEY")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	os.Setenv("DEEPSEEK_API_KEY", "sk-test-key")
	defer os.Unsetenv("DEEPSEEK_API_KEY")
	os.Unsetenv("TARGET_MODEL")
	os.Unsetenv("PROXY_PORT")
	os.Unsetenv("TARGET_BASE_URL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DeepSeekAPIKey != "sk-test-key" {
		t.Errorf("expected sk-test-key, got %s", cfg.DeepSeekAPIKey)
	}
	if cfg.TargetModel != "deepseek-v4-flash" {
		t.Errorf("expected default deepseek-v4-flash, got %s", cfg.TargetModel)
	}
	if cfg.ProxyPort != "8080" {
		t.Errorf("expected default 8080, got %s", cfg.ProxyPort)
	}
	if cfg.TargetBaseURL != "https://api.deepseek.com/anthropic" {
		t.Errorf("expected default https://api.deepseek.com/anthropic, got %s", cfg.TargetBaseURL)
	}
}

func TestLoadConfig_Custom(t *testing.T) {
	os.Setenv("DEEPSEEK_API_KEY", "sk-custom")
	os.Setenv("TARGET_MODEL", "deepseek-v4-pro")
	os.Setenv("PROXY_PORT", "9090")
	os.Setenv("TARGET_BASE_URL", "https://custom.example.com")
	defer func() {
		os.Unsetenv("DEEPSEEK_API_KEY")
		os.Unsetenv("TARGET_MODEL")
		os.Unsetenv("PROXY_PORT")
		os.Unsetenv("TARGET_BASE_URL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DeepSeekAPIKey != "sk-custom" {
		t.Errorf("expected sk-custom, got %s", cfg.DeepSeekAPIKey)
	}
	if cfg.TargetModel != "deepseek-v4-pro" {
		t.Errorf("expected deepseek-v4-pro, got %s", cfg.TargetModel)
	}
	if cfg.ProxyPort != "9090" {
		t.Errorf("expected 9090, got %s", cfg.ProxyPort)
	}
	if cfg.TargetBaseURL != "https://custom.example.com" {
		t.Errorf("expected https://custom.example.com, got %s", cfg.TargetBaseURL)
	}
}
