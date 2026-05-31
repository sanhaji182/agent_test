package gitdiff_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/gitdiff"
)

func TestMapToTests_MatchByFilename(t *testing.T) {
	changed := []string{"src/auth/login.ts", "src/checkout/cart.go"}
	allTests := []string{"login flow", "checkout process", "about page", "signup"}

	matched := gitdiff.MapToTests(changed, allTests)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matched tests, got %d: %v", len(matched), matched)
	}
}

func TestMapToTests_MatchByDirectory(t *testing.T) {
	changed := []string{"controllers/auth/handler.go"}
	allTests := []string{"auth login", "auth signup", "dashboard"}

	matched := gitdiff.MapToTests(changed, allTests)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matched (auth*), got %d: %v", len(matched), matched)
	}
}

func TestMapToTests_NoMatch_FallbackAll(t *testing.T) {
	changed := []string{"config/database.yml"}
	allTests := []string{"login", "checkout"}

	matched := gitdiff.MapToTests(changed, allTests)
	if len(matched) != 2 {
		t.Fatalf("expected fallback to all tests, got %d", len(matched))
	}
}

func TestMapToTests_Empty(t *testing.T) {
	matched := gitdiff.MapToTests(nil, []string{"a", "b"})
	if len(matched) != 2 {
		t.Fatalf("expected all tests on empty diff, got %d", len(matched))
	}
}

func TestMapToTests_StripsSuffixes(t *testing.T) {
	changed := []string{"src/login_controller.go", "src/checkout_test.ts"}
	allTests := []string{"login test", "checkout flow", "other"}

	matched := gitdiff.MapToTests(changed, allTests)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matched after suffix strip, got %d: %v", len(matched), matched)
	}
}
