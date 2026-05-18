package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_HealthCheck(t *testing.T) {
	cfg := &Config{
		DeepSeekAPIKey: "sk-test",
		TargetModel:    "deepseek-v4-flash",
		TargetBaseURL:  "http://localhost:9999",
	}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestHandler_NotFound(t *testing.T) {
	cfg := &Config{
		DeepSeekAPIKey: "sk-test",
		TargetModel:    "deepseek-v4-flash",
		TargetBaseURL:  "http://localhost:9999",
	}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/other", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_ForwardsToUpstream(t *testing.T) {
	// Start a test upstream server that mimics DeepSeek
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request was transformed
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode upstream request: %v", err)
			return
		}

		if body["model"] != "deepseek-v4-flash" {
			t.Errorf("expected model deepseek-v4-flash, got %v", body["model"])
		}
		if r.Header.Get("x-api-key") != "sk-proxy-key" {
			t.Errorf("expected x-api-key sk-proxy-key, got %s", r.Header.Get("x-api-key"))
		}

		// Respond with Anthropic-compatible format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello!"}],
			"model": "deepseek-v4-flash",
			"stop_reason": "end_turn"
		}`))
	}))
	defer upstream.Close()

	cfg := &Config{
		DeepSeekAPIKey: "sk-proxy-key",
		TargetModel:    "deepseek-v4-flash",
		TargetBaseURL:  upstream.URL,
	}
	h := NewHandler(cfg)

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "sk-ignored")
	req.Header.Set("anthropic-beta", "some-beta-2025-01-01")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["id"] != "msg_123" {
		t.Errorf("expected msg_123, got %v", resp["id"])
	}
}

func TestHandler_UpstreamError(t *testing.T) {
	// Upstream that returns an error
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"type":"authentication_error","message":"invalid api key"}}`))
	}))
	defer upstream.Close()

	cfg := &Config{
		DeepSeekAPIKey: "sk-bad-key",
		TargetModel:    "deepseek-v4-flash",
		TargetBaseURL:  upstream.URL,
	}
	h := NewHandler(cfg)

	body := `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "sk-ignored")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
