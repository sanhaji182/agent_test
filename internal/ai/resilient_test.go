package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestMain configures fast resilience timings so the test suite runs in
// milliseconds instead of the production seconds/minutes.
func TestMain(m *testing.M) {
	SetResilienceTimings(
		3,                // retries
		1*time.Millisecond,  // initial backoff
		8*time.Millisecond,  // max backoff
		50*time.Millisecond, // circuit breaker recovery timeout
		5,                // circuit breaker failure threshold
	)
	code := m.Run()
	ResetResilienceTimings()
	os.Exit(code)
}

// mockClient implements Client for testing resilience behavior
type mockClient struct {
	callCount int
	failUntil int // fail until this many calls have been made
	err       error
}

func (m *mockClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	m.callCount++
	if m.callCount <= m.failUntil {
		return "", m.err
	}
	return "success", nil
}

func (m *mockClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	m.callCount++
	if m.callCount <= m.failUntil {
		return "", m.err
	}
	return "image-success", nil
}

func TestResilientClient_SuccessNoRetry(t *testing.T) {
	inner := &mockClient{failUntil: 0}
	rc := NewResilientClient(inner)

	result, err := rc.GenerateText(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %q", result)
	}
	if inner.callCount != 1 {
		t.Fatalf("expected 1 call, got %d", inner.callCount)
	}
}

func TestResilientClient_RetryOnTransientError(t *testing.T) {
	// Fail twice with a retryable error, then succeed
	inner := &mockClient{
		failUntil: 2,
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	result, err := rc.GenerateText(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %q", result)
	}
	if inner.callCount != 3 {
		t.Fatalf("expected 3 calls (2 failures + 1 success), got %d", inner.callCount)
	}
}

func TestResilientClient_NoRetryOnNonRetryableError(t *testing.T) {
	inner := &mockClient{
		failUntil: 10, // always fail
		err:       errors.New("invalid API key"),
	}
	rc := NewResilientClient(inner)

	_, err := rc.GenerateText(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for non-retryable failure")
	}
	// Should only be called once — no retry for non-retryable errors
	if inner.callCount != 1 {
		t.Fatalf("expected 1 call (no retry), got %d", inner.callCount)
	}
}

func TestResilientClient_ExhaustsRetries(t *testing.T) {
	inner := &mockClient{
		failUntil: 100, // always fail
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	_, err := rc.GenerateText(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// maxRetries=3, so 4 total attempts (initial + 3 retries)
	if inner.callCount != maxRetries+1 {
		t.Fatalf("expected %d calls, got %d", maxRetries+1, inner.callCount)
	}
}

func TestResilientClient_CircuitBreakerOpens(t *testing.T) {
	inner := &mockClient{
		failUntil: 100,
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	// Trigger enough failures to open the circuit breaker
	// Each call exhausts retries (4 attempts) and records 1 failure
	for i := 0; i < cbFailureThreshold; i++ {
		rc.GenerateText(context.Background(), "test")
	}

	// Circuit should now be open
	if rc.CircuitState() != CircuitOpen {
		t.Fatalf("expected circuit open, got %v", rc.CircuitState())
	}

	// Next call should fail immediately with circuit breaker error
	inner.callCount = 0
	_, err := rc.GenerateText(context.Background(), "test")
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}
	if inner.callCount != 0 {
		t.Fatalf("expected 0 inner calls when circuit open, got %d", inner.callCount)
	}
}

func TestResilientClient_ContextCancellation(t *testing.T) {
	inner := &mockClient{
		failUntil: 100,
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := rc.GenerateText(ctx, "test")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestResilientClient_GenerateWithImage(t *testing.T) {
	inner := &mockClient{failUntil: 1, err: errors.New("timeout")}
	rc := NewResilientClient(inner)

	result, err := rc.GenerateWithImage(context.Background(), "test", "base64data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "image-success" {
		t.Fatalf("expected 'image-success', got %q", result)
	}
}

func TestIsRetryableError(t *testing.T) {
	cases := []struct {
		err       error
		retryable bool
	}{
		{nil, false},
		{context.Canceled, false},
		{context.DeadlineExceeded, false},
		{errors.New("connection refused"), true},
		{errors.New("connection reset by peer"), true},
		{errors.New("timeout waiting for response"), true},
		{errors.New("temporary failure in name resolution"), true},
		{errors.New("status 429: rate limited"), true},
		{errors.New("status 500: internal server error"), true},
		{errors.New("status 503: service unavailable"), true},
		{errors.New("invalid API key"), false},
		{errors.New("model not found"), false},
		{&httpError{statusCode: 429, message: "rate limited"}, true},
		{&httpError{statusCode: 500, message: "server error"}, true},
		{&httpError{statusCode: 400, message: "bad request"}, false},
		{&httpError{statusCode: 401, message: "unauthorized"}, false},
	}
	for _, tc := range cases {
		got := isRetryableError(tc.err)
		if got != tc.retryable {
			t.Errorf("isRetryableError(%v) = %v, want %v", tc.err, got, tc.retryable)
		}
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := newCircuitBreaker()

	// Starts closed
	if cb.state != CircuitClosed {
		t.Fatalf("expected closed, got %v", cb.state)
	}

	// Record failures up to threshold
	for i := 0; i < cbFailureThreshold-1; i++ {
		cb.recordFailure()
	}
	if cb.state != CircuitClosed {
		t.Fatalf("expected closed before threshold, got %v", cb.state)
	}

	// One more failure opens the circuit
	cb.recordFailure()
	if cb.state != CircuitOpen {
		t.Fatalf("expected open after threshold, got %v", cb.state)
	}

	// Should not allow requests when open
	if cb.allow() {
		t.Fatal("expected allow() = false when open")
	}

	// Simulate recovery timeout elapsed
	cb.mu.Lock()
	cb.lastFailureTime = time.Now().Add(-cbRecoveryTimeout - time.Second)
	cb.mu.Unlock()

	// Should transition to half-open and allow one request
	if !cb.allow() {
		t.Fatal("expected allow() = true after recovery timeout")
	}
	if cb.state != CircuitHalfOpen {
		t.Fatalf("expected half-open, got %v", cb.state)
	}

	// Success closes the circuit
	cb.recordSuccess()
	if cb.state != CircuitClosed {
		t.Fatalf("expected closed after success, got %v", cb.state)
	}
}

func TestCircuitState_String(t *testing.T) {
	cases := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("CircuitState(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := &httpError{statusCode: 429, message: "rate limited"}
	want := "status 429: rate limited"
	if got := err.Error(); got != want {
		t.Errorf("httpError.Error() = %q, want %q", got, want)
	}
}

func TestNewResilientClient_WrapsClient(t *testing.T) {
	inner := &mockClient{}
	rc := NewResilientClient(inner)
	if rc.inner != inner {
		t.Fatal("inner client not set correctly")
	}
	if rc.breaker == nil {
		t.Fatal("circuit breaker not initialized")
	}
	if rc.CircuitState() != CircuitClosed {
		t.Fatalf("expected initial state closed, got %v", rc.CircuitState())
	}
}

func TestResilientClient_BackoffTiming(t *testing.T) {
	// Verify exponential backoff increases delay between retries
	inner := &mockClient{
		failUntil: 2,
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	start := time.Now()
	_, err := rc.GenerateText(context.Background(), "test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With 2 failures: first backoff ~1s, second ~2s = ~3s total
	// Allow some tolerance for timing
	if elapsed < 2*time.Second {
		t.Fatalf("expected at least 2s of backoff, got %v", elapsed)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("backoff took too long: %v", elapsed)
	}
}

func TestResilientClient_MaxBackoffCap(t *testing.T) {
	// Verify backoff is capped at maxBackoff
	backoff := initialBackoff
	for i := 0; i < 10; i++ {
		backoff = time.Duration(float64(backoff) * backoffMultiplier)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	if backoff != maxBackoff {
		t.Fatalf("expected backoff capped at %v, got %v", maxBackoff, backoff)
	}
}

func TestResilientClient_ConcurrentAccess(t *testing.T) {
	inner := &mockClient{failUntil: 0}
	rc := NewResilientClient(inner)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := rc.GenerateText(context.Background(), "concurrent")
			done <- err == nil
		}()
	}

	for i := 0; i < 10; i++ {
		if !<-done {
			t.Fatal("concurrent request failed")
		}
	}
}

func TestResilientClient_AllRetriesFailedErrorMessage(t *testing.T) {
	inner := &mockClient{
		failUntil: 100,
		err:       errors.New("connection refused"),
	}
	rc := NewResilientClient(inner)

	_, err := rc.GenerateText(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}

	// Error should mention retry count
	expected := fmt.Sprintf("all %d retries failed", maxRetries+1)
	if !containsSubstring(err.Error(), expected) {
		t.Fatalf("error %q should contain %q", err.Error(), expected)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
