package schedule

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Frequency string

const (
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
	Cron    Frequency = "cron"
)

type Schedule struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Name          string    `json:"name"`
	ProjectPath   string    `json:"project_path"`
	Requirements  string    `json:"requirements"`
	Mode          string    `json:"mode"`
	Environment   string    `json:"environment"`    // local, staging, production
	BaseURL       string    `json:"base_url"`       // Target URL for tests
	Frequency     Frequency `json:"frequency"`
	CronExpr      string    `json:"cron_expr,omitempty"`
	Timezone      string    `json:"timezone"`
	Enabled       bool      `json:"enabled"`
	NextRunAt     time.Time `json:"next_run_at"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	LastRunID     string    `json:"last_run_id,omitempty"`
	LastRunStatus string    `json:"last_run_status,omitempty"`
	NotifyOnFail  bool      `json:"notify_on_fail"`
	WebhookURL    string    `json:"webhook_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Store struct {
	mu        sync.RWMutex
	schedules map[string]*Schedule
	order     []string
}

func NewStore() *Store {
	return &Store{schedules: make(map[string]*Schedule)}
}

func (s *Store) Create(sch *Schedule) *Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sch.ID == "" {
		sch.ID = uuid.New().String()
	}
	sch.CreatedAt = time.Now()
	sch.UpdatedAt = time.Now()
	if sch.NextRunAt.IsZero() {
		sch.NextRunAt = CalcNextRun(sch.Frequency, sch.CronExpr, time.Now())
	}
	s.schedules[sch.ID] = sch
	s.order = append([]string{sch.ID}, s.order...)
	return sch
}

func (s *Store) Get(id string) (*Schedule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sch, ok := s.schedules[id]
	return sch, ok
}

func (s *Store) List() []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Schedule
	for _, id := range s.order {
		if sch, ok := s.schedules[id]; ok {
			result = append(result, sch)
		}
	}
	return result
}

func (s *Store) Update(id string, fn func(*Schedule)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sch, ok := s.schedules[id]
	if !ok {
		return false
	}
	fn(sch)
	sch.UpdatedAt = time.Now()
	return true
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return false
	}
	delete(s.schedules, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// GetDue returns schedules that are enabled and past their next_run_at
func (s *Store) GetDue(now time.Time) []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var due []*Schedule
	for _, sch := range s.schedules {
		if sch.Enabled && !sch.NextRunAt.After(now) {
			due = append(due, sch)
		}
	}
	return due
}

// CalcNextRun calculates the next run time based on frequency
func CalcNextRun(freq Frequency, cronExpr string, from time.Time) time.Time {
	switch freq {
	case Daily:
		return from.Add(24 * time.Hour)
	case Weekly:
		return from.Add(7 * 24 * time.Hour)
	case Monthly:
		return from.AddDate(0, 1, 0)
	default:
		return from.Add(24 * time.Hour)
	}
}
