# agent-browser Integration

## Overview

GoTest Agent integrates with [agent-browser](https://agent-browser.dev/) from Vercel Labs as an **auxiliary** browser automation tool optimized for LLM-driven page analysis.

### What it provides

- **Compact accessibility tree snapshots** — ~200-400 tokens vs ~3000-5000 for full DOM
- **Ref-based element selection** — deterministic refs like `@e1`, `@e2` that survive page state changes
- **50+ CLI commands** — navigation, forms, screenshots, network, storage, tabs, frames, debugging
- **Stateful sessions** — persistent daemon between commands for faster interactions

### What it does NOT replace

agent-browser is an auxiliary tool. **Playwright remains the primary execution engine** for:

- Test execution with video recording
- Network interception and mocking
- Multi-context browser workflows
- Browser egress guards and security
- Export code generation (Playwright, Cypress, Selenium, etc.)

## Installation

```bash
# macOS (recommended)
brew install agent-browser

# Cross-platform
npm install -g agent-browser

# Or use the project script
./scripts/install-agent-browser.sh

# Download Chrome for agent-browser (first time only)
agent-browser install
```

## Usage in GoTest Agent

### Page snapshot for LLM analysis

```go
import "github.com/go-go-golems/gotest-agent/internal/browser"

// Check if agent-browser is available
if browser.IsAvailable() {
    // Get compact snapshot for LLM
    result, err := browser.GetPageSnapshot(ctx, "https://example.com")
    if err == nil {
        // Convert to LLM-optimized prompt (~200-400 tokens)
        prompt := browser.SnapshotToPrompt(result)
        // Send prompt to LLM for test generation
    }
} else {
    // Fallback: use Playwright for page analysis
}
```

### Ref-based interaction

```go
result, _ := browser.GetPageSnapshot(ctx, "https://example.com")

// Find element by text
submitBtn := browser.FindElementByText(result, "Submit")
if submitBtn != nil {
    browser.ClickRef(ctx, submitBtn.Ref)
}

// Find all inputs
inputs := browser.FindElementsByRole(result, "input")
for _, input := range inputs {
    browser.FillRef(ctx, input.Ref, "test value")
}
```

### Search snapshot elements

```go
result, _ := browser.GetPageSnapshot(ctx, "https://example.com")

// Find by ref
el := browser.FindElement(result, "e3")

// Find by role
buttons := browser.FindElementsByRole(result, "button")

// Find by text content
login := browser.FindElementByText(result, "Sign In")
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│                GoTest Agent                      │
│                                                  │
│  ┌──────────────┐     ┌──────────────────────┐  │
│  │   LLM Engine  │────▶│  browser.GetPageSnapshot │  │
│  │  (Test Gen)   │     │     (agent-browser)    │  │
│  └──────────────┘     └──────────┬───────────┘  │
│                                  │               │
│  ┌──────────────┐               │               │
│  │   Playwright  │◀──── fallback ─┘               │
│  │  (Execution)  │                                │
│  └──────────────┘                                │
└─────────────────────────────────────────────────┘
```

## Version Management

agent-browser is pinned via `MinimumVersion` constant in `internal/browser/agent_browser.go`.

Update process:
1. Check [agent-browser releases](https://github.com/vercel-labs/agent-browser/releases)
2. Update `MinimumVersion` if new features are needed
3. Update `scripts/install-agent-browser.sh` with pinned version
4. Run `./scripts/install-agent-browser.sh <version>`
5. Run `go test ./internal/browser -v` to verify compatibility

## Configuration

No additional configuration needed. agent-browser uses the same Chrome profile as Playwright.

Environment variables (optional):
- `AGENT_BROWSER_IDLE_TIMEOUT_MS` — daemon idle timeout (default: 1 hour)
- `AGENT_BROWSER_HEADLESS` — run in headless mode (default: true)

## Troubleshooting

### "agent-browser not found in PATH"

```bash
# Check if installed
which agent-browser

# If installed via npm, check global bin
npm config get prefix
# Add to PATH: export PATH="$(npm config get prefix)/bin:$PATH"

# Or install via brew (recommended for macOS)
brew install agent-browser
```

### "Chrome not found"

```bash
# Download Chrome for agent-browser
agent-browser install
```

### "Version too old"

```bash
# Update agent-browser
brew upgrade agent-browser
# or
npm update -g agent-browser
```

## References

- [agent-browser Documentation](https://agent-browser.dev/)
- [agent-browser GitHub](https://github.com/vercel-labs/agent-browser)
- [Vercel Labs](https://vercel.com)
