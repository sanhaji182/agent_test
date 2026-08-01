package browser

import (
	"sync"
	"testing"
)

func TestParseSnapshotOutput_StandardFormat(t *testing.T) {
	output := `- heading "Example Domain" [ref=e1]
- paragraph "This domain is for use in illustrative examples in documents." [ref=e2]
- link "More information..." [ref=e3]
- input "Email" [ref=e4]
- button "Submit" [ref=e5]
`

	elements := ParseSnapshotOutput(output)
	if len(elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(elements))
	}

	cases := []struct {
		ref  string
		role string
		text string
		tag  string
	}{
		{"e1", "heading", "Example Domain", "h1-h6"},
		{"e2", "paragraph", "This domain is for use in illustrative examples in documents.", "p"},
		{"e3", "link", "More information...", "a"},
		{"e4", "input", "Email", "input"},
		{"e5", "button", "Submit", "button"},
	}

	for i, tc := range cases {
		el := elements[i]
		if el.Ref != tc.ref {
			t.Errorf("element %d: ref=%q, want %q", i, el.Ref, tc.ref)
		}
		if el.Role != tc.role {
			t.Errorf("element %d: role=%q, want %q", i, el.Role, tc.role)
		}
		if el.Text != tc.text {
			t.Errorf("element %d: text=%q, want %q", i, el.Text, tc.text)
		}
		if el.Tag != tc.tag {
			t.Errorf("element %d: tag=%q, want %q", i, el.Tag, tc.tag)
		}
	}
}

func TestParseSnapshotOutput_NestedElements(t *testing.T) {
	output := `- heading "Test Page" [ref=e1]
  - link "Home" [ref=e2]
  - link "About" [ref=e3]
- form "Login" [ref=e4]
  - input "Email" [ref=e5]
  - input "Password" [ref=e6]
  - button "Sign In" [ref=e7]
`

	elements := ParseSnapshotOutput(output)
	if len(elements) != 7 {
		t.Fatalf("expected 7 elements, got %d", len(elements))
	}

	// Check indent levels
	if elements[0].Indent != 0 {
		t.Errorf("heading indent=%d, want 0", elements[0].Indent)
	}
	if elements[1].Indent == 0 {
		t.Error("nested link should have non-zero indent")
	}
	if elements[4].Indent == 0 {
		t.Error("nested input should have non-zero indent")
	}
}

func TestParseSnapshotOutput_LinesWithoutRefs(t *testing.T) {
	output := `Some introductory text without a ref
- heading "Page Title" [ref=e1]
Comment line without ref
- button "OK" [ref=e2]
`

	elements := ParseSnapshotOutput(output)
	if len(elements) != 2 {
		t.Fatalf("expected 2 elements (lines without refs skipped), got %d", len(elements))
	}
	if elements[0].Ref != "e1" || elements[1].Ref != "e2" {
		t.Errorf("unexpected refs: %q, %q", elements[0].Ref, elements[1].Ref)
	}
}

func TestParseSnapshotOutput_EmptyAndWhitespace(t *testing.T) {
	elements := ParseSnapshotOutput("")
	if len(elements) != 0 {
		t.Errorf("expected 0 elements for empty input, got %d", len(elements))
	}

	elements = ParseSnapshotOutput("   \n\n  \n")
	if len(elements) != 0 {
		t.Errorf("expected 0 elements for whitespace input, got %d", len(elements))
	}
}

func TestParseSnapshotOutput_ComplexRefs(t *testing.T) {
	output := `- heading "Title" [ref=e1]
- button "OK" [ref=btn-submit-1]
- input "Name" [ref=form_name_input]
`

	elements := ParseSnapshotOutput(output)
	if len(elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elements))
	}

	expectedRefs := []string{"e1", "btn-submit-1", "form_name_input"}
	for i, want := range expectedRefs {
		if elements[i].Ref != want {
			t.Errorf("element %d: ref=%q, want %q", i, elements[i].Ref, want)
		}
	}
}

func TestParseSnapshotOutput_QuotedTextEdgeCases(t *testing.T) {
	output := `- button "" [ref=e1]
- heading "Text with 'single quotes'" [ref=e2]
- link "Multi line" [ref=e3]
`

	elements := ParseSnapshotOutput(output)
	if len(elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elements))
	}

	// Empty text
	if elements[0].Text != "" {
		t.Errorf("expected empty text, got %q", elements[0].Text)
	}
	// Single quotes inside double quotes
	if elements[1].Text != "Text with 'single quotes'" {
		t.Errorf("unexpected text: %q", elements[1].Text)
	}
}

func TestSnapshotToPrompt(t *testing.T) {
	result := &SnapshotResult{
		URL: "https://example.com",
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Welcome"},
			{Ref: "e2", Role: "button", Text: "Get Started"},
		},
	}

	prompt := SnapshotToPrompt(result)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !containsSubstring(prompt, "Page: https://example.com") {
		t.Errorf("prompt missing URL: %s", prompt)
	}
	if !containsSubstring(prompt, "[@e1]") {
		t.Errorf("prompt missing ref e1: %s", prompt)
	}
	if !containsSubstring(prompt, "[@e2]") {
		t.Errorf("prompt missing ref e2: %s", prompt)
	}
	if !containsSubstring(prompt, "\"Welcome\"") {
		t.Errorf("prompt missing text: %s", prompt)
	}
}

func TestSnapshotToPrompt_EmptyResult(t *testing.T) {
	prompt := SnapshotToPrompt(nil)
	if prompt != "(empty page snapshot)" {
		t.Errorf("expected empty page snapshot, got %q", prompt)
	}

	prompt = SnapshotToPrompt(&SnapshotResult{URL: "https://example.com"})
	if prompt != "(empty page snapshot)" {
		t.Errorf("expected empty page snapshot for no elements, got %q", prompt)
	}
}

func TestFindElement(t *testing.T) {
	result := &SnapshotResult{
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Title"},
			{Ref: "e2", Role: "button", Text: "Submit"},
		},
	}

	el := FindElement(result, "e2")
	if el == nil {
		t.Fatal("expected to find e2")
	}
	if el.Role != "button" {
		t.Errorf("role=%q, want button", el.Role)
	}

	el = FindElement(result, "e99")
	if el != nil {
		t.Error("expected nil for nonexistent ref")
	}

	el = FindElement(nil, "e1")
	if el != nil {
		t.Error("expected nil for nil result")
	}
}

func TestFindElementsByRole(t *testing.T) {
	result := &SnapshotResult{
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Title"},
			{Ref: "e2", Role: "button", Text: "Submit"},
			{Ref: "e3", Role: "button", Text: "Cancel"},
			{Ref: "e4", Role: "link", Text: "Home"},
		},
	}

	buttons := FindElementsByRole(result, "button")
	if len(buttons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(buttons))
	}
	if buttons[0].Ref != "e2" || buttons[1].Ref != "e3" {
		t.Errorf("unexpected button refs: %q, %q", buttons[0].Ref, buttons[1].Ref)
	}

	empty := FindElementsByRole(result, "checkbox")
	if len(empty) != 0 {
		t.Errorf("expected 0 checkboxes, got %d", len(empty))
	}
}

func TestFindElementByText(t *testing.T) {
	result := &SnapshotResult{
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Welcome to Example"},
			{Ref: "e2", Role: "button", Text: "Get Started"},
		},
	}

	el := FindElementByText(result, "example")
	if el == nil {
		t.Fatal("expected to find element with 'example'")
	}
	if el.Ref != "e1" {
		t.Errorf("ref=%q, want e1", el.Ref)
	}

	el = FindElementByText(result, "nonexistent")
	if el != nil {
		t.Error("expected nil for nonexistent text")
	}

	el = FindElementByText(nil, "anything")
	if el != nil {
		t.Error("expected nil for nil result")
	}
}

func TestInferTag(t *testing.T) {
	cases := []struct {
		role string
		tag  string
	}{
		{"button", "button"},
		{"link", "a"},
		{"heading", "h1-h6"},
		{"input", "input"},
		{"textbox", "input"},
		{"checkbox", "input[type=checkbox]"},
		{"select", "select"},
		{"textarea", "textarea"},
		{"navigation", "nav"},
		{"banner", "header"},
		{"main", "main"},
		{"unknown_role", ""},
		{"", ""},
	}

	for _, tc := range cases {
		got := inferTag(tc.role, "")
		if got != tc.tag {
			t.Errorf("inferTag(%q)=%q, want %q", tc.role, got, tc.tag)
		}
	}
}

func TestTruncateText(t *testing.T) {
	cases := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 11, "exactly10!"},
		{"this is a very long text that should be truncated", 20, "this is a very lo..."},
		{"", 10, ""},
	}

	for _, tc := range cases {
		got := truncateText(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncateText(%q, %d)=%q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}

func TestCheckAvailability_WhenNotInstalled(t *testing.T) {
	// Reset cached availability for test
	cachedAvail = nil
	cachedAvailOnce = sync.Once{}

	// This test verifies graceful handling when agent-browser is not in PATH.
	// In CI where agent-browser is not installed, this should return false.
	info := CheckAvailability()

	if info.Available {
		// If agent-browser happens to be installed, that's fine too.
		t.Logf("agent-browser is installed: version=%s, path=%s", info.Version, info.Path)
	} else {
		if info.Error == "" {
			t.Error("expected non-empty error when not available")
		}
		t.Logf("agent-browser not available (expected in CI): %s", info.Error)
	}
}

func TestIsAvailable_ConsistentWithCheck(t *testing.T) {
	// Reset for clean test
	cachedAvail = nil
	cachedAvailOnce = sync.Once{}

	info := CheckAvailability()
	avail := IsAvailable()

	if info.Available != avail {
		t.Errorf("IsAvailable()=%v, CheckAvailability().Available=%v", avail, info.Available)
	}

	// Second call should use cache
	info2 := CheckAvailability()
	if info.Available != info2.Available || info.Version != info2.Version {
		t.Error("caching should return consistent results")
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsString(s, sub))
}

func containsString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// CDP-specific tests that don't require a real browser

func TestFlattenAccessibilityTree_AssignsRefs(t *testing.T) {
	tree := []AccessibilityNode{
		{
			Role: "heading", Name: "Welcome", HTMLTag: "h1", Interactive: false,
			Children: []AccessibilityNode{
				{Role: "paragraph", Name: "Lorem ipsum", HTMLTag: "p", Interactive: false},
			},
		},
		{Role: "button", Name: "Get Started", HTMLTag: "button", Interactive: true},
	}

	elements := flattenAccessibilityTree(tree)
	if len(elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elements))
	}

	if elements[0].Ref != "e1" {
		t.Errorf("first ref = %q, want e1", elements[0].Ref)
	}
	if elements[1].Ref != "e2" {
		t.Errorf("second ref = %q, want e2", elements[1].Ref)
	}
	if elements[2].Ref != "e3" {
		t.Errorf("third ref = %q, want e3", elements[2].Ref)
	}

	if elements[0].Indent != 0 {
		t.Errorf("heading indent = %d, want 0", elements[0].Indent)
	}
	if elements[1].Indent != 1 {
		t.Errorf("paragraph indent = %d, want 1", elements[1].Indent)
	}
	if elements[1].Role != "paragraph" || elements[1].Text != "Lorem ipsum" {
		t.Errorf("paragraph role/text = %q/%q", elements[1].Role, elements[1].Text)
	}
}

func TestFlattenAccessibilityTree_EmptyNil(t *testing.T) {
	if len(flattenAccessibilityTree(nil)) != 0 {
		t.Error("expected empty for nil tree")
	}
	if len(flattenAccessibilityTree([]AccessibilityNode{})) != 0 {
		t.Error("expected empty for empty tree")
	}
}

func TestFlattenAccessibilityTree_RefDeterministic(t *testing.T) {
	tree := []AccessibilityNode{
		{Role: "heading", Name: "A", HTMLTag: "h1"},
		{Role: "button", Name: "B", HTMLTag: "button"},
	}

	first := flattenAccessibilityTree(tree)
	second := flattenAccessibilityTree(tree)

	if len(first) != len(second) {
		t.Fatal("flatten should be deterministic")
	}
	for i := range first {
		if first[i].Ref != second[i].Ref {
			t.Errorf("ref mismatch at %d: %q vs %q", i, first[i].Ref, second[i].Ref)
		}
	}
}

func TestFormatRawLine_Variants(t *testing.T) {
	node := AccessibilityNode{Role: "button", Name: "Submit", HTMLTag: "button"}
	got := formatRawLine(node, 0)
	if !containsString(got, "button") || !containsString(got, "Submit") {
		t.Errorf("formatRawLine = %q", got)
	}

	// Disabled node
	node.Disabled = true
	got = formatRawLine(node, 0)
	if !containsString(got, "disabled") {
		t.Errorf("disabled node should show [disabled]: %q", got)
	}

	// Node with depth
	node.Disabled = false
	got = formatRawLine(node, 2)
	if len(got) < 4 || got[:4] != "    " {
		t.Errorf("should have indent prefix: %q", got)
	}
}

func TestCDPSnapshotToPrompt_Format(t *testing.T) {
	result := &CDPSnapshotResult{
		URL: "https://example.com",
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Welcome"},
			{Ref: "e2", Role: "button", Text: "Get Started"},
		},
	}

	prompt := CDPSnapshotToPrompt(result)
	if !containsString(prompt, "https://example.com") {
		t.Errorf("missing URL in prompt: %s", prompt)
	}
	if !containsString(prompt, "[@e1]") || !containsString(prompt, "[@e2]") {
		t.Error("missing refs in prompt")
	}
	if !containsString(prompt, "Welcome") || !containsString(prompt, "Get Started") {
		t.Error("missing text in prompt")
	}
}

func TestCDPSnapshotToPrompt_Empty(t *testing.T) {
	if got := CDPSnapshotToPrompt(nil); got != "(empty page snapshot)" {
		t.Errorf("nil result = %q", got)
	}
	result := &CDPSnapshotResult{URL: "https://test"}
	if got := CDPSnapshotToPrompt(result); got != "(empty page snapshot)" {
		t.Errorf("no elements = %q", got)
	}
}

func TestFindElementCDP(t *testing.T) {
	result := &CDPSnapshotResult{
		Elements: []PageElement{
			{Ref: "e1", Role: "heading"},
			{Ref: "e2", Role: "button"},
		},
	}

	el := FindElementCDP(result, "e2")
	if el == nil || el.Role != "button" {
		t.Error("expected to find e2")
	}

	if el := FindElementCDP(result, "e99"); el != nil {
		t.Error("expected nil for nonexistent ref")
	}
	if el := FindElementCDP(nil, "e1"); el != nil {
		t.Error("expected nil for nil result")
	}
}

func TestFindElementByTextCDP(t *testing.T) {
	result := &CDPSnapshotResult{
		Elements: []PageElement{
			{Ref: "e1", Role: "heading", Text: "Welcome to Example"},
			{Ref: "e2", Role: "button", Text: "Get Started"},
		},
	}

	el := FindElementByTextCDP(result, "example")
	if el == nil || el.Ref != "e1" {
		t.Error("expected to find element with 'example'")
	}

	if el := FindElementByTextCDP(result, "nonexistent"); el != nil {
		t.Error("expected nil for nonexistent text")
	}
}
