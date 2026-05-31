// Package events menyediakan model dan store untuk step-level execution events.
// Setiap run menghasilkan event granular yang bisa di-stream via SSE.
package events

import (
	"sync"
	"time"
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
	mu             sync.RWMutex
	events         map[string][]Event          // runID → events
	subscribers    map[string][]chan Event      // runID → subscriber channels
	globalSubs     []chan Event                 // subscribers to ALL events
	counter        int64
}

// NewStore membuat event store baru
func NewStore() *Store {
	return &Store{
		events:      make(map[string][]Event),
		subscribers: make(map[string][]chan Event),
	}
}

// Emit menambahkan event baru dan mengirimnya ke semua subscriber
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

	return &evt
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
