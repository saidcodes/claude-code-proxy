package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdmin_GetStatus_NotConfigured(t *testing.T) {
	cfg := &Config{}
	h := NewAdminHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin/status", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["configured"] != false {
		t.Errorf("expected configured=false, got %v", resp["configured"])
	}
}

func TestAdmin_GetStatus_Configured(t *testing.T) {
	cfg := &Config{DeepSeekAPIKey: "sk-secret-123"}
	h := NewAdminHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin/status", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["configured"] != true {
		t.Errorf("expected configured=true, got %v", resp["configured"])
	}
	masked := resp["masked_key"].(string)
	if !strings.Contains(masked, "sk-") || strings.Contains(masked, "123") {
		t.Errorf("key should be masked, got %s", masked)
	}
}

func TestAdmin_GetIndex(t *testing.T) {
	cfg := &Config{}
	h := NewAdminHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Claude Code Proxy") {
		t.Error("page should contain title")
	}
}

func TestAdmin_PostKey(t *testing.T) {
	dir := t.TempDir()
	original := keyFilePath
	keyFilePath = func() string { return filepath.Join(dir, "key.json") }
	defer func() { keyFilePath = original }()

	cfg := &Config{}
	h := NewAdminHandler(cfg)

	body := `{"api_key":"sk-new-key"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/key", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if cfg.DeepSeekAPIKey != "sk-new-key" {
		t.Errorf("expected sk-new-key, got %s", cfg.DeepSeekAPIKey)
	}

	// Verify it was saved to disk
	loaded, err := LoadKeyFile()
	if err != nil {
		t.Fatalf("failed to load key file: %v", err)
	}
	if loaded != "sk-new-key" {
		t.Errorf("expected sk-new-key from file, got %s", loaded)
	}
}

func TestAdmin_DeleteKey(t *testing.T) {
	dir := t.TempDir()
	original := keyFilePath
	keyFilePath = func() string { return filepath.Join(dir, "key.json") }
	defer func() { keyFilePath = original }()

	SaveKeyFile("sk-old-key")

	cfg := &Config{DeepSeekAPIKey: "sk-old-key"}
	h := NewAdminHandler(cfg)

	req := httptest.NewRequest(http.MethodDelete, "/admin/key", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if cfg.DeepSeekAPIKey != "" {
		t.Errorf("expected empty key, got %s", cfg.DeepSeekAPIKey)
	}
	// Verify file was removed from disk
	if _, err := LoadKeyFile(); err == nil {
		t.Error("expected key file to be removed after delete")
	}
}
