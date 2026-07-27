package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

var playwrightInstallOnce sync.Once
var playwrightInstallErr error

// isSafeBrowserURL validates that a URL won't navigate the browser to internal
// infrastructure (AUDIT SEC-06). Rejects loopback, RFC1918 private networks,
// link-local, and cloud metadata endpoints.
func isSafeBrowserURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	lower := strings.ToLower(host)

	// Block cloud metadata endpoints
	if lower == "169.254.169.254" || lower == "metadata.google.internal" || lower == "metadata" {
		return false
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return false
		}
		if ip.IsPrivate() {
			return false
		}
		return true
	}

	// Block hostnames that resolve to loopback
	if ips, err := net.LookupIP(host); err == nil {
		for _, resolved := range ips {
			if resolved.IsLoopback() || resolved.IsLinkLocalUnicast() || resolved.IsPrivate() {
				return false
			}
		}
	}

	return true
}

type PlaywrightRunner struct {
	VideoDir string
	llm      LLM
}

func NewPlaywrightRunner(videoDir string, llm LLM) *PlaywrightRunner {
	return &PlaywrightRunner{VideoDir: videoDir, llm: llm}
}

type BrowserAction struct {
	Action   string `json:"action"`
	URL      string `json:"url,omitempty"`
	Selector string `json:"selector,omitempty"`
	Value    string `json:"value,omitempty"`
	Y        int    `json:"y,omitempty"`
	Ms       int    `json:"ms,omitempty"`
}

func (r *PlaywrightRunner) Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error) {
	// Install Playwright once per process lifetime (AUDIT P-01)
	playwrightInstallOnce.Do(func() {
		playwrightInstallErr = playwright.Install()
	})
	if playwrightInstallErr != nil {
		return nil, fmt.Errorf("playwright not available: %w", playwrightInstallErr)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	// Setup context with video recording
	bCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: r.VideoDir,
			Size: &playwright.Size{
				Width:  1280,
				Height: 720,
			},
		},
		Viewport: &playwright.Size{Width: 1280, Height: 720},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create context: %w", err)
	}
	defer bCtx.Close()

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	time.Sleep(1 * time.Second)

	// Execute dynamic scripts
	for _, tf := range testFiles {
		var actions []BrowserAction
		if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
			// If not JSON or empty, ignore
			continue
		}

		for i := 0; i < len(actions); i++ {
			a := actions[i]
			var err error
			switch a.Action {
			case "goto":
				if !isSafeBrowserURL(a.URL) {
					return nil, fmt.Errorf("browser egress blocked: unsafe URL %q", a.URL)
				}
				_, err = page.Goto(a.URL)
				page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State: playwright.LoadStateNetworkidle,
				})
			case "fill":
				err = page.Locator(a.Selector).Fill(a.Value, playwright.LocatorFillOptions{Timeout: playwright.Float(3000)})
			case "click":
				err = page.Locator(a.Selector).First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
			case "scroll":
				_, err = page.Evaluate(fmt.Sprintf("window.scrollBy(0, %d)", a.Y))
			case "wait":
				time.Sleep(time.Duration(a.Ms) * time.Millisecond)
			}

			if err != nil && r.llm != nil {
				// SELF-HEALING LOOP
				for attempt := 1; attempt <= 3; attempt++ {
					// Extract a simplified version of the DOM to help the AI find the right element
					domSnapshot, _ := page.Evaluate(`() => {
						let result = "";
						document.querySelectorAll("button, input, a, select, [role='button']").forEach(el => {
							if (el.offsetParent !== null) { // only visible elements
								result += el.tagName + (el.id ? "#"+el.id : "") + (el.className ? "."+el.className.split(" ").join(".") : "") + " '" + (el.innerText || el.value || "").trim() + "'\n";
							}
						});
						return result;
					}`)

					// Capture screenshot for Vision AI
					var imageBase64 string
					if screenshot, errImg := page.Screenshot(playwright.PageScreenshotOptions{
						Type:    playwright.ScreenshotTypeJpeg,
						Quality: playwright.Int(50),
					}); errImg == nil {
						imageBase64 = base64.StdEncoding.EncodeToString(screenshot)
					}

					actionJSON, _ := json.Marshal(a)
					var newActionStr string
					var healErr error
					if imageBase64 != "" {
						newActionStr, healErr = r.llm.HealActionWithVision(ctx, string(actionJSON), fmt.Sprintf("%v", domSnapshot), err.Error(), imageBase64)
					} else {
						newActionStr, healErr = r.llm.HealAction(ctx, string(actionJSON), fmt.Sprintf("%v", domSnapshot), err.Error())
					}

					if healErr != nil {
						break // AI could not provide a healing suggestion
					}

					var newAction BrowserAction
					if json.Unmarshal([]byte(newActionStr), &newAction) == nil {
						a = newAction

						// Retry the healed action
						switch a.Action {
						case "goto":
							if !isSafeBrowserURL(a.URL) {
								return nil, fmt.Errorf("browser egress blocked: unsafe URL %q", a.URL)
							}
							_, err = page.Goto(a.URL)
						case "fill":
							err = page.Locator(a.Selector).Fill(a.Value, playwright.LocatorFillOptions{Timeout: playwright.Float(3000)})
						case "click":
							err = page.Locator(a.Selector).First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
						}

						if err == nil {
							break // Healing successful!
						}
					}
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

	// Close page and context to finalize video
	page.Close()
	bCtx.Close()

	// Find the recorded video path
	video := page.Video()
	var videoPath string
	if video != nil {
		if path, err := video.Path(); err == nil {
			videoPath = path
			// Convert absolute path to relative path for URL serving
			videoPath = "/videos/" + filepath.Base(videoPath)
		}
	}

	return &RunResult{
		Passed:    0,
		Failed:    0,
		Total:     1,
		Failures:  []Failure{},
		VideoPath: videoPath,
	}, nil
}
