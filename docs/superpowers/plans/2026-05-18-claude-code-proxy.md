# Claude Code Proxy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a lightweight Go HTTP reverse proxy that lets Claude Code CLI use DeepSeek's Anthropic-compatible API.

**Architecture:** Single Go binary. Receives Claude Code's `/v1/messages` requests, transforms the JSON body (model name, cache_control stripping), swaps the API key, forwards to DeepSeek's API, and streams the response back. No external dependencies beyond the Go standard library.

**Tech Stack:** Go (standard library only — `net/http`, `encoding/json`, `io`)

---

## File Structure

```
claude-code-proxy/
├── main.go                    # Entry point, starts HTTP server
├── go.mod                     # Module definition (claude-code-proxy)
├── proxy/
│   ├── config.go              # Config struct + env var loading
│   ├── config_test.go         # Tests for config loading
│   ├── transforms.go          # Request body JSON transformations
│   ├── transforms_test.go     # Tests for transformations
│   ├── proxy.go               # HTTP reverse proxy handler
│   └── proxy_test.go          # Tests for proxy handler
```

### Task 1: Go Module + Config Package

**Files:**
- Create: `go.mod`
- Create: `proxy/config.go`
- Create: `proxy/config_test.go`

- [ ] **Step 1: Create go.mod**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go mod init claude-code-proxy
```

Expected: `go.mod` created with module `claude-code-proxy` and Go version.

- [ ] **Step 2: Write the config test**

`proxy/config_test.go`:
```go
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
```

- [ ] **Step 3: Run test to verify it fails**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestLoadConfig -v
```

Expected: Compilation error — `LoadConfig` not defined.

- [ ] **Step 4: Write minimal config implementation**

`proxy/config.go`:
```go
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
```

- [ ] **Step 5: Run test to verify it passes**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestLoadConfig -v
```

Expected: All 3 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod proxy/config.go proxy/config_test.go
git commit -m "feat: add config loading with env var support"
```

### Task 2: Request Transformations

**Files:**
- Create: `proxy/transforms.go`
- Create: `proxy/transforms_test.go`

- [ ] **Step 1: Write the transforms test**

`proxy/transforms_test.go`:
```go
package proxy

import (
	"encoding/json"
	"testing"
)

func TestTransformRequest_ModelReplacement(t *testing.T) {
	input := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)
	if result["model"] != "deepseek-v4-flash" {
		t.Errorf("expected deepseek-v4-flash, got %v", result["model"])
	}
}

func TestTransformRequest_StripCacheControlFromSystem(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"system":[{"type":"text","text":"be helpful","cache_control":{"type":"ephemeral"}}],
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)

	sys := result["system"].([]any)
	block := sys[0].(map[string]any)
	if _, ok := block["cache_control"]; ok {
		t.Error("cache_control should be stripped from system blocks")
	}
}

func TestTransformRequest_StripCacheControlFromMessages(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}
		]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)

	msgs := result["messages"].([]any)
	content := msgs[0].(map[string]any)["content"].([]any)
	block := content[0].(map[string]any)
	if _, ok := block["cache_control"]; ok {
		t.Error("cache_control should be stripped from content blocks")
	}
}

func TestTransformRequest_StripDisableParallelToolUse(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"tool_choice":{"type":"auto","disable_parallel_tool_use":true},
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)

	tc := result["tool_choice"].(map[string]any)
	if _, ok := tc["disable_parallel_tool_use"]; ok {
		t.Error("disable_parallel_tool_use should be stripped")
	}
}

func TestTransformRequest_PreservesOtherFields(t *testing.T) {
	input := []byte(`{
		"model":"claude-sonnet-4-6",
		"max_tokens":4096,
		"temperature":0.7,
		"stream":true,
		"system":"be helpful",
		"tools":[{"name":"bash","input_schema":{"type":"object"}}],
		"messages":[{"role":"user","content":"hi"}]
	}`)
	output, err := TransformRequest(input, "deepseek-v4-pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(output, &result)

	if result["model"] != "deepseek-v4-pro" {
		t.Errorf("model should be deepseek-v4-pro, got %v", result["model"])
	}
	if result["max_tokens"] != float64(4096) {
		t.Errorf("max_tokens should be preserved")
	}
	if result["stream"] != true {
		t.Errorf("stream should be preserved")
	}
	if result["temperature"] != 0.7 {
		t.Errorf("temperature should be preserved")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestTransformRequest -v
```

Expected: Compilation error — `TransformRequest` not defined.

- [ ] **Step 3: Write transforms implementation**

`proxy/transforms.go`:
```go
package proxy

import (
	"encoding/json"
)

// TransformRequest modifies the request body for DeepSeek compatibility:
// - Replaces the model name
// - Strips cache_control from system and content blocks
// - Removes disable_parallel_tool_use from tool_choice
func TransformRequest(body []byte, targetModel string) ([]byte, error) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	req["model"] = targetModel

	if system, ok := req["system"]; ok {
		req["system"] = stripCacheControl(system)
	}

	if messages, ok := req["messages"]; ok {
		req["messages"] = stripCacheControlFromMessages(messages)
	}

	if toolChoice, ok := req["tool_choice"]; ok {
		if tc, ok := toolChoice.(map[string]any); ok {
			delete(tc, "disable_parallel_tool_use")
			if len(tc) == 0 {
				delete(req, "tool_choice")
			}
		}
	}

	return json.Marshal(req)
}

func stripCacheControl(v any) any {
	switch val := v.(type) {
	case map[string]any:
		delete(val, "cache_control")
		return val
	case []any:
		for i, item := range val {
			val[i] = stripCacheControl(item)
		}
		return val
	default:
		return v
	}
}

func stripCacheControlFromMessages(v any) any {
	messages, ok := v.([]any)
	if !ok {
		return v
	}
	for _, msg := range messages {
		if m, ok := msg.(map[string]any); ok {
			if content, ok := m["content"]; ok {
				m["content"] = stripCacheControl(content)
			}
		}
	}
	return messages
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestTransformRequest -v
```

Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add proxy/transforms.go proxy/transforms_test.go
git commit -m "feat: add request body transformations for DeepSeek compatibility"
```

### Task 3: Proxy Handler

**Files:**
- Create: `proxy/proxy.go`
- Create: `proxy/proxy_test.go`

- [ ] **Step 1: Write proxy test**

`proxy/proxy_test.go`:
```go
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
	json.Unmarshal(w.Body.Bytes(), &resp)
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
		json.NewDecoder(r.Body).Decode(&body)

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
	json.Unmarshal(w.Body.Bytes(), &resp)
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
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestHandler -v
```

Expected: Compilation error — `NewHandler` not defined.

- [ ] **Step 3: Write proxy handler implementation**

`proxy/proxy.go`:
```go
package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Handler is the HTTP handler for the proxy.
type Handler struct {
	config *Config
	client *http.Client
}

// NewHandler creates a new proxy handler with the given config.
func NewHandler(config *Config) *Handler {
	return &Handler{
		config: config,
		client: &http.Client{},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" && r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	if r.URL.Path != "/v1/messages" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "api_error", "failed to read request body")
		return
	}

	transformed, err := TransformRequest(body, h.config.TargetModel)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "api_error", "failed to transform request")
		return
	}

	upstreamURL := h.config.TargetBaseURL + "/v1/messages"
	upstreamReq, err := http.NewRequest(http.MethodPost, upstreamURL, bytes.NewReader(transformed))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "api_error", "failed to create upstream request")
		return
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("x-api-key", h.config.DeepSeekAPIKey)
	upstreamReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := h.client.Do(upstreamReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "api_error", "upstream request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func writeError(w http.ResponseWriter, status int, errType, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"type":    errType,
			"message": msg,
		},
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -run TestHandler -v
```

Expected: All 4 tests PASS.

- [ ] **Step 5: Run all tests**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -v
```

Expected: All 9 tests PASS (5 transform + 4 handler).

- [ ] **Step 6: Commit**

```bash
git add proxy/proxy.go proxy/proxy_test.go
git commit -m "feat: add reverse proxy handler with request forwarding"
```

### Task 4: Main Entry Point

**Files:**
- Create: `main.go`

- [ ] **Step 1: Write main.go**

`main.go`:
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"claude-code-proxy/proxy"
)

func main() {
	cfg, err := proxy.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handler := proxy.NewHandler(cfg)

	addr := fmt.Sprintf(":%s", cfg.ProxyPort)
	log.Printf("Claude Code Proxy starting on %s", addr)
	log.Printf("Upstream: %s", cfg.TargetBaseURL)
	log.Printf("Model: %s", cfg.TargetModel)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

- [ ] **Step 2: Build and verify**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go build -o claude-code-proxy.exe .
```

Expected: Binary `claude-code-proxy.exe` created with no errors.

- [ ] **Step 3: Run all tests once more**

Run:
```bash
cd /c/Users/Pc/projects/claude-code-proxy
go test ./proxy/ -v
```

Expected: All 9 tests PASS. No vet warnings.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: add main entry point"
```
