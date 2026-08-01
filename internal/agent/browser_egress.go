package agent

import (
	"net"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// IsSafeBrowserURL validates that a browser navigation/request target will not
// reach internal infrastructure by default. Local/private targets may be allowed
// explicitly via allowedHosts (comma-separated strings are accepted).
func IsSafeBrowserURL(rawURL string, allowedHosts ...string) bool {
	u, ok := parseBrowserURL(rawURL)
	if !ok {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	if isCloudMetadataHost(host) {
		return false
	}
	if isAllowedBrowserHost(host, allowedHosts) {
		return true
	}
	if isBlockedBrowserHostname(host) {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return isSafeBrowserIP(ip)
	}

	if ips, err := net.LookupIP(host); err == nil {
		for _, resolved := range ips {
			if !isSafeBrowserIP(resolved) {
				return false
			}
		}
	}

	return true
}

func isSafeBrowserURL(rawURL string) bool {
	return IsSafeBrowserURL(rawURL)
}

// InstallContextEgressGuard blocks unsafe requests at the browser-context level.
// This catches redirects and link-click navigations, not just explicit page.Goto calls.
func InstallContextEgressGuard(bCtx playwright.BrowserContext, allowedHosts ...string) error {
	return bCtx.Route("**/*", func(route playwright.Route) {
		if !IsSafeBrowserURL(route.Request().URL(), allowedHosts...) {
			_ = route.Abort("blockedbyclient")
			return
		}
		_ = route.Continue()
	})
}

// InstallPageEgressGuard blocks unsafe requests for a single page.
func InstallPageEgressGuard(page playwright.Page, allowedHosts ...string) error {
	return page.Route("**/*", func(route playwright.Route) {
		if !IsSafeBrowserURL(route.Request().URL(), allowedHosts...) {
			_ = route.Abort("blockedbyclient")
			return
		}
		_ = route.Continue()
	})
}

func parseBrowserURL(rawURL string) (*url.URL, bool) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, false
	}
	if u.User != nil {
		return nil, false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, false
	}
	if u.Hostname() == "" {
		return nil, false
	}
	return u, true
}

func isCloudMetadataHost(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	return host == "169.254.169.254" || host == "metadata" || host == "metadata.google.internal"
}

func isBlockedBrowserHostname(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "localhost" {
		return true
	}
	if strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") ||
		strings.HasSuffix(host, ".lan") ||
		strings.HasSuffix(host, ".home") ||
		strings.HasSuffix(host, ".corp") {
		return true
	}
	return false
}

func isSafeBrowserIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return !ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsPrivate() &&
		!ip.IsUnspecified() &&
		!ip.IsMulticast()
}

func isAllowedBrowserHost(host string, allowedHosts []string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if host == "" {
		return false
	}
	for _, raw := range allowedHosts {
		for _, entry := range strings.Split(raw, ",") {
			entry = normalizeAllowedBrowserHost(entry)
			if entry == "" {
				continue
			}
			if entry == "*" || entry == host {
				return true
			}
			if strings.HasPrefix(entry, "*.") {
				suffix := strings.TrimPrefix(entry, "*")
				if strings.HasSuffix(host, suffix) && host != strings.TrimPrefix(suffix, ".") {
					return true
				}
			}
		}
	}
	return false
}

func normalizeAllowedBrowserHost(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		if u, err := url.Parse(value); err == nil {
			value = u.Hostname()
		}
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	value = strings.TrimSuffix(value, ".")
	return value
}
