package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/llmprofile"
)

// handleListLLMProfiles mengembalikan semua profil LLM (API key tersamar).
func (s *Server) handleListLLMProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := s.llmProfiles.List()
	if profiles == nil {
		profiles = []llmprofile.Profile{}
	}
	writeJSON(w, http.StatusOK, profiles)
}

// handleCreateLLMProfile membuat profil LLM baru.
func (s *Server) handleCreateLLMProfile(w http.ResponseWriter, r *http.Request) {
	var p llmprofile.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, err := s.llmProfiles.Create(&p)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// handleUpdateLLMProfile memperbarui profil LLM yang ada.
func (s *Server) handleUpdateLLMProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p llmprofile.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, ok := s.llmProfiles.Update(id, &p)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleDeleteLLMProfile menghapus profil LLM.
func (s *Server) handleDeleteLLMProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.llmProfiles.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleActivateLLMProfile mengaktifkan profil (dipakai saat run test).
func (s *Server) handleActivateLLMProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.llmProfiles.SetActive(id) {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleTestLLMProfile mengetes koneksi profil LLM (memakai API key asli).
func (s *Server) handleTestLLMProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, ok := s.llmProfiles.GetRaw(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}
	client := agent.NewLLM(p.Provider, p.Model, p.APIKey, p.BaseURL)
	if client == nil {
		writeJSONError(w, http.StatusBadRequest, "provider not supported for test")
		return
	}
	if _, err := client.AnalyzeCodebase(r.Context(), "ping"); err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
