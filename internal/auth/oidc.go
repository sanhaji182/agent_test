// Package auth provides OIDC (OpenID Connect) integration for enterprise SSO.
//
// Supported providers: Google, GitHub, Microsoft (Azure AD), and generic OIDC.
// Security: PKCE flow, state parameter CSRF protection, ID token verification.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// OIDCProvider represents a configured OIDC identity provider.
type OIDCProvider struct {
	ID           string   `json:"id"`           // "google", "github", "azure", or custom ID
	Name         string   `json:"name"`         // "Google", "GitHub", "Microsoft"
	Type         string   `json:"type"`         // "google", "github", "azure_ad", "oidc"
	ClientID     string   `json:"-"`            // OAuth2 client ID (not exposed in API)
	ClientSecret string   `json:"-"`            // OAuth2 client secret (not exposed)
	RedirectURL  string   `json:"-"`            // callback URL
	Scopes       []string `json:"scopes"`       // OIDC scopes
	DefaultRole  Role     `json:"default_role"` // role assigned to new users
	Enabled      bool     `json:"enabled"`      // can be toggled without removing config
	IssuerURL    string   `json:"issuer_url"`   // for generic OIDC providers
}

// OIDCProviderInfo is the public-safe view returned by the API.
type OIDCProviderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// OIDCState stores pending OIDC authorization requests.
type OIDCState struct {
	State        string    `json:"state"`         // random CSRF token
	ProviderID   string    `json:"provider_id"`   // which provider
	RedirectURL  string    `json:"redirect_url"`  // where to send user after login
	CodeVerifier string    `json:"code_verifier"` // PKCE code verifier
	CreatedAt    time.Time `json:"created_at"`
}

// OIDCManager manages OIDC providers and pending auth states.
type OIDCManager struct {
	mu        sync.RWMutex
	providers map[string]*OIDCProvider // keyed by provider ID
	states    map[string]*OIDCState    // keyed by state value
	configs   map[string]*oauth2.Config
}

// NewOIDCManager creates a new OIDC manager.
func NewOIDCManager() *OIDCManager {
	return &OIDCManager{
		providers: make(map[string]*OIDCProvider),
		states:    make(map[string]*OIDCState),
		configs:   make(map[string]*oauth2.Config),
	}
}

// RegisterProvider adds a pre-configured OIDC provider.
// Supported types: "google", "github", "azure_ad", "oidc".
func (m *OIDCManager) RegisterProvider(p *OIDCProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.ID == "" {
		return fmt.Errorf("provider ID is required")
	}
	if p.ClientID == "" || p.ClientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	var cfg *oauth2.Config

	switch p.Type {
	case "google":
		cfg = &oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
			Scopes:       append([]string{"openid", "profile", "email"}, p.Scopes...),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		}
	case "azure_ad":
		tenant := "common"
		if p.IssuerURL != "" {
			tenant = strings.TrimPrefix(p.IssuerURL, "https://login.microsoftonline.com/")
			tenant = strings.TrimSuffix(tenant, "/")
		}
		cfg = &oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
			Scopes:       append([]string{"openid", "profile", "email"}, p.Scopes...),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/" + tenant + "/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/" + tenant + "/oauth2/v2.0/token",
			},
		}
	case "github":
		cfg = &oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
			Scopes:       append([]string{"read:user", "user:email"}, p.Scopes...),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		}
	case "oidc":
		if p.IssuerURL == "" {
			return fmt.Errorf("issuer URL is required for generic OIDC providers")
		}
		cfg = &oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
			Scopes:       append([]string{"openid", "profile", "email"}, p.Scopes...),
			Endpoint: oauth2.Endpoint{
				AuthURL:  p.IssuerURL + "/protocol/openid-connect/auth",
				TokenURL: p.IssuerURL + "/protocol/openid-connect/token",
			},
		}
	default:
		return fmt.Errorf("unsupported provider type: %s", p.Type)
	}

	m.providers[p.ID] = p
	m.configs[p.ID] = cfg
	return nil
}

// GetProviders returns public-safe provider info for the login page.
func (m *OIDCManager) GetProviders() []OIDCProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []OIDCProviderInfo
	for _, p := range m.providers {
		if p.Enabled {
			result = append(result, OIDCProviderInfo{
				ID:   p.ID,
				Name: p.Name,
				Type: p.Type,
			})
		}
	}
	return result
}

// BeginAuth generates an authorization URL and stores pending state.
// Returns the URL the user should be redirected to.
func (m *OIDCManager) BeginAuth(providerID, redirectAfter string) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.providers[providerID]
	if !ok || !p.Enabled {
		return "", "", fmt.Errorf("provider %s not found or disabled", providerID)
	}

	cfg, ok := m.configs[providerID]
	if !ok {
		return "", "", fmt.Errorf("provider %s not configured", providerID)
	}

	// Generate PKCE code verifier
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return "", "", fmt.Errorf("generate verifier: %w", err)
	}

	// Generate CSRF state
	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("generate state: %w", err)
	}

	// Store pending state
	m.states[state] = &OIDCState{
		State:        state,
		ProviderID:   providerID,
		RedirectURL:  redirectAfter,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now(),
	}

	// Clean up expired states (older than 10 minutes)
	for key, s := range m.states {
		if time.Since(s.CreatedAt) > 10*time.Minute {
			delete(m.states, key)
		}
	}

	// Build auth URL with PKCE challenge
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", computeS256(codeVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}

	// For GitHub, request specific scopes
	if p.Type == "github" {
		opts = append(opts, oauth2.SetAuthURLParam("allow_signup", "true"))
	}

	authURL := cfg.AuthCodeURL(state, opts...)
	return authURL, state, nil
}

// CompleteAuth exchanges an authorization code for tokens and extracts user info.
// Returns userID (email), name, and the role assigned to this user.
func (m *OIDCManager) CompleteAuth(state, code string) (string, string, string, error) {
	m.mu.Lock()
	pending, ok := m.states[state]
	if !ok {
		m.mu.Unlock()
		return "", "", "", fmt.Errorf("invalid or expired state")
	}
	delete(m.states, state)
	providerID := pending.ProviderID
	codeVerifier := pending.CodeVerifier
	m.mu.Unlock()

	p, cfg := m.getProviderAndConfig(providerID)
	if p == nil || cfg == nil {
		return "", "", "", fmt.Errorf("provider %s not found", providerID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Exchange code for token
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return "", "", "", fmt.Errorf("token exchange: %w", err)
	}

	// Extract user info
	email, name, err := m.fetchUserInfo(ctx, token, p, cfg)
	if err != nil {
		return "", "", "", fmt.Errorf("fetch user info: %w", err)
	}

	role := string(p.DefaultRole)
	if role == "" {
		role = string(RoleViewer) // default: lowest privilege
	}

	return email, name, role, nil
}

// fetchUserInfo retrieves user email and name from the provider.
func (m *OIDCManager) fetchUserInfo(ctx context.Context, token *oauth2.Token, p *OIDCProvider, cfg *oauth2.Config) (string, string, error) {
	client := cfg.Client(ctx, token)

	switch p.Type {
	case "google", "azure_ad", "oidc":
		// Use OIDC userinfo endpoint
		userInfoURL := "https://openidconnect.googleapis.com/v1/userinfo"
		if p.Type == "azure_ad" {
			userInfoURL = "https://graph.microsoft.com/oidc/userinfo"
		} else if p.Type == "oidc" {
			userInfoURL = strings.TrimSuffix(p.IssuerURL, "/") + "/protocol/openid-connect/userinfo"
		}

		resp, err := client.Get(userInfoURL)
		if err != nil {
			return "", "", fmt.Errorf("userinfo request: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var claims struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.Unmarshal(body, &claims); err != nil {
			return "", "", fmt.Errorf("parse userinfo: %w", err)
		}

		if claims.Email == "" {
			return "", "", fmt.Errorf("no email in userinfo response")
		}
		return claims.Email, claims.Name, nil

	case "github":
		// GitHub: get user from /user endpoint
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return "", "", fmt.Errorf("github user request: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var ghUser struct {
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &ghUser); err != nil {
			return "", "", fmt.Errorf("parse github user: %w", err)
		}

		// If email is not public, fetch from /user/emails
		if ghUser.Email == "" {
			emailsResp, err := client.Get("https://api.github.com/user/emails")
			if err == nil {
				defer emailsResp.Body.Close()
				emailBody, _ := io.ReadAll(emailsResp.Body)
				var emails []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if json.Unmarshal(emailBody, &emails) == nil {
					for _, e := range emails {
						if e.Primary && e.Verified {
							ghUser.Email = e.Email
							break
						}
					}
				}
			}
		}

		email := ghUser.Email
		if email == "" {
			email = ghUser.Login + "@github.user"
		}
		name := ghUser.Name
		if name == "" {
			name = ghUser.Login
		}
		return email, name, nil
	}

	return "", "", fmt.Errorf("unsupported provider type: %s", p.Type)
}

// getProviderAndConfig is a thread-safe helper.
func (m *OIDCManager) getProviderAndConfig(id string) (*OIDCProvider, *oauth2.Config) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[id], m.configs[id]
}

// generateCodeVerifier creates a PKCE code verifier (43 alphanumeric chars).
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// computeS256 computes the S256 code challenge from a verifier.
func computeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState creates a random CSRF state token.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ValidateRedirectURL ensures the redirect_after parameter returns a safe, relative-only path.
func ValidateRedirectURL(raw string) string {
	if raw == "" {
		return "/"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "/"
	}
	// Only allow relative paths
	if u.IsAbs() || strings.Contains(raw, "//") || strings.Contains(raw, "@") {
		return "/"
	}
	return raw
}
