package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/project"
)

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var p project.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(p.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	p.FeatureMap = s.deriveFeatureMap(r.Context(), p.Spec, p.FocusHints)
	if err := s.projects.Create(r.Context(), &p); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projects.List(r.Context(), 100, 0)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	redacted := make([]*project.Project, len(projects))
	for i, p := range projects {
		redacted[i] = redactProject(p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redacted)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redactProject(p))
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	current, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var patch project.Project
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	mergeProject(current, &patch)
	if patch.Spec != "" || patch.FocusHints != "" {
		current.FeatureMap = s.deriveFeatureMap(r.Context(), current.Spec, current.FocusHints)
	}
	if err := s.projects.Update(r.Context(), current); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(current)
}

func (s *Server) handleUploadAPIDocs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		APIDocs string `json:"api_docs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	p, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	p.APIDocs = req.APIDocs
	if err := s.projects.Update(r.Context(), p); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (s *Server) handleParseAPIDocs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if strings.TrimSpace(p.APIDocs) == "" {
		writeJSONError(w, http.StatusBadRequest, "api_docs is empty")
		return
	}

	s.createDraftPlanResponse(w, r, p.ID, s.parseAPIDocsWithAI(r.Context(), p))
}

func (s *Server) handleExtractProjectFeatures(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	p.FeatureMap = s.deriveFeatureMap(r.Context(), p.Spec, p.FocusHints)
	if err := s.projects.Update(r.Context(), p); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p.FeatureMap)
}

func mergeProject(dst, src *project.Project) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.TestType != "" {
		dst.TestType = src.TestType
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.Environment != "" {
		dst.Environment = src.Environment
	}
	if src.Spec != "" {
		dst.Spec = src.Spec
	}
	if src.APIDocs != "" {
		dst.APIDocs = src.APIDocs
	}
	if src.AuthType != "" {
		dst.AuthType = src.AuthType
	}
	if src.Credentials != "" {
		dst.Credentials = src.Credentials
	}
	if src.FocusHints != "" {
		dst.FocusHints = src.FocusHints
	}
	if src.SkipHints != "" {
		dst.SkipHints = src.SkipHints
	}
	if src.FeatureMap != nil {
		dst.FeatureMap = src.FeatureMap
	}
}
