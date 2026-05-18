package proxy

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

//go:embed admin.html
var adminPage string

// AdminHandler serves the admin UI and handles key management.
type AdminHandler struct {
	config *Config
}

// NewAdminHandler creates an admin handler tied to the given config.
func NewAdminHandler(config *Config) *AdminHandler {
	return &AdminHandler{config: config}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/", "":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(adminPage))

	case "/admin/status":
		h.handleStatus(w, r)

	case "/admin/key":
		switch r.Method {
		case http.MethodPost:
			h.handlePostKey(w, r)
		case http.MethodDelete:
			h.handleDeleteKey(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

	default:
		http.NotFound(w, r)
	}
}

func (h *AdminHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	configured := h.config.DeepSeekAPIKey != ""
	resp := map[string]any{
		"configured": configured,
	}
	if configured {
		key := h.config.DeepSeekAPIKey
		if len(key) >= 8 {
			resp["masked_key"] = key[:7] + "..."
		} else {
			resp["masked_key"] = "sk-...****"
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *AdminHandler) handlePostKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if !strings.HasPrefix(body.APIKey, "sk-") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key must start with sk-"})
		return
	}

	if err := SaveKeyFile(body.APIKey); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save key"})
		return
	}
	h.config.SetAPIKey(body.APIKey)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminHandler) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	h.config.SetAPIKey("")
	os.Remove(keyFilePath())
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
