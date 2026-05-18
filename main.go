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
