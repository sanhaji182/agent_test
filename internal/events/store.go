// Package events menyediakan model dan store untuk step-level execution events.
// Setiap run menghasilkan event granular yang bisa di-stream via SSE.
package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventType adalah tipe event yang dihasilkan selama eksekusi run
type EventType string

const (
	RunStarted          EventType = "run_started"
	AnalysisStarted     EventType = "analysis_started"
	AnalysisCompleted   EventType = "analysis_completed"
	PlanGenerated       EventType = "plan_generated"
	ScriptGenerated     EventType = "script_generated"
	TestStarted         EventType = "test_started"
	StepStarted         EventType = "step_started"
	StepCompleted       EventType = "step_completed"
	ScreenshotCaptured  EventType = "screenshot_captured"
	AssertionPassed     EventType = "assertion_passed"
	AssertionFailed     EventType = "assertion_failed"
	FixAttemptStarted   EventType = "fix_attempt_started"
	FixAttemptCompleted EventType = "fix_attempt_completed"
	RunCompleted        EventType = "run_completed"
	RunFailed           EventType = "run_failed"

	// MaxEventsPerRun is the maximum number of in-memory events retained per
	// run to bound memory usage (AUDIT S-02). At ~16 events/second over 10
	// minutes this accommodates realistic runs. Events beyond the cap are
	// dropped from in-memory storage (on a first-in-first-out basis) but
	// are still delivered to subscribers and persisted to DB.
	MaxEventsPerRun = 10_000
)

// Event adalah satu event eksekusi dalam sebuah run
type Event struct {
	ID        string            `json:"id"`
	RunID     string            `json:"run_id"`
	Type      EventType         `json:"type"`
	Phase     string            `json:"phase"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Store menyimpan events per run di memori dan mendukung subscribe untuk streaming
type Store struct {
	mu          sync.RWMutex
	events      map[string][]Event      // runID → events
	subscribers map[string][]chan Event // runID → subscriber channels
	globalSubs  []chan Event            // subscribers to ALL events
	counter     int64

	dbPool *pgxpool.Pool // optional PostgreSQL pool for event persistence (ADR-003)
}

// EnableDB enables PostgreSQL persistence. Events will be written to the
// run_events table in addition to in-memory delivery. The pool must be
// non-nil; call this before any Emit calls.
func (s *Store) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

// NewStore membuat event store baru
func NewStore() *Store {
	return &Store{
		events:      make(map[string][]Event),
		subscribers: make(map[string][]chan Event),
	}
}

// Emit menambahkan event baru dan mengirimnya ke semua subscriber.
// If PostgreSQL persistence is enabled (EnableDB), the event is also written
// to the run_events table asynchronously.
func (s *Store) Emit(runID string, eventType EventType, phase, message string, metadata map[string]string) *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	evt := Event{
		ID:        runID + "-" + itoa(s.counter),
		RunID:     runID,
		Type:      eventType,
		Phase:     phase,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	s.events[runID] = append(s.events[runID], evt)

	// Cap in-memory events per run to prevent unbounded growth (AUDIT S-02).
	// Events beyond the cap are pruned from the head of the slice (FIFO).
	// Subscribers still receive all events; DB persistence is unaffected.
	if len(s.events[runID]) > MaxEventsPerRun {
		overflow := len(s.events[runID]) - MaxEventsPerRun
		s.events[runID] = s.events[runID][overflow:]
	}

	// Kirim ke semua subscriber aktif (per-run)
	for _, ch := range s.subscribers[runID] {
		select {
		case ch <- evt:
		default:
		}
	}

	// Kirim ke global subscribers (control room)
	for _, ch := range s.globalSubs {
		select {
		case ch <- evt:
		default:
		}
	}

	// Persist to PostgreSQL if enabled (ADR-003 Phase 1)
	if s.dbPool != nil {
		go s.persistToDB(runID, eventType, phase, message, metadata)
	}

	return &evt
}

func (s *Store) persistToDB(runID string, eventType EventType, phase, message string, metadata map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	metaJSON := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaJSON = b
		}
	}
	_, err := s.dbPool.Exec(ctx,
		`INSERT INTO run_events (run_id, event_type, phase, message, metadata)
		 VALUES ($1, $2, $3, $4, $5)`,
		runID, string(eventType), phase, message, metaJSON,
	)
	if err != nil {
		slog.Error("event persistence failed", "error", err, "run_id", runID, "event_type", string(eventType))
	}
}

// GetDBEvents returns historical events from the PostgreSQL run_events table.
// For active runs, use GetEvents (in-memory hot path).
func (s *Store) GetDBEvents(ctx context.Context, runID string) ([]Event, error) {
	if s.dbPool == nil {
		return nil, nil
	}
	rows, err := s.dbPool.Query(ctx,
		`SELECT id, run_id, event_type, phase, message, metadata, created_at
		 FROM run_events WHERE run_id = $1 ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var e Event
		var metaBytes []byte
		if err := rows.Scan(&e.ID, &e.RunID, &e.Type, &e.Phase, &e.Message, &metaBytes, &e.Timestamp); err != nil {
			return events, err
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &e.Metadata)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return events, err
	}
	return events, nil
}

// GetEvents mengembalikan semua events untuk sebuah run
func (s *Store) GetEvents(runID string) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events[runID]
}

// Subscribe membuat channel untuk menerima events baru secara realtime
func (s *Store) Subscribe(runID string) (chan Event, func()) {
	ch := make(chan Event, 64)
	s.mu.Lock()
	s.subscribers[runID] = append(s.subscribers[runID], ch)
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		subs := s.subscribers[runID]
		for i, sub := range subs {
			if sub == ch {
				s.subscribers[runID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsubscribe
}

// SubscribeAll membuat channel untuk menerima SEMUA events dari semua run (untuk control room)
func (s *Store) SubscribeAll() (chan Event, func()) {
	ch := make(chan Event, 128)
	s.mu.Lock()
	s.globalSubs = append(s.globalSubs, ch)
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range s.globalSubs {
			if sub == ch {
				s.globalSubs = append(s.globalSubs[:i], s.globalSubs[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsubscribe
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return string(buf[i+1:])
}
