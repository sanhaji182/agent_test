// Package browser provides an integration layer for agent-browser,
// an AI-optimized browser automation CLI from Vercel Labs.
//
// agent-browser returns compact accessibility tree snapshots with ref-based
// element identifiers (~200-400 tokens vs ~3000-5000 for full DOM), making
// it ideal for LLM-driven test generation and page analysis.
//
// This package is an auxiliary tool — it does NOT replace Playwright for
// test execution. Playwright remains the primary execution engine for
// video recording, network interception, and multi-context workflows.
//
// Integration pattern:
//  1. Use GetPageSnapshot() for LLM-efficient page analysis
//  2. Use ExecuteAction() for ref-based element interaction
//  3. Fall back to Playwright when agent-browser is unavailable
//
// Version pinning:
//
//	agent-browser is an external CLI binary. Pin the expected version
//	in MinimumVersion and install via scripts/install-agent-browser.sh.
package browser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MinimumVersion is the minimum supported agent-browser version.
// Update this when adopting features from newer releases.
// If the installed version is older, IsAvailable() returns false
// and all calls fall through to the Playwright fallback path.
const MinimumVersion = "0.0.0" // accept any installed version for now

// PageElement represents a single interactive element from an agent-browser
// accessibility snapshot. Refs are deterministic within a snapshot — the same
// page state always produces the same refs.
type PageElement struct {
	Ref     string // e.g. "e1", "e2" — unique within a snapshot
	Role    string // e.g. "heading", "link", "button", "input"
	Text    string // visible text content
	Tag     string // HTML tag if identifiable (input, button, a, etc.)
	Indent  int    // nesting depth in the accessibility tree
	RawLine string // original snapshot line for debugging
}

// SnapshotResult holds a parsed agent-browser snapshot.
type SnapshotResult struct {
	URL        string
	Elements   []PageElement
	RawOutput  string
	SnapshotAt time.Time
}

// AvailabilityInfo reports whether agent-browser is installed and usable.
type AvailabilityInfo struct {
	Available bool
	Version   string
	Path      string
	Error     string
}

var (
	ErrNotAvailable  = errors.New("agent-browser is not installed or not compatible")
	ErrCommandFailed = errors.New("agent-browser command failed")
	ErrParseFailed   = errors.New("failed to parse agent-browser output")
)

// cachedAvailability caches the availability check so we only run
// `agent-browser --version` once per process lifetime.
var (
	cachedAvail     *AvailabilityInfo
	cachedAvailOnce sync.Once
)

// refPattern matches [ref=e1] or [ref=abc123] in snapshot output.
var refPattern = regexp.MustCompile(`\[ref=([a-zA-Z0-9_-]+)\]`)

// rolePattern matches the leading role name: "- heading", "- link", "- button", etc.
var rolePattern = regexp.MustCompile(`^(\s*)- (\w+)`)

// textPattern matches quoted text: "Some text here"
var textPattern = regexp.MustCompile(`"([^"]*)"`)

// CheckAvailability verifies that agent-browser is installed and returns
// version info. Results are cached for the process lifetime.
func CheckAvailability() AvailabilityInfo {
	cachedAvailOnce.Do(func() {
		info := checkAvailabilityUncached()
		cachedAvail = &info
	})
	return *cachedAvail
}

// IsAvailable returns true if agent-browser can be used.
func IsAvailable() bool {
	return CheckAvailability().Available
}

// GetPageSnapshot opens a URL via agent-browser and returns a compact
// accessibility tree snapshot. This is the primary integration point
// for LLM-efficient page analysis.
//
// If agent-browser is not available, returns ErrNotAvailable.
// Callers should fall back to Playwright in that case.
func GetPageSnapshot(ctx context.Context, url string) (*SnapshotResult, error) {
	avail := CheckAvailability()
	if !avail.Available {
		return nil, fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	args := []string{"snapshot", "-i"}
	if url != "" {
		args = append(args, url)
	}

	out, err := runCommand(ctx, "agent-browser", args...)
	if err != nil {
		return nil, fmt.Errorf("%w: snapshot command failed: %v", ErrCommandFailed, err)
	}

	elements := ParseSnapshotOutput(string(out))
	if len(elements) == 0 {
		slog.Warn("agent-browser snapshot returned no elements", "url", url)
	}

	return &SnapshotResult{
		URL:        url,
		Elements:   elements,
		RawOutput:  string(out),
		SnapshotAt: time.Now(),
	}, nil
}

// OpenPage navigates to a URL via agent-browser.
// Does not return a snapshot — use GetPageSnapshot for that.
func OpenPage(ctx context.Context, url string) error {
	avail := CheckAvailability()
	if !avail.Available {
		return fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	_, err := runCommand(ctx, "agent-browser", "open", url)
	if err != nil {
		return fmt.Errorf("%w: open command failed: %v", ErrCommandFailed, err)
	}
	return nil
}

// ClickRef clicks an element by its ref identifier.
func ClickRef(ctx context.Context, ref string) error {
	avail := CheckAvailability()
	if !avail.Available {
		return fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	_, err := runCommand(ctx, "agent-browser", "click", "@"+ref)
	if err != nil {
		return fmt.Errorf("%w: click command failed: %v", ErrCommandFailed, err)
	}
	return nil
}

// FillRef fills an input element by its ref with the given value.
func FillRef(ctx context.Context, ref string, value string) error {
	avail := CheckAvailability()
	if !avail.Available {
		return fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	_, err := runCommand(ctx, "agent-browser", "fill", "@"+ref, value)
	if err != nil {
		return fmt.Errorf("%w: fill command failed: %v", ErrCommandFailed, err)
	}
	return nil
}

// Screenshot takes a screenshot via agent-browser and saves it to the given path.
func Screenshot(ctx context.Context, path string) error {
	avail := CheckAvailability()
	if !avail.Available {
		return fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	_, err := runCommand(ctx, "agent-browser", "screenshot", path)
	if err != nil {
		return fmt.Errorf("%w: screenshot command failed: %v", ErrCommandFailed, err)
	}
	return nil
}

// Close closes the current browser session.
func Close(ctx context.Context) error {
	avail := CheckAvailability()
	if !avail.Available {
		return fmt.Errorf("%w: %s", ErrNotAvailable, avail.Error)
	}

	_, err := runCommand(ctx, "agent-browser", "close")
	if err != nil {
		return fmt.Errorf("%w: close command failed: %v", ErrCommandFailed, err)
	}
	return nil
}

// ParseSnapshotOutput parses the compact text output from agent-browser
// snapshot into structured PageElements. This function is pure parsing
// with no side effects — safe to test with fixture strings.
//
// Expected format:
//
//   - heading "Example Domain" [ref=e1]
//   - paragraph "Some description..." [ref=e2]
//   - link "More information..." [ref=e3]
//   - input "Email" [ref=e4]
//   - button "Submit" [ref=e5]
func ParseSnapshotOutput(output string) []PageElement {
	var elements []PageElement

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Extract ref — lines without refs are not interactive elements
		refMatch := refPattern.FindStringSubmatch(trimmed)
		if refMatch == nil {
			continue
		}

		// Extract role from "- role" pattern
		role := ""
		roleMatch := rolePattern.FindStringSubmatch(line)
		if len(roleMatch) >= 3 {
			role = roleMatch[2]
		}

		// Extract quoted text
		text := ""
		textMatch := textPattern.FindStringSubmatch(trimmed)
		if len(textMatch) >= 2 {
			text = textMatch[1]
		}

		// Calculate indent depth
		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}

		elements = append(elements, PageElement{
			Ref:     refMatch[1],
			Role:    role,
			Text:    text,
			Tag:     inferTag(role, text),
			Indent:  indent,
			RawLine: trimmed,
		})
	}

	return elements
}

// SnapshotToPrompt converts a SnapshotResult into a compact text
// representation suitable for LLM prompts. This is the primary value
// proposition of the agent-browser integration — ~200-400 tokens vs
// ~3000-5000 for full DOM.
func SnapshotToPrompt(result *SnapshotResult) string {
	if result == nil || len(result.Elements) == 0 {
		return "(empty page snapshot)"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Page: %s\n", result.URL))
	sb.WriteString(fmt.Sprintf("Elements: %d\n\n", len(result.Elements)))

	for _, el := range result.Elements {
		sb.WriteString(fmt.Sprintf("[@%s] %s", el.Ref, el.Role))
		if el.Text != "" {
			sb.WriteString(fmt.Sprintf(" \"%s\"", truncateText(el.Text, 80)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FindElement finds a PageElement by ref in a snapshot result.
func FindElement(result *SnapshotResult, ref string) *PageElement {
	if result == nil {
		return nil
	}
	for i := range result.Elements {
		if result.Elements[i].Ref == ref {
			return &result.Elements[i]
		}
	}
	return nil
}

// FindElementsByRole returns all elements matching a given role.
func FindElementsByRole(result *SnapshotResult, role string) []PageElement {
	if result == nil {
		return nil
	}
	var matches []PageElement
	for _, el := range result.Elements {
		if strings.EqualFold(el.Role, role) {
			matches = append(matches, el)
		}
	}
	return matches
}

// FindElementByText searches for an element containing the given text (case-insensitive).
func FindElementByText(result *SnapshotResult, text string) *PageElement {
	if result == nil {
		return nil
	}
	lower := strings.ToLower(text)
	for i := range result.Elements {
		if strings.Contains(strings.ToLower(result.Elements[i].Text), lower) {
			return &result.Elements[i]
		}
	}
	return nil
}

// checkAvailabilityUncached performs the actual availability check.
func checkAvailabilityUncached() AvailabilityInfo {
	path, err := exec.LookPath("agent-browser")
	if err != nil {
		return AvailabilityInfo{
			Available: false,
			Error:     "agent-browser not found in PATH",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := runCommand(ctx, "agent-browser", "--version")
	if err != nil {
		return AvailabilityInfo{
			Available: false,
			Path:      path,
			Error:     fmt.Sprintf("version check failed: %v", err),
		}
	}

	version := strings.TrimSpace(string(out))

	return AvailabilityInfo{
		Available: true,
		Version:   version,
		Path:      path,
	}
}

// runCommand executes an external command with timeout and returns its output.
func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	// Default timeout if not provided in context
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out, fmt.Errorf("command %s timed out: %w", name, ctx.Err())
		}
		return out, fmt.Errorf("command %s failed: %w\noutput: %s", name, err, string(out))
	}
	return out, nil
}

// inferTag attempts to map an accessibility role to an HTML tag.
func inferTag(role, text string) string {
	switch strings.ToLower(role) {
	case "button":
		return "button"
	case "link":
		return "a"
	case "heading":
		return "h1-h6"
	case "input", "textbox", "searchbox":
		return "input"
	case "checkbox":
		return "input[type=checkbox]"
	case "radio":
		return "input[type=radio]"
	case "select", "listbox", "combobox":
		return "select"
	case "textarea":
		return "textarea"
	case "img", "image":
		return "img"
	case "paragraph":
		return "p"
	case "list":
		return "ul/ol"
	case "listitem":
		return "li"
	case "navigation":
		return "nav"
	case "banner":
		return "header"
	case "contentinfo":
		return "footer"
	case "main":
		return "main"
	case "form":
		return "form"
	case "table":
		return "table"
	case "row":
		return "tr"
	case "cell":
		return "td"
	default:
		return ""
	}
}

// truncateText truncates a string to maxLen characters, adding "..." if truncated.
func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
