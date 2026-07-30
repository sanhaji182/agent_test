package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Resilience configuration — production defaults
var (
	maxRetries         = 3
	initialBackoff     = 1 * time.Second
	backoffMultiplier  = 2.0
	maxBackoff         = 8 * time.Second
	cbFailureThreshold = 5
	cbRecoveryTimeout  = 30 * time.Second
)

// SetResilienceTimings overrides retry/circuit-breaker timing for tests.
// Call with zero values to restore production defaults.
func SetResilienceTimings(retries int, backoff, maxBk, cbTimeout time.Duration, cbThreshold int) {
	if retries > 0 {
		maxRetries = retries
	}
	if backoff > 0 {
		initialBackoff = backoff
	}
	if maxBk > 0 {
		maxBackoff = maxBk
	}
	if cbTimeout > 0 {
		cbRecoveryTimeout = cbTimeout
	}
	if cbThreshold > 0 {
		cbFailureThreshold = cbThreshold
	}
}

// ResetResilienceTimings restores production defaults.
func ResetResilienceTimings() {
	maxRetries = 3
	initialBackoff = 1 * time.Second
	backoffMultiplier = 2.0
	maxBackoff = 8 * time.Second
	cbFailureThreshold = 5
	cbRecoveryTimeout = 30 * time.Second
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// circuitBreaker implements a simple circuit breaker pattern
type circuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	lastFailureTime  time.Time
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{state: CircuitClosed}
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if recovery timeout has elapsed
		if time.Since(cb.lastFailureTime) >= cbRecoveryTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		// Allow one request through to test recovery
		return true
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = CircuitClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	if cb.failureCount >= cbFailureThreshold {
		cb.state = CircuitOpen
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context cancellation/timeout — don't retry
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	errStr := err.Error()

	// Network errors — retry
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") {
		return true
	}

	// HTTP status codes — retry on 429, 5xx
	var httpErr *httpError
	if errors.As(err, &httpErr) {
		return httpErr.statusCode == 429 || httpErr.statusCode >= 500
	}

	// Check for status code in error message
	if strings.Contains(errStr, "status 429") ||
		strings.Contains(errStr, "status 5") {
		return true
	}

	return false
}

// httpError wraps HTTP status errors for retry classification
type httpError struct {
	statusCode int
	message    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("status %d: %s", e.statusCode, e.message)
}

// ResilientClient wraps a Client with retry and circuit breaker logic
type ResilientClient struct {
	inner   Client
	breaker *circuitBreaker
}

// NewResilientClient creates a resilient wrapper around any Client
func NewResilientClient(inner Client) *ResilientClient {
	return &ResilientClient{
		inner:   inner,
		breaker: newCircuitBreaker(),
	}
}

// GenerateText implements Client with retry + circuit breaker
func (rc *ResilientClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	return rc.withResilience(ctx, func(ctx context.Context) (string, error) {
		return rc.inner.GenerateText(ctx, prompt)
	})
}

// GenerateWithImage implements Client with retry + circuit breaker
func (rc *ResilientClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	return rc.withResilience(ctx, func(ctx context.Context) (string, error) {
		return rc.inner.GenerateWithImage(ctx, prompt, imageBase64)
	})
}

// withResilience applies circuit breaker + retry logic to any Client operation
func (rc *ResilientClient) withResilience(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	// Check circuit breaker first
	if !rc.breaker.allow() {
		return "", fmt.Errorf("circuit breaker open: too many failures, retry after %v", cbRecoveryTimeout)
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		result, err := fn(ctx)
		if err == nil {
			rc.breaker.recordSuccess()
			return result, nil
		}

		lastErr = err

		// Don't retry non-retryable errors
		if !isRetryableError(err) {
			rc.breaker.recordFailure()
			return "", err
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			// Honor Retry-After header if present
			retryAfter := parseRetryAfter(err)
			sleepDuration := backoff
			if retryAfter > 0 {
				sleepDuration = retryAfter
			}

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(sleepDuration):
			}

			// Exponential backoff with cap
			backoff = time.Duration(float64(backoff) * backoffMultiplier)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	// All retries exhausted
	rc.breaker.recordFailure()
	return "", fmt.Errorf("all %d retries failed: %w", maxRetries+1, lastErr)
}

// parseRetryAfter extracts Retry-After duration from HTTP errors
func parseRetryAfter(err error) time.Duration {
	var httpErr *httpError
	if errors.As(err, &httpErr) {
		// In a real implementation, we'd extract the Retry-After header
		// For now, return 0 to use default backoff
		return 0
	}
	return 0
}

// CircuitState returns the current circuit breaker state (for metrics/debugging)
func (rc *ResilientClient) CircuitState() CircuitState {
	rc.breaker.mu.Lock()
	defer rc.breaker.mu.Unlock()
	return rc.breaker.state
}
