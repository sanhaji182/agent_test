package browser

import (
	"os"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestCDPSnapshotLiveSmoke(t *testing.T) {
	if os.Getenv("GOTEST_CDP_SNAPSHOT_SMOKE") != "1" {
		t.Skip("set GOTEST_CDP_SNAPSHOT_SMOKE=1 to run live CDP snapshot browser smoke test")
	}

	if err := playwright.Install(); err != nil {
		t.Fatalf("install Playwright: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("start Playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("launch Chromium: %v", err)
	}
	defer browser.Close()

	ctx, err := browser.NewContext()
	if err != nil {
		t.Fatalf("new context: %v", err)
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	// Navigate to a well-known page with predictable structure
	snapshot, err := SnapshotPage(t.Context(), page, "https://example.com")
	if err != nil {
		t.Fatalf("CDP snapshot: %v", err)
	}

	if len(snapshot.Elements) == 0 {
		t.Fatal("expected at least one element in snapshot")
	}

	// Verify heading exists (example.com has "Example Domain")
	heading := FindElementByTextCDP(snapshot, "Example")
	if heading == nil {
		t.Logf("elements found: %d", len(snapshot.Elements))
		for _, el := range snapshot.Elements {
			t.Logf("  [@%s] %s %q", el.Ref, el.Role, el.Text)
		}
		t.Fatal("expected heading with 'Example Domain' text")
	}
	if heading.Role != "heading" {
		t.Errorf("heading role = %q, want heading", heading.Role)
	}

	// Verify links exist
	links := FindElementsByRoleCDP(snapshot, "link")
	if len(links) == 0 {
		t.Fatal("expected at least one link on example.com")
	}

	// Verify prompt format is compact
	prompt := CDPSnapshotToPrompt(snapshot)
	if !strings.Contains(prompt, "https://example.com") {
		t.Error("prompt missing URL")
	}
	if !strings.Contains(prompt, "[@e") {
		t.Error("prompt missing element refs")
	}
	if len(prompt) > 5000 {
		t.Errorf("prompt too large: %d bytes (expected < 5000)", len(prompt))
	}
	t.Logf("snapshot elements: %d, prompt size: %d bytes", len(snapshot.Elements), len(prompt))

	// Verify refs are unique
	refs := make(map[string]bool)
	for _, el := range snapshot.Elements {
		if refs[el.Ref] {
			t.Errorf("duplicate ref: %s", el.Ref)
		}
		refs[el.Ref] = true
	}

	// Verify GetPageSnapshotFromPlaywright works separately
	snapshot2, err := GetPageSnapshotFromPlaywright(t.Context(), page)
	if err != nil {
		t.Fatalf("GetPageSnapshotFromPlaywright: %v", err)
	}
	if len(snapshot2.Elements) == 0 {
		t.Fatal("expected elements from GetPageSnapshotFromPlaywright")
	}

	// Verify FindElementCDP works
	el := FindElementCDP(snapshot, heading.Ref)
	if el == nil || el.Role != "heading" {
		t.Error("FindElementCDP failed for heading ref")
	}
}
