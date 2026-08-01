package intelligence

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestReviewGeneratedTest_GoodTest(t *testing.T) {
	code := `
import { test, expect } from '@playwright/test';

test('user can login with valid credentials', async ({ page }) => {
  await page.goto('/login');
  await page.getByTestId('email').fill('user@example.com');
  await page.getByTestId('password').fill('password123');
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
});
`
	review := ReviewGeneratedTest(code)
	if !review.Passed {
		t.Error("expected review to pass for well-structured test")
	}
	if review.Score < 80 {
		t.Errorf("expected high score, got %d", review.Score)
	}
}

func TestReviewGeneratedTest_FragileSelectors(t *testing.T) {
	code := `
test('login', async ({ page }) => {
  await page.click('.css-abc123 > div > div > div');
});
`
	review := ReviewGeneratedTest(code)
	if review.Score >= 100 {
		t.Error("expected score reduction for fragile selectors")
	}
	if len(review.Findings) == 0 {
		t.Error("expected findings for fragile selectors")
	}
}

func TestReviewGeneratedTest_MissingAwait(t *testing.T) {
	code := `
test('click button', async ({ page }) => {
  page.click('#submit');
  await expect(page.locator('.success')).toBeVisible();
});
`
	review := ReviewGeneratedTest(code)
	if review.Score >= 85 {
		t.Errorf("expected score reduction for missing await, got %d", review.Score)
	}
}

func TestReviewGeneratedTest_NoAssertions(t *testing.T) {
	code := `
test('navigate', async ({ page }) => {
  await page.goto('/home');
  await page.click('#menu');
});
`
	review := ReviewGeneratedTest(code)
	if review.Score >= 75 {
		t.Errorf("expected significant score reduction for no assertions, got %d", review.Score)
	}
	hasFinding := false
	for _, f := range review.Findings {
		if f.Category == "assertion" {
			hasFinding = true
			break
		}
	}
	if !hasFinding {
		t.Error("expected assertion finding")
	}
}

func TestReviewGeneratedTest_HardcodedCredentials(t *testing.T) {
	code := `
test('api call', async ({ page }) => {
  const token = 'sk-abc123';
  await page.goto('/api');
});
`
	review := ReviewGeneratedTest(code)
	hasFinding := false
	for _, f := range review.Findings {
		if f.Category == "structure" && f.Severity == "error" {
			hasFinding = true
			break
		}
	}
	if !hasFinding {
		t.Error("expected finding for hardcoded credentials")
	}
}

func TestReviewGeneratedTest_ShortName(t *testing.T) {
	code := `test('a', async ({ page }) => {});`
	review := ReviewGeneratedTest(code)
	hasFinding := false
	for _, f := range review.Findings {
		if f.Category == "naming" {
			hasFinding = true
			break
		}
	}
	if !hasFinding {
		t.Error("expected naming finding for short test name")
	}
}

func TestReviewTestRun(t *testing.T) {
	run := &agent.TestRun{
		ID: "r1",
		TestFiles: []agent.TestFile{
			{Name: "test1.spec.ts", Content: `test('good test', async ({ page }) => { await page.goto('/'); await expect(page).toHaveTitle('Home'); });`},
			{Name: "test2.spec.ts", Content: `test('bad', async ({ page }) => { page.click('#btn'); });`},
		},
	}

	reviews := ReviewTestRun(run)
	if len(reviews) != 2 {
		t.Fatalf("expected 2 reviews, got %d", len(reviews))
	}
	if reviews[0].TestName != "test1.spec.ts" {
		t.Errorf("expected test1.spec.ts, got %s", reviews[0].TestName)
	}
}

func TestExtractTestName(t *testing.T) {
	cases := map[string]string{
		`test('login flow', async () => {})`:   "login flow",
		`test("checkout", async () => {})`:     "checkout",
		`test("edge ' quote", async () => {})`: "edge ' quote",
		`some random code`:                     "unknown",
	}
	for code, want := range cases {
		got := extractTestName(code)
		if got != want {
			t.Errorf("extractTestName(%q) = %q, want %q", code, got, want)
		}
	}
}
