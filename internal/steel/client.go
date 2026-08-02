// Package steel menyediakan client untuk berkomunikasi dengan Steel Browser API.
// Steel Browser adalah headless browser self-hosted untuk menjalankan Playwright test.
//
// EXPERIMENTAL: Not wired to cmd/server. The primary browser backend is
// playwright-go (local) or Docker Playwright runner. Wire this if you want
// Steel Browser as an alternative execution backend.
package steel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client adalah HTTP client untuk Steel Browser API
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// Session merepresentasikan satu sesi browser di Steel
type Session struct {
	ID          string `json:"id"`           // ID unik sesi
	CDPURL      string `json:"websocketUrl"` // WebSocket URL untuk Chrome DevTools Protocol
	SeleniumURL string `json:"seleniumUrl"`  // URL Selenium (alternatif)
	CreatedAt   string `json:"createdAt"`
}

// NewClient membuat Steel client baru
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateSession membuat sesi browser baru di Steel
// Mengembalikan session dengan CDP URL untuk koneksi Playwright
func (c *Client) CreateSession(ctx context.Context) (*Session, error) {
	body, _ := json.Marshal(map[string]any{"timeout": 300000}) // 5 menit timeout
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/sessions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create session: status %d: %s", resp.StatusCode, string(b))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	c.rewriteSessionHost(&session)
	return &session, nil
}

// rewriteSessionHost mengganti host internal Steel (mis. 0.0.0.0) di CDPURL
// dengan host baseURL. Server WebSocket Steel menolak Host header yang bukan
// IP atau localhost (proteksi DNS-rebinding), jadi kita resolve hostname ke IP
// agar Host header koneksi CDP lolos validasi.
func (c *Client) rewriteSessionHost(session *Session) {
	if session.CDPURL == "" {
		return
	}
	u, err := url.Parse(session.CDPURL)
	if err != nil {
		return
	}
	base, err := url.Parse(c.baseURL)
	if err != nil || base.Host == "" {
		return
	}
	host := base.Hostname()
	port := base.Port()
	if port == "" {
		port = u.Port()
	}
	if ip := lookupFirstIP(host); ip != "" {
		host = ip
	}
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}
	session.CDPURL = u.String()
}

// lookupFirstIP mengembalikan IP pertama hasil resolve host. Jika host sudah
// berupa IP, dikembalikan apa adanya. Kembali "" jika resolve gagal.
func lookupFirstIP(host string) string {
	if net.ParseIP(host) != nil {
		return host
	}
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0]
}

// GetSession mengambil detail sesi berdasarkan ID
func (c *Client) GetSession(ctx context.Context, id string) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/sessions/"+id, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DestroySession menghapus sesi browser dan membersihkan resource
func (c *Client) DestroySession(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/v1/sessions/"+id, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Screenshot mengambil screenshot dari sesi browser
// fullPage=true untuk screenshot seluruh halaman
func (c *Client) Screenshot(ctx context.Context, id string, fullPage bool) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/sessions/%s/screenshot?fullPage=%t", c.baseURL, id, fullPage)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ListSessions menampilkan semua sesi browser yang aktif
func (c *Client) ListSessions(ctx context.Context) ([]*Session, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/sessions", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []*Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// setHeaders mengatur header standar untuk setiap request
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}
