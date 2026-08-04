package ai

import (
	"context"
	"errors"
	"testing"
)

// stubClient is a minimal ai.Client for fallback tests.
type stubClient struct {
	out   string
	err   error
	calls int
}

func (s *stubClient) GenerateText(context.Context, string) (string, error) {
	s.calls++
	return s.out, s.err
}

func (s *stubClient) GenerateWithImage(context.Context, string, string) (string, error) {
	s.calls++
	return s.out, s.err
}

func TestFallbackClient_FailsOverOnError(t *testing.T) {
	primary := &stubClient{err: errors.New("status 402: insufficient balance")}
	fallback := &stubClient{out: "hello"}

	c := NewFallbackClient(primary, fallback)
	out, err := c.GenerateText(context.Background(), "hi")
	if err != nil {
		t.Fatalf("expected success via fallback, got %v", err)
	}
	if out != "hello" {
		t.Fatalf("expected fallback output, got %q", out)
	}
	if primary.calls != 1 || fallback.calls != 1 {
		t.Fatalf("expected each provider tried once, got primary=%d fallback=%d", primary.calls, fallback.calls)
	}
}

func TestFallbackClient_PrimarySucceedsNoFailover(t *testing.T) {
	primary := &stubClient{out: "primary"}
	fallback := &stubClient{out: "fallback"}

	c := NewFallbackClient(primary, fallback)
	out, err := c.GenerateText(context.Background(), "hi")
	if err != nil || out != "primary" {
		t.Fatalf("expected primary output, got %q err=%v", out, err)
	}
	if fallback.calls != 0 {
		t.Fatalf("fallback must not be called when primary succeeds, got %d", fallback.calls)
	}
}

func TestFallbackClient_AllFail(t *testing.T) {
	c := NewFallbackClient(&stubClient{err: errors.New("a")}, &stubClient{err: errors.New("b")})
	if _, err := c.GenerateText(context.Background(), "hi"); err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestFallbackClient_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fallback := &stubClient{out: "x"}

	c := NewFallbackClient(&stubClient{err: errors.New("a")}, fallback)
	if _, err := c.GenerateText(ctx, "hi"); err == nil {
		t.Fatal("expected error on cancelled context")
	}
	if fallback.calls != 0 {
		t.Fatalf("fallback must not be tried after context cancel, got %d calls", fallback.calls)
	}
}

func TestFallbackClient_SinglePassthrough(t *testing.T) {
	only := &stubClient{out: "solo"}
	if c := NewFallbackClient(only); c != Client(only) {
		t.Fatal("a single usable client should be returned directly, unwrapped")
	}
}

func TestFallbackClient_NilFiltered(t *testing.T) {
	if c := NewFallbackClient(nil, nil); c != nil {
		t.Fatal("all-nil input should return nil")
	}
	only := &stubClient{out: "x"}
	if c := NewFallbackClient(nil, only, nil); c != Client(only) {
		t.Fatal("nils should be filtered, leaving the single usable client")
	}
}

func TestFallbackClient_VisionFailsOver(t *testing.T) {
	primary := &stubClient{err: errors.New("boom")}
	fallback := &stubClient{out: "img-ok"}

	c := NewFallbackClient(primary, fallback)
	out, err := c.GenerateWithImage(context.Background(), "describe", "base64")
	if err != nil || out != "img-ok" {
		t.Fatalf("expected vision fallback output, got %q err=%v", out, err)
	}
}
