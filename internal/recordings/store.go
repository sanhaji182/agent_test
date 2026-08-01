// Package recordings menyediakan model dan store untuk metadata rekaman eksekusi test.
package recordings

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Recording menyimpan metadata satu rekaman (screenshot/step capture)
type Recording struct {
	ID            string    `json:"id"`
	RunID         string    `json:"run_id"`
	TestName      string    `json:"test_name"`
	StepName      string    `json:"step_name"`
	ScreenshotURL string    `json:"screenshot_url"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"` // "captured", "failed", "pending"
}

// Session menyimpan metadata satu sesi record & playback.
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ProjectPath  string                 `json:"project_path"`
	BaseURL      string                 `json:"base_url"`
	Status       string                 `json:"status"` // "recording", "paused", "completed", "aborted"
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	eventCounter int64                  `json:"-"` // internal sequence counter
}

// EventType represents the kind of recorded DOM interaction.
type EventType string

const (
	EventClick         EventType = "click"
	EventFill          EventType = "fill"
	EventNavigate      EventType = "navigate"
	EventSelect        EventType = "select"
	EventHover         EventType = "hover"
	EventPress         EventType = "press"
	EventScroll        EventType = "scroll"
	EventWait          EventType = "wait"
	EventAssertText    EventType = "assert_text"
	EventAssertVisible EventType = "assert_visible"
)

// Event menyimpan satu event yang direkam dalam sebuah sesi.
type Event struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	EventType     EventType              `json:"event_type"`
	Selector      string                 `json:"selector,omitempty"`
	Value         string                 `json:"value,omitempty"`
	URL           string                 `json:"url,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	SequenceOrder int                    `json:"sequence_order"`
	CreatedAt     time.Time              `json:"created_at"`
}

// MetadataString returns the JSON-encoded metadata, or an empty string if none.
func (e Event) MetadataString() string {
	if len(e.Metadata) == 0 {
		return ""
	}
	b, err := json.Marshal(e.Metadata)
	if err != nil {
		return ""
	}
	return string(b)
}

// EventTypes returns the distinct event types present in the slice, preserving order.
func EventTypes(events []Event) []EventType {
	seen := make(map[EventType]struct{})
	var result []EventType
	for _, e := range events {
		if _, ok := seen[e.EventType]; ok {
			continue
		}
		seen[e.EventType] = struct{}{}
		result = append(result, e.EventType)
	}
	return result
}

// Store menyimpan recordings di memori
type Store struct {
	mu         sync.RWMutex
	recordings []Recording
	sessions   map[string]*Session
	events     map[string][]Event
	counter    int64
	dbPool     *pgxpool.Pool
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
		events:   make(map[string][]Event),
	}
}

// EnableDB enables PostgreSQL persistence for recording sessions and events.
func (s *Store) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

// Add menambahkan recording baru
func (s *Store) Add(rec Recording) *Recording {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if rec.ID == "" {
		rec.ID = rec.RunID + "-rec-" + itoa(s.counter)
	}
	s.recordings = append(s.recordings, rec)
	return &rec
}

// ByRun mengembalikan semua recordings untuk sebuah run
func (s *Store) ByRun(runID string) []Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Recording
	for _, r := range s.recordings {
		if r.RunID == runID {
			result = append(result, r)
		}
	}
	return result
}

// All mengembalikan semua recordings
func (s *Store) All() []Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Recording{}, s.recordings...)
}

// CreateSession menyimpan sesi baru. Jika ID kosong, ID akan digenerate.
func (s *Store) CreateSession(sess Session) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	now := time.Now()
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	if sess.UpdatedAt.IsZero() {
		sess.UpdatedAt = now
	}
	if sess.Status == "" {
		sess.Status = "recording"
	}
	stored := sess
	s.sessions[sess.ID] = &stored
	if s.dbPool != nil {
		if err := s.persistSessionDB(&stored); err != nil {
			slog.Warn("recording session persistence failed", "session_id", stored.ID, "error", err)
		}
	}
	return &stored
}

// GetSession mengembalikan sesi berdasarkan ID.
func (s *Store) GetSession(id string) (*Session, bool) {
	if s.dbPool != nil {
		if sess, err := s.getSessionDB(id); err == nil {
			return sess, true
		} else {
			slog.Warn("recording session DB read failed, falling back to memory", "session_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	copied := *sess
	return &copied, true
}

// ListSessions mengembalikan semua sesi yang tersimpan, diurutkan dari yang terbaru.
func (s *Store) ListSessions() []Session {
	if s.dbPool != nil {
		if sessions, err := s.listSessionsDB(); err == nil {
			return sessions
		} else {
			slog.Warn("recording sessions DB list failed, falling back to memory", "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		result = append(result, *sess)
	}
	sortSessionsDesc(result)
	return result
}

// UpdateSessionStatus memperbarui status sesi dan timestamp updated_at-nya.
func (s *Store) UpdateSessionStatus(id, status string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	sess.Status = status
	sess.UpdatedAt = time.Now().Add(time.Millisecond)
	copied := *sess
	if s.dbPool != nil {
		if err := s.persistSessionDB(&copied); err != nil {
			slog.Warn("recording session status persistence failed", "session_id", id, "error", err)
		}
	}
	return &copied, true
}

// UpdateSession applies an in-place mutation to the session matching id.
func (s *Store) UpdateSession(id string, fn func(*Session)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return false
	}
	fn(sess)
	sess.UpdatedAt = time.Now()
	copied := *sess
	if s.dbPool != nil {
		if err := s.persistSessionDB(&copied); err != nil {
			slog.Warn("recording session update persistence failed", "session_id", id, "error", err)
		}
	}
	return true
}

// DeleteSession menghapus sesi dan semua event-nya.
func (s *Store) DeleteSession(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return false
	}
	delete(s.sessions, id)
	delete(s.events, id)
	if s.dbPool != nil {
		if err := s.deleteSessionDB(id); err != nil {
			slog.Warn("recording session DB delete failed", "session_id", id, "error", err)
		}
	}
	return true
}

// AddEvent menambahkan event ke dalam sesi. SequenceOrder otomatis bertambah.
func (s *Store) AddEvent(ev Event) *Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if ev.ID == "" {
		ev.ID = uuid.New().String()
	}
	now := time.Now()
	if ev.Timestamp.IsZero() {
		ev.Timestamp = now
	}
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = now
	}
	sess, ok := s.sessions[ev.SessionID]
	if ok {
		sess.eventCounter++
		ev.SequenceOrder = int(sess.eventCounter)
	}
	s.events[ev.SessionID] = append(s.events[ev.SessionID], ev)
	if s.dbPool != nil {
		if err := s.persistEventDB(&ev); err != nil {
			slog.Warn("recorded event persistence failed", "event_id", ev.ID, "session_id", ev.SessionID, "error", err)
		}
	}
	return &ev
}

// GetEventsBySession mengembalikan semua event untuk sessionID, diurutkan berdasarkan sequence_order.
func (s *Store) GetEventsBySession(sessionID string) []Event {
	if s.dbPool != nil {
		if events, err := s.getEventsBySessionDB(sessionID); err == nil {
			return events
		} else {
			slog.Warn("recorded events DB read failed, falling back to memory", "session_id", sessionID, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := s.events[sessionID]
	if events == nil {
		return nil
	}
	result := append([]Event{}, events...)
	sortEvents(result)
	return result
}

// GetSessionWithEvents mengembalikan sesi beserta semua event-nya.
func (s *Store) GetSessionWithEvents(sessionID string) (*Session, []Event, bool) {
	if s.dbPool != nil {
		if sess, err := s.getSessionDB(sessionID); err == nil {
			events, err := s.getEventsBySessionDB(sessionID)
			if err != nil {
				slog.Warn("recorded events DB read failed, falling back to memory", "session_id", sessionID, "error", err)
			} else {
				return sess, events, true
			}
		} else {
			slog.Warn("recording session DB read failed, falling back to memory", "session_id", sessionID, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, nil, false
	}
	copied := *sess
	events := append([]Event{}, s.events[sessionID]...)
	sortEvents(events)
	return &copied, events, true
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
