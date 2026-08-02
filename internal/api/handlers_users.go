package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/auth"
)

// handleListUsers mengembalikan semua user dashboard (admin-only).
// Password hash tidak ikut (field json:"-").
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users := s.userStore.List()
	if users == nil {
		users = []auth.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

// handleCreateUser membuat user dashboard baru (admin-only).
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	role := auth.Role(req.Role)
	if req.Role == "" {
		role = auth.RoleViewer
	}
	user, err := s.userStore.Create(req.Email, req.Password, req.Name, role)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// handleUpdateUser memperbarui name/role/active/password user (admin-only).
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		IsActive    *bool  `json:"is_active"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	existing, ok := s.userStore.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	role := existing.Role
	if req.Role != "" {
		role = auth.Role(req.Role)
	}
	updated, err := s.userStore.Update(id, req.Name, role, isActive, req.NewPassword)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleDeleteUser menghapus user (admin-only).
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.userStore.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
