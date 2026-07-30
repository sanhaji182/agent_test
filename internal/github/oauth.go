package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OAuthConfig holds GitHub OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// OAuthClient handles GitHub OAuth flow
type OAuthClient struct {
	config OAuthConfig
	http   *http.Client
}

// NewOAuthClient creates a new GitHub OAuth client
func NewOAuthClient(config OAuthConfig) *OAuthClient {
	return &OAuthClient{
		config: config,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthorizationURL returns the GitHub OAuth authorization URL
func (c *OAuthClient) GetAuthorizationURL(state string) string {
	scopes := "repo read:user"
	if len(c.config.Scopes) > 0 {
		scopes = strings.Join(c.config.Scopes, " ")
	}

	// GitHub OAuth expects space-separated scopes with %20 encoding (not + like form data)
	encodedScopes := strings.ReplaceAll(scopes, " ", "%20")

	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		c.config.ClientID,
		c.config.RedirectURL,
		encodedScopes,
		state,
	)
}

// Token represents GitHub OAuth access token
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// ExchangeCode exchanges authorization code for access token
func (c *OAuthClient) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	reqBody := fmt.Sprintf(
		`{"client_id":"%s","client_secret":"%s","code":"%s","redirect_uri":"%s"}`,
		c.config.ClientID,
		c.config.ClientSecret,
		code,
		c.config.RedirectURL,
	)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	return &token, nil
}

// User represents GitHub user information
type User struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GetUser fetches user information using access token
func (c *OAuthClient) GetUser(ctx context.Context, token string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}

	return &user, nil
}

// Repository represents GitHub repository information
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

// GetRepos fetches user repositories using access token
func (c *OAuthClient) GetRepos(ctx context.Context, token string) ([]Repository, error) {
	var repos []Repository
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/user/repos?per_page=100&page=%d", page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("get repos: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("get repos failed: status %d, body: %s", resp.StatusCode, string(body))
		}

		var pageRepos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&pageRepos); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode repos: %w", err)
		}
		resp.Body.Close()

		if len(pageRepos) == 0 {
			break
		}

		repos = append(repos, pageRepos...)
		page++
	}

	return repos, nil
}
