package agent

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ExportPlaywrightScript converts a TestFile's action JSON into a runnable
// Playwright TypeScript test file. This lets users take AI-generated actions
// and run them independently in their own Playwright setup.
func ExportPlaywrightScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	testName := testNameFromFile(tf.Name)

	sb.WriteString(`import { test, expect } from '@playwright/test';

`)
	sb.WriteString(fmt.Sprintf("test('%s', async ({ page }) => {\n", testName))

	for _, a := range actions {
		sb.WriteString(actionToPlaywrightCode(a, options))
	}

	sb.WriteString("});\n")
	return sb.String()
}

// ExportCypressScript converts action JSON into a Cypress spec.
func ExportCypressScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	testName := testNameFromFile(tf.Name)
	sb.WriteString(fmt.Sprintf("describe('%s', () => {\n", testName))
	sb.WriteString(fmt.Sprintf("%sit('replays generated flow', () => {\n", exportIndent(options)))
	for _, a := range actions {
		sb.WriteString(actionToCypressCode(a, options, exportIndent(options)+exportIndent(options)))
	}
	sb.WriteString(fmt.Sprintf("%s});\n", exportIndent(options)))
	sb.WriteString("});\n")
	return sb.String()
}

// ExportPuppeteerScript converts action JSON into a standalone Puppeteer script.
func ExportPuppeteerScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	sb.WriteString(`import puppeteer from 'puppeteer';

const browser = await puppeteer.launch({ headless: true });
const page = await browser.newPage();

try {
`)
	for _, a := range actions {
		sb.WriteString(actionToPuppeteerCode(a, options, exportIndent(options)))
	}
	sb.WriteString(`} finally {
  await browser.close();
}
`)
	return sb.String()
}

// ExportSeleniumScript converts action JSON into a Python Selenium script.
func ExportSeleniumScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	timeout := exportTimeoutSeconds(options)
	sb.WriteString(fmt.Sprintf(`from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
import time


def by_css(selector):
    return (By.CSS_SELECTOR, selector)


driver = webdriver.Chrome()
wait = WebDriverWait(driver, %d)

try:
`, timeout))
	for _, a := range actions {
		sb.WriteString(actionToSeleniumCode(a, options, "    "))
	}
	sb.WriteString(`finally:
    driver.quit()
`)
	return sb.String()
}

// ExportAppiumScript converts browser actions into an Appium/WebdriverIO mobile-web spec.
// Native mobile element locators can be supplied in the same selector field; web selectors
// work for mobile browser sessions.
func ExportAppiumScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	testName := testNameFromFile(tf.Name)
	sb.WriteString(fmt.Sprintf(`import { remote } from 'webdriverio';

const driver = await remote({
  protocol: process.env.APPIUM_PROTOCOL || 'http',
  hostname: process.env.APPIUM_HOST || '127.0.0.1',
  port: Number(process.env.APPIUM_PORT || 4723),
  path: process.env.APPIUM_PATH || '/',
  capabilities: {
    platformName: process.env.APPIUM_PLATFORM || 'Android',
    'appium:automationName': process.env.APPIUM_AUTOMATION || 'UiAutomator2',
    browserName: process.env.APPIUM_BROWSER || 'Chrome',
    'appium:deviceName': process.env.APPIUM_DEVICE || 'Android Emulator'
  }
});

try {
  // %s
`, testName))
	for _, a := range actions {
		sb.WriteString(actionToWebdriverIOCode(a, options, "  ", true))
	}
	sb.WriteString(`} finally {
  await driver.deleteSession();
}
`)
	return sb.String()
}

// ExportWebdriverIOScript converts browser actions into a WebdriverIO spec.
// It can target desktop browsers, mobile browsers, Selenium Grid, or cloud providers
// using a standard WebdriverIO runner configuration.
func ExportWebdriverIOScript(tf TestFile, options ExportOptions) string {
	actions, err := parseExportActions(tf)
	if err != nil {
		return exportParseError(err)
	}

	var sb strings.Builder
	testName := testNameFromFile(tf.Name)
	sb.WriteString(fmt.Sprintf("describe('%s', () => {\n", testName))
	sb.WriteString("  it('replays generated flow', async () => {\n")
	for _, a := range actions {
		sb.WriteString(actionToWebdriverIOCode(a, options, "    ", false))
	}
	sb.WriteString("  });\n")
	sb.WriteString("});\n")
	return sb.String()
}

// ExportOptions configures code generation style.
type ExportOptions struct {
	Timeout    int    // default timeout in ms (default 5000)
	Language   string // "typescript" (default), "javascript", or target-specific language
	AddWaits   bool   // add waitForLoadState after navigation where supported
	IndentWith string // "  " (default) or "\t"
}

// ExportTarget describes a supported export target.
type ExportTarget struct {
	Key       string `json:"key"`
	Language  string `json:"language"`
	Framework string `json:"framework"`
	Ext       string `json:"-"`
}

// ResolveExportTarget normalizes aliases from the API/UI into a supported export target.
func ResolveExportTarget(target string) ExportTarget {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "cypress":
		return ExportTarget{Key: "cypress", Language: "javascript", Framework: "Cypress", Ext: ".cy.js"}
	case "puppeteer":
		return ExportTarget{Key: "puppeteer", Language: "javascript", Framework: "Puppeteer", Ext: ".mjs"}
	case "selenium", "selenium-python", "python":
		return ExportTarget{Key: "selenium", Language: "python", Framework: "Selenium", Ext: ".py"}
	case "appium", "mobile", "mobile-web":
		return ExportTarget{Key: "appium", Language: "javascript", Framework: "Appium WebdriverIO", Ext: ".appium.mjs"}
	case "webdriverio", "wdio", "desktop":
		return ExportTarget{Key: "webdriverio", Language: "javascript", Framework: "WebdriverIO", Ext: ".wdio.js"}
	default:
		return ExportTarget{Key: "playwright", Language: "typescript", Framework: "Playwright", Ext: ".spec.ts"}
	}
}

func actionToPlaywrightCode(a BrowserAction, opts ExportOptions) string {
	indent := exportIndent(opts)
	timeout := exportTimeoutMS(opts)

	switch a.Action {
	case "goto":
		code := fmt.Sprintf("%sawait page.goto('%s');\n", indent, escapeJS(a.URL))
		if opts.AddWaits {
			code += fmt.Sprintf("%sawait page.waitForLoadState('networkidle');\n", indent)
		}
		return code
	case "fill":
		return fmt.Sprintf("%sawait page.locator('%s').fill('%s', { timeout: %d });\n",
			indent, escapeJS(a.Selector), escapeJS(a.Value), timeout)
	case "click":
		return fmt.Sprintf("%sawait page.locator('%s').first().click({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "hover":
		return fmt.Sprintf("%sawait page.locator('%s').first().hover({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "press":
		return fmt.Sprintf("%sawait page.locator('%s').first().press('%s');\n",
			indent, escapeJS(a.Selector), escapeJS(a.Key))
	case "scroll":
		return fmt.Sprintf("%sawait page.evaluate(() => window.scrollBy(0, %d));\n",
			indent, a.Y)
	case "wait":
		return fmt.Sprintf("%sawait page.waitForTimeout(%d);\n", indent, a.Ms)
	case "screenshot":
		return fmt.Sprintf("%sawait page.screenshot({ path: 'screenshot.png', fullPage: true });\n", indent)
	case "network_wait":
		return fmt.Sprintf("%sawait page.waitForResponse(resp => resp.url().includes('%s'), { timeout: %d });\n",
			indent, escapeJS(a.NetworkURL), max(a.Ms, 10000))
	case "assert":
		return assertToPlaywrightCode(a, indent, timeout)
	default:
		return fmt.Sprintf("%s// Unknown action: %s\n", indent, a.Action)
	}
}

func actionToCypressCode(a BrowserAction, opts ExportOptions, indent string) string {
	timeout := exportTimeoutMS(opts)
	switch a.Action {
	case "goto":
		return fmt.Sprintf("%scy.visit('%s');\n", indent, escapeJS(a.URL))
	case "fill":
		return fmt.Sprintf("%scy.get('%s', { timeout: %d }).clear().type('%s');\n", indent, escapeJS(a.Selector), timeout, escapeJS(a.Value))
	case "click":
		return fmt.Sprintf("%scy.get('%s', { timeout: %d }).first().click();\n", indent, escapeJS(a.Selector), timeout)
	case "hover":
		return fmt.Sprintf("%scy.get('%s', { timeout: %d }).first().trigger('mouseover');\n", indent, escapeJS(a.Selector), timeout)
	case "press":
		return fmt.Sprintf("%scy.get('%s').first().type('{%s}');\n", indent, escapeJS(a.Selector), strings.ToLower(escapeJS(a.Key)))
	case "scroll":
		return fmt.Sprintf("%scy.window().then((win) => win.scrollBy(0, %d));\n", indent, a.Y)
	case "wait":
		return fmt.Sprintf("%scy.wait(%d);\n", indent, a.Ms)
	case "screenshot":
		return fmt.Sprintf("%scy.screenshot('screenshot');\n", indent)
	case "network_wait":
		return fmt.Sprintf("%scy.intercept('**%s**').as('networkWait');\n%scy.wait('@networkWait', { timeout: %d });\n", indent, escapeJS(a.NetworkURL), indent, max(a.Ms, 10000))
	case "assert":
		return assertToCypressCode(a, indent)
	default:
		return fmt.Sprintf("%s// Unknown action: %s\n", indent, a.Action)
	}
}

func actionToPuppeteerCode(a BrowserAction, opts ExportOptions, indent string) string {
	timeout := exportTimeoutMS(opts)
	switch a.Action {
	case "goto":
		waitUntil := "load"
		if opts.AddWaits {
			waitUntil = "networkidle0"
		}
		return fmt.Sprintf("%sawait page.goto('%s', { waitUntil: '%s', timeout: %d });\n", indent, escapeJS(a.URL), waitUntil, timeout)
	case "fill":
		return fmt.Sprintf("%sawait page.waitForSelector('%s', { timeout: %d });\n%sawait page.click('%s', { clickCount: 3 });\n%sawait page.type('%s', '%s');\n", indent, escapeJS(a.Selector), timeout, indent, escapeJS(a.Selector), indent, escapeJS(a.Selector), escapeJS(a.Value))
	case "click":
		return fmt.Sprintf("%sawait page.waitForSelector('%s', { timeout: %d });\n%sawait page.click('%s');\n", indent, escapeJS(a.Selector), timeout, indent, escapeJS(a.Selector))
	case "hover":
		return fmt.Sprintf("%sawait page.hover('%s');\n", indent, escapeJS(a.Selector))
	case "press":
		return fmt.Sprintf("%sawait page.keyboard.press('%s');\n", indent, escapeJS(a.Key))
	case "scroll":
		return fmt.Sprintf("%sawait page.evaluate(() => window.scrollBy(0, %d));\n", indent, a.Y)
	case "wait":
		return fmt.Sprintf("%sawait new Promise((resolve) => setTimeout(resolve, %d));\n", indent, a.Ms)
	case "screenshot":
		return fmt.Sprintf("%sawait page.screenshot({ path: 'screenshot.png', fullPage: true });\n", indent)
	case "network_wait":
		return fmt.Sprintf("%sawait page.waitForResponse((resp) => resp.url().includes('%s'), { timeout: %d });\n", indent, escapeJS(a.NetworkURL), max(a.Ms, 10000))
	case "assert":
		return assertToPuppeteerCode(a, indent)
	default:
		return fmt.Sprintf("%s// Unknown action: %s\n", indent, a.Action)
	}
}

func actionToSeleniumCode(a BrowserAction, opts ExportOptions, indent string) string {
	timeout := exportTimeoutSeconds(opts)
	switch a.Action {
	case "goto":
		return fmt.Sprintf("%sdriver.get('%s')\n", indent, escapePy(a.URL))
	case "fill":
		return fmt.Sprintf("%sel = wait.until(EC.presence_of_element_located(by_css('%s')))\n%sel.clear()\n%sel.send_keys('%s')\n", indent, escapePy(a.Selector), indent, indent, escapePy(a.Value))
	case "click":
		return fmt.Sprintf("%swait.until(EC.element_to_be_clickable(by_css('%s'))).click()\n", indent, escapePy(a.Selector))
	case "hover":
		return fmt.Sprintf("%s# Hover action requires ActionChains; selector: %s\n", indent, escapePy(a.Selector))
	case "press":
		return fmt.Sprintf("%swait.until(EC.presence_of_element_located(by_css('%s'))).send_keys(Keys.%s)\n", indent, escapePy(a.Selector), seleniumKey(a.Key))
	case "scroll":
		return fmt.Sprintf("%sdriver.execute_script('window.scrollBy(0, %d)')\n", indent, a.Y)
	case "wait":
		return fmt.Sprintf("%stime.sleep(%0.3f)\n", indent, float64(a.Ms)/1000)
	case "screenshot":
		return fmt.Sprintf("%sdriver.save_screenshot('screenshot.png')\n", indent)
	case "network_wait":
		return fmt.Sprintf("%s# Network waits are not natively supported by Selenium WebDriver. Waited URL pattern: %s\n%stime.sleep(%d)\n", indent, escapePy(a.NetworkURL), indent, max(a.Ms/1000, timeout))
	case "assert":
		return assertToSeleniumCode(a, indent)
	default:
		return fmt.Sprintf("%s# Unknown action: %s\n", indent, a.Action)
	}
}

func actionToWebdriverIOCode(a BrowserAction, opts ExportOptions, indent string, appium bool) string {
	timeout := exportTimeoutMS(opts)
	selector := escapeJS(a.Selector)
	client := "browser"
	subject := "$"
	if appium {
		client = "driver"
		subject = "driver.$"
	}
	if appium && a.Action == "assert" {
		return assertToAppiumCode(a, indent, timeout)
	}
	switch a.Action {
	case "goto":
		return fmt.Sprintf("%sawait %s.url('%s');\n", indent, client, escapeJS(a.URL))
	case "fill":
		return fmt.Sprintf("%sawait (%s('%s')).setValue('%s');\n", indent, subject, selector, escapeJS(a.Value))
	case "click":
		return fmt.Sprintf("%sawait (%s('%s')).click();\n", indent, subject, selector)
	case "hover":
		return fmt.Sprintf("%sawait (%s('%s')).moveTo();\n", indent, subject, selector)
	case "press":
		return fmt.Sprintf("%sawait %s.keys('%s');\n", indent, client, escapeJS(a.Key))
	case "scroll":
		return fmt.Sprintf("%sawait %s.execute(() => window.scrollBy(0, %d));\n", indent, client, a.Y)
	case "wait":
		return fmt.Sprintf("%sawait %s.pause(%d);\n", indent, client, a.Ms)
	case "screenshot":
		return fmt.Sprintf("%sawait %s.saveScreenshot('screenshot.png');\n", indent, client)
	case "network_wait":
		return fmt.Sprintf("%s// Network waits depend on your WebdriverIO service/plugin. Pattern: %s, timeout: %d\n", indent, escapeJS(a.NetworkURL), max(a.Ms, 10000))
	case "assert":
		return assertToWebdriverIOCode(a, indent, subject, client, timeout)
	default:
		return fmt.Sprintf("%s// Unknown action: %s\n", indent, a.Action)
	}
}

func assertToPlaywrightCode(a BrowserAction, indent string, timeout int) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toBeVisible({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "hidden":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toBeHidden({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "text_contains":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toContainText('%s', { timeout: %d });\n",
			indent, escapeJS(a.Selector), escapeJS(a.Text), timeout)
	case "url_contains":
		return fmt.Sprintf("%sawait expect(page).toHaveURL(/%s/);\n",
			indent, escapeRegex(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sawait expect(page).toHaveTitle(/%s/);\n",
			indent, escapeRegex(a.Text))
	case "count":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toHaveCount(%s);\n",
			indent, escapeJS(a.Selector), a.Text)
	case "attribute":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toHaveAttribute('%s', '%s');\n",
			indent, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

func assertToCypressCode(a BrowserAction, indent string) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%scy.get('%s').should('be.visible');\n", indent, escapeJS(a.Selector))
	case "hidden":
		return fmt.Sprintf("%scy.get('%s').should('not.be.visible');\n", indent, escapeJS(a.Selector))
	case "text_contains":
		return fmt.Sprintf("%scy.get('%s').should('contain.text', '%s');\n", indent, escapeJS(a.Selector), escapeJS(a.Text))
	case "url_contains":
		return fmt.Sprintf("%scy.url().should('include', '%s');\n", indent, escapeJS(a.Text))
	case "title_contains":
		return fmt.Sprintf("%scy.title().should('include', '%s');\n", indent, escapeJS(a.Text))
	case "count":
		return fmt.Sprintf("%scy.get('%s').should('have.length', %s);\n", indent, escapeJS(a.Selector), a.Text)
	case "attribute":
		return fmt.Sprintf("%scy.get('%s').should('have.attr', '%s', '%s');\n", indent, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

func assertToPuppeteerCode(a BrowserAction, indent string) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%sif (!(await page.$('%s'))) throw new Error('Expected element to be visible: %s');\n", indent, escapeJS(a.Selector), escapeJS(a.Selector))
	case "hidden":
		return fmt.Sprintf("%sif (await page.$('%s')) throw new Error('Expected element to be hidden: %s');\n", indent, escapeJS(a.Selector), escapeJS(a.Selector))
	case "text_contains":
		return fmt.Sprintf("%sif (!((await page.$eval('%s', el => el.textContent || '')).includes('%s'))) throw new Error('Expected text to contain %s');\n", indent, escapeJS(a.Selector), escapeJS(a.Text), escapeJS(a.Text))
	case "url_contains":
		return fmt.Sprintf("%sif (!page.url().includes('%s')) throw new Error('Expected URL to contain %s');\n", indent, escapeJS(a.Text), escapeJS(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sif (!((await page.title()).includes('%s'))) throw new Error('Expected title to contain %s');\n", indent, escapeJS(a.Text), escapeJS(a.Text))
	case "count":
		return fmt.Sprintf("%sif ((await page.$$('%s')).length !== %s) throw new Error('Expected count %s');\n", indent, escapeJS(a.Selector), a.Text, a.Text)
	case "attribute":
		return fmt.Sprintf("%sif ((await page.$eval('%s', el => el.getAttribute('%s'))) !== '%s') throw new Error('Expected attribute %s=%s');\n", indent, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

func assertToSeleniumCode(a BrowserAction, indent string) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%swait.until(EC.visibility_of_element_located(by_css('%s')))\n", indent, escapePy(a.Selector))
	case "hidden":
		return fmt.Sprintf("%swait.until(EC.invisibility_of_element_located(by_css('%s')))\n", indent, escapePy(a.Selector))
	case "text_contains":
		return fmt.Sprintf("%sassert '%s' in wait.until(EC.presence_of_element_located(by_css('%s'))).text\n", indent, escapePy(a.Text), escapePy(a.Selector))
	case "url_contains":
		return fmt.Sprintf("%sassert '%s' in driver.current_url\n", indent, escapePy(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sassert '%s' in driver.title\n", indent, escapePy(a.Text))
	case "count":
		return fmt.Sprintf("%sassert len(driver.find_elements(*by_css('%s'))) == %s\n", indent, escapePy(a.Selector), a.Text)
	case "attribute":
		return fmt.Sprintf("%sassert wait.until(EC.presence_of_element_located(by_css('%s'))).get_attribute('%s') == '%s'\n", indent, escapePy(a.Selector), escapePy(a.Key), escapePy(a.Text))
	default:
		return fmt.Sprintf("%s# Unknown assert: %s\n", indent, a.Assert)
	}
}

func assertToAppiumCode(a BrowserAction, indent string, timeout int) string {
	subject := "driver.$"
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%sawait (%s('%s')).waitForDisplayed({ timeout: %d });\n", indent, subject, escapeJS(a.Selector), timeout)
	case "hidden":
		return fmt.Sprintf("%sawait (%s('%s')).waitForDisplayed({ timeout: %d, reverse: true });\n", indent, subject, escapeJS(a.Selector), timeout)
	case "text_contains":
		return fmt.Sprintf("%sif (!((await (%s('%s')).getText()).includes('%s'))) throw new Error('Expected text to contain %s');\n", indent, subject, escapeJS(a.Selector), escapeJS(a.Text), escapeJS(a.Text))
	case "url_contains":
		return fmt.Sprintf("%sif (!((await driver.getUrl()).includes('%s'))) throw new Error('Expected URL to contain %s');\n", indent, escapeJS(a.Text), escapeJS(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sif (!((await driver.getTitle()).includes('%s'))) throw new Error('Expected title to contain %s');\n", indent, escapeJS(a.Text), escapeJS(a.Text))
	case "count":
		return fmt.Sprintf("%sif ((await driver.$$('%s')).length !== %s) throw new Error('Expected count %s');\n", indent, escapeJS(a.Selector), a.Text, a.Text)
	case "attribute":
		return fmt.Sprintf("%sif ((await (%s('%s')).getAttribute('%s')) !== '%s') throw new Error('Expected attribute %s=%s');\n", indent, subject, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

func assertToWebdriverIOCode(a BrowserAction, indent string, subject string, client string, timeout int) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%sawait expect(%s('%s')).toBeDisplayed({ timeout: %d });\n", indent, subject, escapeJS(a.Selector), timeout)
	case "hidden":
		return fmt.Sprintf("%sawait expect(%s('%s')).not.toBeDisplayed({ timeout: %d });\n", indent, subject, escapeJS(a.Selector), timeout)
	case "text_contains":
		return fmt.Sprintf("%sawait expect(%s('%s')).toHaveText(expect.stringContaining('%s'));\n", indent, subject, escapeJS(a.Selector), escapeJS(a.Text))
	case "url_contains":
		return fmt.Sprintf("%sawait expect(%s).toHaveUrl(expect.stringContaining('%s'));\n", indent, client, escapeJS(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sawait expect(%s).toHaveTitle(expect.stringContaining('%s'));\n", indent, client, escapeJS(a.Text))
	case "count":
		return fmt.Sprintf("%sawait expect(await $$('%s')).toBeElementsArrayOfSize(%s);\n", indent, escapeJS(a.Selector), a.Text)
	case "attribute":
		return fmt.Sprintf("%sawait expect(%s('%s')).toHaveAttribute('%s', '%s');\n", indent, subject, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

// ExportAllScripts converts all test files to a map of filename → Playwright code.
func ExportAllScripts(testFiles []TestFile, opts ExportOptions) map[string]string {
	scripts, _ := ExportAllScriptsForTarget(testFiles, "playwright", opts)
	return scripts
}

// ExportAllScriptsForTarget converts test files to the requested framework target.
func ExportAllScriptsForTarget(testFiles []TestFile, target string, opts ExportOptions) (map[string]string, ExportTarget) {
	resolved := ResolveExportTarget(target)
	result := make(map[string]string, len(testFiles))
	for _, tf := range testFiles {
		name := exportFilename(tf.Name, resolved.Ext)
		result[name] = ExportScriptForTarget(tf, resolved, opts)
	}
	return result, resolved
}

func ExportScriptForTarget(tf TestFile, target ExportTarget, opts ExportOptions) string {
	switch target.Key {
	case "cypress":
		return ExportCypressScript(tf, opts)
	case "puppeteer":
		return ExportPuppeteerScript(tf, opts)
	case "selenium":
		return ExportSeleniumScript(tf, opts)
	case "appium":
		return ExportAppiumScript(tf, opts)
	case "webdriverio":
		return ExportWebdriverIOScript(tf, opts)
	default:
		return ExportPlaywrightScript(tf, opts)
	}
}

func parseExportActions(tf TestFile) ([]BrowserAction, error) {
	var actions []BrowserAction
	if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
		return nil, err
	}
	return actions, nil
}

func exportParseError(err error) string {
	return fmt.Sprintf("// Error: could not parse actions: %v\n", err)
}

func exportFilename(name string, ext string) string {
	base := testNameFromFile(name)
	if ext == "" {
		ext = ".spec.ts"
	}
	return base + ext
}

func testNameFromFile(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.TrimSuffix(base, ".spec")
	base = strings.TrimSuffix(base, ".test")
	if base == "" || base == "." {
		return "generated_test"
	}
	return base
}

func exportIndent(opts ExportOptions) string {
	if opts.IndentWith == "" {
		return "  "
	}
	return opts.IndentWith
}

func exportTimeoutMS(opts ExportOptions) int {
	if opts.Timeout == 0 {
		return 5000
	}
	return opts.Timeout
}

func exportTimeoutSeconds(opts ExportOptions) int {
	ms := exportTimeoutMS(opts)
	seconds := ms / 1000
	if seconds <= 0 {
		seconds = 5
	}
	return seconds
}

func seleniumKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "enter", "return":
		return "ENTER"
	case "tab":
		return "TAB"
	case "escape", "esc":
		return "ESCAPE"
	case "backspace":
		return "BACKSPACE"
	case "delete":
		return "DELETE"
	case "space":
		return "SPACE"
	default:
		return "ENTER"
	}
}

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func escapePy(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func escapeRegex(s string) string {
	special := []string{"\\", ".", "*", "+", "?", "(", ")", "[", "]", "{", "}", "^", "$", "|", "/"}
	for _, c := range special {
		s = strings.ReplaceAll(s, c, "\\"+c)
	}
	return s
}
