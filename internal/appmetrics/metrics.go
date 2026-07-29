// Package appmetrics provides lightweight application metrics in Prometheus
// exposition format, using only the standard library (no external deps).
// Counters use sync/atomic for lock-free increments; gauges are read on scrape.
package appmetrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds all application counters and gauges.
type Metrics struct {
	// Counters (monotonic, atomic)
	RunsCreated    atomic.Int64
	RunsCompleted  atomic.Int64
	RunsFailed     atomic.Int64
	RunsCancelled  atomic.Int64
	ActionsExecuted atomic.Int64
	ActionsHealed  atomic.Int64
	ActionsRetried atomic.Int64
	APIRequests    atomic.Int64
	APIErrors      atomic.Int64

	// Gauges (point-in-time, guarded by mutex)
	mu          sync.RWMutex
	activeRuns  int64
	lastRunUnix int64

	// Histograms: action durations in ms (bucketed)
	histMu        sync.Mutex
	durationBuckets map[string]int64 // bucket label -> count
}

// Default bucket boundaries for action duration (ms).
var durationBounds = []int64{100, 500, 1000, 5000, 10000, 30000}

// New creates a Metrics instance with initialized histogram buckets.
func New() *Metrics {
	m := &Metrics{durationBuckets: make(map[string]int64)}
	for _, b := range durationBounds {
		m.durationBuckets[fmt.Sprintf("le_%d", b)] = 0
	}
	m.durationBuckets["le_Inf"] = 0
	return m
}

// SetActiveRuns updates the active-runs gauge.
func (m *Metrics) SetActiveRuns(n int64) {
	m.mu.Lock()
	m.activeRuns = n
	m.mu.Unlock()
}

// RecordRunFinish records a completed run's timestamp.
func (m *Metrics) RecordRunFinish() {
	m.mu.Lock()
	m.lastRunUnix = time.Now().Unix()
	m.mu.Unlock()
}

// ObserveActionDuration records an action's execution time into histogram buckets.
func (m *Metrics) ObserveActionDuration(ms int64) {
	m.histMu.Lock()
	defer m.histMu.Unlock()
	for _, b := range durationBounds {
		if ms <= b {
			m.durationBuckets[fmt.Sprintf("le_%d", b)]++
		}
	}
	m.durationBuckets["le_Inf"]++
}

// Render produces the Prometheus text exposition format.
func (m *Metrics) Render() string {
	var sb strings.Builder

	writeCounter(&sb, "gotest_runs_created_total", "Total test runs created", m.RunsCreated.Load())
	writeCounter(&sb, "gotest_runs_completed_total", "Total test runs completed", m.RunsCompleted.Load())
	writeCounter(&sb, "gotest_runs_failed_total", "Total test runs failed", m.RunsFailed.Load())
	writeCounter(&sb, "gotest_runs_cancelled_total", "Total test runs cancelled", m.RunsCancelled.Load())
	writeCounter(&sb, "gotest_actions_executed_total", "Total browser actions executed", m.ActionsExecuted.Load())
	writeCounter(&sb, "gotest_actions_healed_total", "Total actions recovered by self-healing", m.ActionsHealed.Load())
	writeCounter(&sb, "gotest_actions_retried_total", "Total actions recovered by auto-retry", m.ActionsRetried.Load())
	writeCounter(&sb, "gotest_api_requests_total", "Total API requests", m.APIRequests.Load())
	writeCounter(&sb, "gotest_api_errors_total", "Total API error responses", m.APIErrors.Load())

	m.mu.RLock()
	writeGauge(&sb, "gotest_active_runs", "Currently executing test runs", m.activeRuns)
	writeGauge(&sb, "gotest_last_run_timestamp_seconds", "Unix timestamp of last finished run", m.lastRunUnix)
	m.mu.RUnlock()

	// Histogram: action durations
	sb.WriteString("# HELP gotest_action_duration_ms Action execution duration in milliseconds\n")
	sb.WriteString("# TYPE gotest_action_duration_ms histogram\n")
	m.histMu.Lock()
	keys := make([]string, 0, len(m.durationBuckets))
	for k := range m.durationBuckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		le := strings.TrimPrefix(k, "le_")
		fmt.Fprintf(&sb, "gotest_action_duration_ms_bucket{le=\"%s\"} %d\n", le, m.durationBuckets[k])
	}
	m.histMu.Unlock()

	return sb.String()
}

func writeCounter(sb *strings.Builder, name, help string, val int64) {
	fmt.Fprintf(sb, "# HELP %s %s\n# TYPE %s counter\n%s %d\n", name, help, name, name, val)
}

func writeGauge(sb *strings.Builder, name, help string, val int64) {
	fmt.Fprintf(sb, "# HELP %s %s\n# TYPE %s gauge\n%s %d\n", name, help, name, name, val)
}
