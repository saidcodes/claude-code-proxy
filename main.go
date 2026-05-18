package main

import (
	"fmt"
	"log"
	"net/http"

	"claude-code-proxy/proxy"
)

func main() {
	cfg, err := proxy.LoadConfigWithKeyFile()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handler := proxy.NewHandler(cfg)

	addr := fmt.Sprintf(":%s", cfg.ProxyPort)
	log.Printf("Claude Code Proxy starting on %s", addr)
	log.Printf("Upstream: %s", cfg.TargetBaseURL)
	log.Printf("Model: %s", cfg.TargetModel)

	if cfg.DeepSeekAPIKey == "" {
		log.Printf("No API key configured — admin UI available at http://localhost%s/", addr)
	} else {
		log.Println("API key configured")
	}

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
