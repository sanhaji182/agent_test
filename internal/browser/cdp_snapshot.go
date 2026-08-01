package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/playwright-community/playwright-go"
)

// refCounter provides unique ref IDs across snapshots within a process.
// Reset on each new snapshot call for determinism within a session.
var refCounter atomic.Int64

// AccessibilityNode represents a single node from the Chrome accessibility tree
// as extracted via CDP's page.Evaluate().
type AccessibilityNode struct {
	Role        string              `json:"role"`
	Name        string              `json:"name,omitempty"`
	Value       string              `json:"value,omitempty"`
	Description string              `json:"description,omitempty"`
	HTMLTag     string              `json:"htmlTag,omitempty"`
	Children    []AccessibilityNode `json:"children,omitempty"`
	Interactive bool                `json:"interactive,omitempty"`
	Focused     bool                `json:"focused,omitempty"`
	Disabled    bool                `json:"disabled,omitempty"`
	Checked     *bool               `json:"checked,omitempty"`
	Selected    *bool               `json:"selected,omitempty"`
	Level       int                 `json:"level,omitempty"`
}

// CDPSnapshotResult holds the parsed accessibility tree from a Playwright page.
type CDPSnapshotResult struct {
	URL      string
	Elements []PageElement
	RawTree  []AccessibilityNode
}

// extractAccessibilityTreeJS is the JavaScript function injected into the browser
// via page.Evaluate() to extract the accessibility tree. It traverses the DOM
// using native accessibility APIs and returns a structured tree.
const extractAccessibilityTreeJS = `() => {
	function getRole(el) {
		var ariaRole = el.getAttribute('role');
		if (ariaRole) return ariaRole;
		var tag = el.tagName.toLowerCase();
		var type = (el.getAttribute('type') || '').toLowerCase();
		switch (tag) {
			case 'a': return el.href ? 'link' : 'text';
			case 'button': return 'button';
			case 'h1': case 'h2': case 'h3': case 'h4': case 'h5': case 'h6': return 'heading';
			case 'input':
				if (type === 'checkbox') return 'checkbox';
				if (type === 'radio') return 'radio';
				if (type === 'submit' || type === 'button') return 'button';
				if (type === 'search') return 'searchbox';
				return 'textbox';
			case 'textarea': return 'textbox';
			case 'select': return 'combobox';
			case 'img': return 'img';
			case 'nav': return 'navigation';
			case 'main': return 'main';
			case 'header': return 'banner';
			case 'footer': return 'contentinfo';
			case 'form': return 'form';
			case 'table': return 'table';
			case 'tr': return 'row';
			case 'td': case 'th': return 'cell';
			case 'ul': case 'ol': return 'list';
			case 'li': return 'listitem';
			case 'p': return 'paragraph';
			default: return tag;
		}
	}
	function isInteractive(el) {
		var tag = el.tagName.toLowerCase();
		var role = el.getAttribute('role');
		var itags = ['a','button','input','textarea','select','details','summary'];
		var iroles = ['button','link','textbox','checkbox','radio','combobox','menuitem','tab','switch','slider','searchbox'];
		if (itags.indexOf(tag) !== -1) return true;
		if (role && iroles.indexOf(role) !== -1) return true;
		if (el.hasAttribute('onclick') || el.hasAttribute('tabindex')) return true;
		if (el.hasAttribute('href')) return true;
		return false;
	}
	function getTextContent(el) {
		if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
			return el.value || el.placeholder || el.getAttribute('aria-label') || '';
		}
		if (el.tagName === 'IMG') {
			return el.alt || el.title || '';
		}
		var text = '';
		var children = el.childNodes;
		for (var i = 0; i < children.length; i++) {
			if (children[i].nodeType === 3) {
				text += children[i].textContent.trim() + ' ';
			}
		}
		text = text.trim();
		if (!text) {
			text = el.getAttribute('aria-label') || el.getAttribute('title') || '';
		}
		return text;
	}
	function traverse(el, depth) {
		if (depth > 15) return null;
		if (!el || !el.tagName) return null;
		var style = window.getComputedStyle(el);
		if (style.display === 'none' || style.visibility === 'hidden') return null;
		var role = getRole(el);
		var text = getTextContent(el);
		var interactive = isInteractive(el);
		var children = [];
		var childElements = el.children;
		for (var i = 0; i < childElements.length; i++) {
			var node = traverse(childElements[i], depth + 1);
			if (node) children.push(node);
		}
		if (!text && !interactive && children.length === 0 &&
			(role === 'div' || role === 'span' || role === 'section' || role === 'article')) {
			return null;
		}
		return {
			role: role,
			name: text,
			htmlTag: el.tagName.toLowerCase(),
			interactive: interactive,
			disabled: el.disabled || false,
			children: children
		};
	}
	var body = document.body;
	if (!body) return [];
	return [traverse(body, 0)].filter(Boolean);
}`

// GetPageSnapshotFromPlaywright extracts a compact accessibility tree snapshot
// from a Playwright page using CDP. This is the PRIMARY snapshot method —
// no external binary required.
func GetPageSnapshotFromPlaywright(ctx context.Context, page playwright.Page) (*CDPSnapshotResult, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	rawResult, err := page.Evaluate(extractAccessibilityTreeJS)
	if err != nil {
		return nil, fmt.Errorf("evaluate accessibility tree: %w", err)
	}

	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return nil, fmt.Errorf("marshal accessibility tree: %w", err)
	}

	var tree []AccessibilityNode
	if err := json.Unmarshal(jsonBytes, &tree); err != nil {
		return nil, fmt.Errorf("unmarshal accessibility tree: %w", err)
	}

	url := page.URL()
	elements := flattenAccessibilityTree(tree)

	slog.Debug("CDP snapshot extracted",
		"url", url,
		"tree_nodes", len(tree),
		"elements", len(elements),
	)

	return &CDPSnapshotResult{
		URL:      url,
		Elements: elements,
		RawTree:  tree,
	}, nil
}

// SnapshotPage navigates to a URL and returns a compact snapshot.
func SnapshotPage(ctx context.Context, page playwright.Page, url string) (*CDPSnapshotResult, error) {
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout: playwright.Float(15000),
	}); err != nil {
		return nil, fmt.Errorf("navigate to %s: %w", url, err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		slog.Warn("wait for network idle failed", "url", url, "error", err)
	}

	return GetPageSnapshotFromPlaywright(ctx, page)
}

// flattenAccessibilityTree converts the recursive accessibility tree into
// a flat list of PageElements with assigned refs.
func flattenAccessibilityTree(tree []AccessibilityNode) []PageElement {
	refCounter.Store(0)
	var elements []PageElement
	flattenNode(tree, 0, &elements)
	return elements
}

func flattenNode(nodes []AccessibilityNode, depth int, elements *[]PageElement) {
	for _, node := range nodes {
		refCounter.Add(1)
		ref := fmt.Sprintf("e%d", refCounter.Load())

		*elements = append(*elements, PageElement{
			Ref:     ref,
			Role:    node.Role,
			Text:    node.Name,
			Tag:     node.HTMLTag,
			Indent:  depth,
			RawLine: formatRawLine(node, depth),
		})

		if len(node.Children) > 0 {
			flattenNode(node.Children, depth+1, elements)
		}
	}
}

func formatRawLine(node AccessibilityNode, depth int) string {
	prefix := strings.Repeat("  ", depth)
	text := ""
	if node.Name != "" {
		text = fmt.Sprintf(" %q", truncateText(node.Name, 60))
	}
	disabled := ""
	if node.Disabled {
		disabled = " [disabled]"
	}
	return fmt.Sprintf("%s- %s%s%s", prefix, node.Role, text, disabled)
}

// CDPSnapshotToPrompt converts a CDPSnapshotResult into a compact text
// representation optimized for LLM prompts.
func CDPSnapshotToPrompt(result *CDPSnapshotResult) string {
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

// FindElementCDP finds a PageElement by ref in a CDP snapshot result.
func FindElementCDP(result *CDPSnapshotResult, ref string) *PageElement {
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

// FindElementsByRoleCDP returns all elements matching a given role.
func FindElementsByRoleCDP(result *CDPSnapshotResult, role string) []PageElement {
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

// FindElementByTextCDP searches for an element containing the given text (case-insensitive).
func FindElementByTextCDP(result *CDPSnapshotResult, text string) *PageElement {
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

// GetCompactSnapshot is the HIGH-LEVEL API that uses the best available method
// to get a compact page snapshot:
//  1. If a Playwright page is provided → use CDP (fastest, no external deps)
//  2. If agent-browser is installed → fall back to CLI
func GetCompactSnapshot(ctx context.Context, page playwright.Page, url string) (*SnapshotResult, error) {
	// Primary: CDP via Playwright
	if page != nil {
		cdpResult, err := SnapshotPage(ctx, page, url)
		if err == nil {
			return &SnapshotResult{
				URL:        cdpResult.URL,
				Elements:   cdpResult.Elements,
				RawOutput:  CDPSnapshotToPrompt(cdpResult),
				SnapshotAt: time.Now(),
			}, nil
		}
		slog.Warn("CDP snapshot failed, trying fallback", "url", url, "error", err)
	}

	// Fallback: agent-browser CLI (if installed)
	if IsAvailable() {
		return GetPageSnapshot(ctx, url)
	}

	return nil, fmt.Errorf("no snapshot method available: CDP page is nil and agent-browser is not installed")
}
