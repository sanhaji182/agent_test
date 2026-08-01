package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestBrowserEgressGuardBlocksUnsafeRedirectSmoke(t *testing.T) {
	if os.Getenv("GOTEST_BROWSER_EGRESS_SMOKE") != "1" {
		t.Skip("set GOTEST_BROWSER_EGRESS_SMOKE=1 to run live browser redirect/egress smoke test")
	}

	const unsafeRedirectURL = "http://169.254.169.254/latest/meta-data/"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte("<html><body>safe fixture</body></html>"))
		case "/redirect-metadata":
			http.Redirect(w, r, unsafeRedirectURL, http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	fixtureURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse fixture URL: %v", err)
	}
	fixtureHost := fixtureURL.Hostname()
	if fixtureHost == "" {
		t.Fatalf("fixture host is empty for %q", server.URL)
	}

	if err := playwright.Install(); err != nil {
		t.Fatalf("install Playwright browsers: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("start Playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		t.Fatalf("launch Chromium: %v", err)
	}
	defer browser.Close()

	ctx, err := browser.NewContext()
	if err != nil {
		t.Fatalf("new browser context: %v", err)
	}
	defer ctx.Close()
	if err := InstallContextEgressGuard(ctx, fixtureHost); err != nil {
		t.Fatalf("install egress guard: %v", err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	if _, err := page.Goto(server.URL+"/ok", playwright.PageGotoOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("expected allowlisted fixture page to load: %v", err)
	}

	_, err = page.Goto(server.URL+"/redirect-metadata", playwright.PageGotoOptions{Timeout: playwright.Float(5000)})
	if err == nil {
		if strings.HasPrefix(page.URL(), unsafeRedirectURL) {
			t.Fatalf("unsafe redirect loaded despite egress guard: %s", page.URL())
		}
		t.Fatalf("expected unsafe redirect to be blocked; final URL was %s", page.URL())
	}
	if strings.HasPrefix(page.URL(), unsafeRedirectURL) {
		t.Fatalf("unsafe redirect reached metadata endpoint despite navigation error: %s", page.URL())
	}
}
