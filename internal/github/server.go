package github

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-go-golems/gotest-agent/internal/parser"
)

// WebhookServer handles GitHub webhook HTTP requests
type WebhookServer struct {
	integration *Integration
	parser      *parser.Registry
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(integration *Integration, parserRegistry *parser.Registry) *WebhookServer {
	return &WebhookServer{
		integration: integration,
		parser:      parserRegistry,
	}
}

// ServeHTTP implements http.Handler interface
func (s *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get webhook handler from integration
	webhookHandler := s.integration.GetWebhookHandler()

	// Parse webhook event using the handler
	event, err := webhookHandler.ParseEvent(r)
	if err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle ping event separately
	if event.Type == "ping" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	// Process the webhook event using the existing integration logic
	ctx := r.Context()
	switch event.Type {
	case "push":
		pushEvent, err := webhookHandler.ParsePushEvent(event.Payload)
		if err != nil {
			log.Printf("Failed to parse push event: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.integration.processPushEvent(ctx, pushEvent); err != nil {
			log.Printf("Failed to process push event: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "pull_request":
		prEvent, err := webhookHandler.ParsePullRequestEvent(event.Payload)
		if err != nil {
			log.Printf("Failed to parse pull_request event: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.integration.processPullRequestEvent(ctx, prEvent); err != nil {
			log.Printf("Failed to process pull_request event: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   event.Type,
	})
}



