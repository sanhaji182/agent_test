package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
)

// handleRegisterWebhook registers a new GitHub webhook for continuous sync.
func (s *Server) handleRegisterWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RepositoryURL string `json:"repository_url"`
		GithubToken   string `json:"github_token"`
		Secret        string `json:"secret,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.RepositoryURL == "" {
		writeJSONError(w, http.StatusBadRequest, "repository_url is required")
		return
	}
	reg := s.webhooks.Create(webhook.WebhookRegistration{
		RepositoryURL: body.RepositoryURL,
		GithubToken:   body.GithubToken,
		Secret:        body.Secret,
	})
	writeJSON(w, http.StatusCreated, map[string]string{
		"status":     "registered",
		"webhook_id": reg.ID,
	})
}

// handleListWebhooks returns all registered webhooks.
func (s *Server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	regs := s.webhooks.List()
	writeJSON(w, http.StatusOK, regs)
}

// handleGetWebhook returns a single webhook registration.
func (s *Server) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	reg, ok := s.webhooks.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "webhook not found")
		return
	}
	writeJSON(w, http.StatusOK, reg)
}

// handleUpdateWebhookStatus updates a webhook's status.
func (s *Server) handleUpdateWebhookStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := s.webhooks.UpdateStatus(id, body.Status)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleDeleteWebhook deletes a webhook registration.
func (s *Server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.webhooks.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "webhook not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSyncWebhook triggers a manual sync for a webhook registration.
func (s *Server) handleSyncWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.webhooks.UpdateLastSync(id) {
		writeJSONError(w, http.StatusNotFound, "webhook not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "synced",
	})
}
