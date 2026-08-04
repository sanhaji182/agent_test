package ai

import (
	"context"
	"fmt"
	"log/slog"
)

// FallbackClient chains multiple Client implementations and fails over from one
// to the next on error. This gives the execution layer provider-level redundancy:
// if the primary provider is unavailable (e.g. out of credits, rate-limited, or
// returning malformed responses), requests automatically continue against the
// next configured provider instead of failing the whole run.
//
// Failover stops early once the context is cancelled or its deadline is exceeded,
// so a run-level timeout is respected across all providers rather than retrying
// providers that can no longer do useful work.
type FallbackClient struct {
	clients []Client
}

// NewFallbackClient returns a Client that tries the given clients in order.
// Nil clients are ignored. If exactly one usable client remains it is returned
// directly (no wrapping overhead). Returns nil when no usable client is given.
func NewFallbackClient(clients ...Client) Client {
	usable := make([]Client, 0, len(clients))
	for _, c := range clients {
		if c != nil {
			usable = append(usable, c)
		}
	}
	switch len(usable) {
	case 0:
		return nil
	case 1:
		return usable[0]
	default:
		return &FallbackClient{clients: usable}
	}
}

// GenerateText tries each provider in order, returning the first successful
// response. If all providers fail, the last error is returned wrapped.
func (f *FallbackClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	for i, c := range f.clients {
		out, err := c.GenerateText(ctx, prompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		slog.Warn("llm provider failed, failing over to next", "provider_index", i, "error", err)
	}
	return "", fmt.Errorf("all %d LLM providers failed: %w", len(f.clients), lastErr)
}

// GenerateWithImage mirrors GenerateText for vision-capable requests.
func (f *FallbackClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	var lastErr error
	for i, c := range f.clients {
		out, err := c.GenerateWithImage(ctx, prompt, imageBase64)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		slog.Warn("llm provider failed (vision), failing over to next", "provider_index", i, "error", err)
	}
	return "", fmt.Errorf("all %d LLM providers failed: %w", len(f.clients), lastErr)
}
