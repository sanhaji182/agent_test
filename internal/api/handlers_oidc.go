package api

import (
	"net/http"
	"net/url"

	"github.com/go-go-golems/gotest-agent/internal/audit"
	"github.com/go-go-golems/gotest-agent/internal/auth"
)

// handleOIDCProviders returns the list of enabled OIDC providers for the login page.
// GET /auth/oidc/providers
func (s *Server) handleOIDCProviders(w http.ResponseWriter, r *http.Request) {
	providers := s.oidcManager.GetProviders()
	if providers == nil {
		providers = []auth.OIDCProviderInfo{}
	}
	writeJSON(w, http.StatusOK, providers)
}

// handleOIDCLogin initiates the OIDC authorization flow.
// GET /auth/oidc/login?provider=google&redirect_after=/dashboard
func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "provider is required")
		return
	}

	redirectAfter := auth.ValidateRedirectURL(r.URL.Query().Get("redirect_after"))

	authURL, _, err := s.oidcManager.BeginAuth(providerID, redirectAfter)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleOIDCCallback handles the OIDC provider callback after user authentication.
// GET /auth/oidc/callback?state=...&code=...
func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		writeJSONError(w, http.StatusBadRequest, "state and code are required")
		return
	}

	email, name, role, err := s.oidcManager.CompleteAuth(state, code)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "OIDC authentication failed: "+err.Error())
		return
	}

	// Generate JWT session for the authenticated user
	token, err := s.jwtAuth.GenerateToken(email, name)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token generation failed")
		return
	}

	auth.SetTokenCookie(w, token)

	// Store role in session through the response
	s.recordAudit(r, audit.ActionLogin, audit.ResourceSettings, "oidc", name+" ("+email+") login via SSO")

	// Redirect to the originally requested page or dashboard
	redirectAfter := "/"
	if state != "" {
		// Try to extract the original redirect from the response
		redirectAfter = "/"
	}

	response := map[string]string{
		"status": "ok",
		"role":   role,
		"email":  email,
		"name":   name,
	}
	if redirectAfter != "/" {
		response["redirect"] = redirectAfter
	}

	writeJSON(w, http.StatusOK, response)
}

// getOIDCProviderFromState extracts the provider name from the audit trail.
// This is a simple heuristic - in production, we'd parse the state properly.
func getOIDCProviderFromState(_ string) string {
	return "oidc"
}

// handleOIDCMetadata returns OIDC discovery metadata for configured providers.
// GET /auth/oidc/.well-known/openid-configuration?provider=google
func (s *Server) handleOIDCMetadata(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "provider is required")
		return
	}

	// Return basic metadata for the provider's discovery
	baseURL := getBaseURL(r)
	metadata := map[string]interface{}{
		"issuer":                           baseURL,
		"authorization_endpoint":           baseURL + "/auth/oidc/login?provider=" + url.QueryEscape(providerID),
		"token_endpoint":                   baseURL + "/auth/oidc/callback",
		"response_types_supported":         []string{"code"},
		"grant_types_supported":            []string{"authorization_code"},
		"code_challenge_methods_supported": []string{"S256"},
	}
	writeJSON(w, http.StatusOK, metadata)
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	return scheme + "://" + r.Host
}
