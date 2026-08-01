package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-go-golems/gotest-agent/internal/audit"
	"github.com/go-go-golems/gotest-agent/internal/auth"
)

func (s *Server) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.APIKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		// Try JWT token via cookie (dashboard), header, or query param (SSE)
		if token := auth.GetTokenFromRequest(r); token != "" {
			if _, err := s.jwtAuth.ValidateToken(token); err == nil {
				next.ServeHTTP(w, r)
				return
			}
		}
		// Fallback: X-Api-Key header for API clients
		key := r.Header.Get("X-Api-Key")
		if key == s.cfg.APIKey {
			next.ServeHTTP(w, r)
			return
		}
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
	})
}

// handleLogin authenticates with API key and returns a JWT cookie for dashboard use.
// This endpoint is outside the normal apiKeyAuth middleware so the browser can
// exchange its API key for a cookie-capable session.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}

	userID := "dashboard"
	email := "dashboard"
	role := string(auth.RoleAdmin)
	label := "default"

	// Priority 1: check multi-key store
	if keyRole, keyLabel, keyID, ok := s.keyStore.Validate(req.APIKey); ok {
		userID = keyID
		email = keyLabel
		role = string(keyRole)
		label = keyLabel
	} else if s.cfg.APIKey != "" && req.APIKey == s.cfg.APIKey {
		// Priority 2: fallback to single global API key (backward compatibility)
	} else {
		writeJSONError(w, http.StatusUnauthorized, "invalid api key")
		return
	}

	token, err := s.jwtAuth.GenerateToken(userID, email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	auth.SetTokenCookie(w, token)
	s.recordAudit(r, audit.ActionLogin, audit.ResourceSettings, "auth", label+" login")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "role": role})
}

// handleLogout clears the JWT cookie session.
func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request) {
	auth.ClearTokenCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
