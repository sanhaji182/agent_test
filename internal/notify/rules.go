package notify

import (
	"log/slog"

	"github.com/go-go-golems/gotest-agent/internal/drift"
)

// AlertRule defines a notification rule for test health monitoring.
type AlertRule struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`      // "drift", "failure", "flake"
	Condition map[string]string `json:"condition"` // e.g. {"severity": "high"}
	Channels  []string          `json:"channels"`  // "slack", "email", "webhook"
	Config    map[string]string `json:"config,omitempty"`
	Enabled   bool              `json:"enabled"`
}

// AlertResult captures the outcome of an alert rule evaluation.
type AlertResult struct {
	RuleID    string `json:"rule_id"`
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	Delivered bool   `json:"delivered"`
	Error     string `json:"error,omitempty"`
}

// RuleEngine evaluates alert rules against drift/failure events.
type RuleEngine struct {
	Rules []AlertRule
}

// NewRuleEngine creates a rule engine with built-in default rules.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		Rules: []AlertRule{
			{
				ID:        "drift-high-severity",
				Name:      "High-severity drift alert",
				Type:      "drift",
				Condition: map[string]string{"severity": "high"},
				Channels:  []string{"webhook"},
				Enabled:   true,
			},
			{
				ID:        "drift-missing-test",
				Name:      "Missing test drift alert",
				Type:      "drift",
				Condition: map[string]string{"type": "missing_test"},
				Channels:  []string{"webhook"},
				Enabled:   true,
			},
		},
	}
}

// EvaluateDriftAlert evaluates drift alert rules and returns results.
func (engine *RuleEngine) EvaluateDriftAlert(drifts []drift.Drift) []AlertResult {
	var results []AlertResult
	for _, rule := range engine.Rules {
		if !rule.Enabled || rule.Type != "drift" {
			continue
		}
		for _, d := range drifts {
			if !matchesCondition(d, rule.Condition) {
				continue
			}
			msg := "Drift detected: " + d.Description + " in " + d.Repository + " (" + d.FilePath + ")"
			for _, ch := range rule.Channels {
				err := NotifyDrift(ch, rule.Config, d)
				result := AlertResult{
					RuleID:    rule.ID,
					Channel:   ch,
					Message:   msg,
					Delivered: err == nil,
				}
				if err != nil {
					result.Error = err.Error()
				}
				results = append(results, result)
			}
		}
	}
	return results
}

func matchesCondition(d drift.Drift, cond map[string]string) bool {
	if sev, ok := cond["severity"]; ok && d.Severity != sev {
		return false
	}
	if typ, ok := cond["type"]; ok && d.Type != typ {
		return false
	}
	return true
}

// NotifyDrift sends a drift notification through the specified channel.
func NotifyDrift(channel string, config map[string]string, d drift.Drift) error {
	switch channel {
	case "webhook":
		url := ""
		if config != nil {
			url = config["webhook_url"]
		}
		if url == "" {
			slog.Info("drift notification skipped: no webhook URL configured")
			return nil
		}
		return DeliverWebhook(url, map[string]string{
			"type":        "drift",
			"repository":  d.Repository,
			"severity":    d.Severity,
			"file_path":   d.FilePath,
			"description": d.Description,
			"drift_type":  d.Type,
		})
	case "email":
		return SendEmail(d)
	default:
		slog.Info("drift notification: unsupported channel", "channel", channel)
		return nil
	}
}

// SendEmail is a placeholder for email notification delivery.
func SendEmail(d drift.Drift) error {
	slog.Info("drift notification: email would be sent (placefolder)",
		"repository", d.Repository,
		"file_path", d.FilePath,
		"severity", d.Severity,
	)
	return nil
}

// BatchNotifyDrift evaluates alert rules against detected drifts.
func BatchNotifyDrift(drifts []drift.Drift, engine *RuleEngine) {
	if engine == nil {
		engine = NewRuleEngine()
	}
	results := engine.EvaluateDriftAlert(drifts)
	for _, r := range results {
		if !r.Delivered {
			slog.Warn("drift alert delivery failed", "rule", r.RuleID, "channel", r.Channel, "error", r.Error)
		}
	}
}
