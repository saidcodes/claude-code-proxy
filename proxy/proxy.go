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
