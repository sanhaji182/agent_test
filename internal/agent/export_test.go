package agent

import (
	"strings"
	"testing"
)

func TestExportPlaywrightScript_BasicActions(t *testing.T) {
	tf := TestFile{
		Name:    "login_test.json",
		Content: `[{"action":"goto","url":"https://example.com"},{"action":"fill","selector":"#email","value":"user@test.com"},{"action":"click","selector":"button[type='submit']"},{"action":"assert","selector":".dashboard","assert":"visible"}]`,
	}

	code := ExportPlaywrightScript(tf, ExportOptions{AddWaits: true})

	checks := []string{
		"import { test, expect } from '@playwright/test';",
		"test('login_test', async ({ page }) => {",
		"await page.goto('https://example.com');",
		"await page.waitForLoadState('networkidle');",
		"await page.locator('#email').fill('user@test.com'",
		"await page.locator('button[type=\\'submit\\']').first().click(",
		"await expect(page.locator('.dashboard')).toBeVisible(",
		"});",
	}
	for _, want := range checks {
		if !strings.Contains(code, want) {
			t.Errorf("export missing %q\nGot:\n%s", want, code)
		}
	}
}

func TestExportPlaywrightScript_AllAssertTypes(t *testing.T) {
	tf := TestFile{
		Name: "asserts.json",
		Content: `[
			{"action":"assert","selector":"h1","assert":"visible"},
			{"action":"assert","selector":".modal","assert":"hidden"},
			{"action":"assert","selector":"p","assert":"text_contains","text":"hello"},
			{"action":"assert","assert":"url_contains","text":"/dashboard"},
			{"action":"assert","assert":"title_contains","text":"Home"},
			{"action":"assert","selector":"li","assert":"count","text":"5"},
			{"action":"assert","selector":"a","assert":"attribute","key":"href","text":"/home"}
		]`,
	}

	code := ExportPlaywrightScript(tf, ExportOptions{})

	checks := []string{
		"toBeVisible(",
		"toBeHidden(",
		"toContainText('hello'",
		"toHaveURL(/\\/dashboard/)",
		"toHaveTitle(/Home/)",
		"toHaveCount(5)",
		"toHaveAttribute('href', '/home')",
	}
	for _, want := range checks {
		if !strings.Contains(code, want) {
			t.Errorf("export missing %q\nGot:\n%s", want, code)
		}
	}
}

func TestExportPlaywrightScript_SpecialActions(t *testing.T) {
	tf := TestFile{
		Name: "special.json",
		Content: `[
			{"action":"hover","selector":".menu"},
			{"action":"press","selector":"input","key":"Enter"},
			{"action":"scroll","y":500},
			{"action":"wait","ms":2000},
			{"action":"screenshot"},
			{"action":"network_wait","network_url":"/api/data","ms":5000}
		]`,
	}

	code := ExportPlaywrightScript(tf, ExportOptions{})

	checks := []string{
		".hover(",
		".press('Enter')",
		"window.scrollBy(0, 500)",
		"waitForTimeout(2000)",
		"page.screenshot(",
		"waitForResponse(",
	}
	for _, want := range checks {
		if !strings.Contains(code, want) {
			t.Errorf("export missing %q\nGot:\n%s", want, code)
		}
	}
}

func TestExportAllScriptsForTarget_MultiFramework(t *testing.T) {
	files := []TestFile{{
		Name:    "login.json",
		Content: `[{"action":"goto","url":"https://example.com"},{"action":"fill","selector":"#email","value":"user@test.com"},{"action":"click","selector":"button"},{"action":"assert","selector":".dashboard","assert":"visible"}]`,
	}}

	cases := []struct {
		target    string
		file      string
		language  string
		framework string
		contains  []string
	}{
		{target: "playwright", file: "login.spec.ts", language: "typescript", framework: "Playwright", contains: []string{"@playwright/test", "page.goto"}},
		{target: "cypress", file: "login.cy.js", language: "javascript", framework: "Cypress", contains: []string{"cy.visit", "cy.get('#email'"}},
		{target: "puppeteer", file: "login.mjs", language: "javascript", framework: "Puppeteer", contains: []string{"import puppeteer", "page.goto"}},
		{target: "selenium", file: "login.py", language: "python", framework: "Selenium", contains: []string{"from selenium import webdriver", "driver.get"}},
		{target: "appium", file: "login.appium.mjs", language: "javascript", framework: "Appium WebdriverIO", contains: []string{"import { remote } from 'webdriverio'", "appium:automationName"}},
		{target: "webdriverio", file: "login.wdio.js", language: "javascript", framework: "WebdriverIO", contains: []string{"describe('login'", "await browser.url"}},
	}

	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			scripts, meta := ExportAllScriptsForTarget(files, tc.target, ExportOptions{AddWaits: true})
			if meta.Language != tc.language || meta.Framework != tc.framework {
				t.Fatalf("unexpected target metadata: %+v", meta)
			}
			code, ok := scripts[tc.file]
			if !ok {
				t.Fatalf("missing script %s in %+v", tc.file, scripts)
			}
			for _, want := range tc.contains {
				if !strings.Contains(code, want) {
					t.Fatalf("export for %s missing %q\nGot:\n%s", tc.target, want, code)
				}
			}
		})
	}
}

func TestResolveExportTarget_DefaultsToPlaywright(t *testing.T) {
	meta := ResolveExportTarget("unknown")
	if meta.Key != "playwright" || meta.Language != "typescript" || meta.Framework != "Playwright" {
		t.Fatalf("unexpected default export target: %+v", meta)
	}
}

func TestExportAllScripts(t *testing.T) {
	files := []TestFile{
		{Name: "a.json", Content: `[{"action":"goto","url":"https://a.com"}]`},
		{Name: "b.json", Content: `[{"action":"goto","url":"https://b.com"}]`},
	}
	result := ExportAllScripts(files, ExportOptions{})
	if len(result) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result))
	}
	if _, ok := result["a.spec.ts"]; !ok {
		t.Error("missing a.spec.ts")
	}
	if _, ok := result["b.spec.ts"]; !ok {
		t.Error("missing b.spec.ts")
	}
}

func TestExportPlaywrightScript_InvalidJSON(t *testing.T) {
	tf := TestFile{Name: "bad.json", Content: "not json"}
	code := ExportPlaywrightScript(tf, ExportOptions{})
	if !strings.Contains(code, "Error") {
		t.Error("expected error comment for invalid JSON")
	}
}

func TestEscapeJS(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello", "hello"},
		{"it's", "it\\'s"},
		{"back\\slash", "back\\\\slash"},
		{"line\nbreak", "line\\nbreak"},
	}
	for _, tc := range cases {
		if got := escapeJS(tc.in); got != tc.want {
			t.Errorf("escapeJS(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
