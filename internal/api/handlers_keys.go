package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/audit"
	"github.com/go-go-golems/gotest-agent/internal/auth"
)

// handleCreateAPIKey creates a new API key with role assignment (admin-only).
// POST /api/v1/keys
func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Label string `json:"label"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}

	claims := auth.GetClaims(r.Context())
	createdBy := "system"
	if claims != nil && claims.Email != "" {
		createdBy = claims.Email
	}

	entry, err := s.keyStore.Create(req.Label, auth.Role(req.Role), createdBy)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.recordAudit(r, audit.ActionCreate, audit.ResourceSettings, entry.ID, "created API key: "+entry.Label)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// handleListAPIKeys returns all API keys without plain keys (admin-only).
// GET /api/v1/keys
func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys := s.keyStore.List()
	if keys == nil {
		keys = []auth.APIKeyEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// handleRevokeAPIKey enables/disables an API key (admin-only).
// POST /api/v1/keys/{id}/revoke
func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Active bool `json:"active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if !s.keyStore.Revoke(id, req.Active) {
		writeJSONError(w, http.StatusNotFound, "key not found")
		return
	}

	action := "revoked"
	if req.Active {
		action = "enabled"
	}
	s.recordAudit(r, audit.ActionUpdate, audit.ResourceSettings, id, action+" API key")

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleDeleteAPIKey permanently removes an API key (admin-only).
// DELETE /api/v1/keys/{id}
func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.keyStore.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "key not found")
		return
	}

	s.recordAudit(r, audit.ActionDelete, audit.ResourceSettings, id, "deleted API key")
	w.WriteHeader(http.StatusNoContent)
}
