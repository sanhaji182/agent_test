// Package drift detects drift between source code and its tests (Phase 3).
package drift

import (
	"fmt"
	"sync"
	"time"
)

// Drift types
const (
	TypeMissingTest  = "missing_test"
	TypeOutdatedTest = "outdated_test"
	TypeRemovedTest  = "removed_test"
)

// Severities
const (
	SeverityHigh   = "high"
	SeverityMedium = "medium"
	SeverityLow    = "low"
)

// Statuses
const (
	StatusPending = "pending"
	StatusFixed   = "fixed"
	StatusIgnored = "ignored"
)

// Drift represents a detected mismatch between code and tests.
type Drift struct {
	ID          string    `json:"id"`
	Repository  string    `json:"repository"`
	Type        string    `json:"type"`
	FilePath    string    `json:"file_path"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store is an in-memory drift store.
type Store struct {
	mu      sync.RWMutex
	items   []Drift
	counter int64
}

func NewStore() *Store { return &Store{} }

func (s *Store) Add(d Drift) *Drift {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	d.ID = fmt.Sprintf("drift-%d", s.counter)
	if d.Status == "" {
		d.Status = StatusPending
	}
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now
	s.items = append(s.items, d)
	return &d
}

// List returns drifts, optionally filtered by repository, type, and status
// (empty string matches all).
func (s *Store) List(repository, driftType, status string) []Drift {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Drift{}
	for _, d := range s.items {
		if repository != "" && d.Repository != repository {
			continue
		}
		if driftType != "" && d.Type != driftType {
			continue
		}
		if status != "" && d.Status != status {
			continue
		}
		result = append(result, d)
	}
	return result
}

// HasPending reports whether a pending drift already exists for the same
// repository, type, and file path (used to dedup repeated pushes).
func (s *Store) HasPending(repository, driftType, filePath string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.items {
		if s.items[i].Status == StatusPending &&
			s.items[i].Repository == repository &&
			s.items[i].Type == driftType &&
			s.items[i].FilePath == filePath {
			return true
		}
	}
	return false
}

// UpdateStatus transitions a drift to pending, fixed, or ignored.
func (s *Store) UpdateStatus(id, status string) (*Drift, error) {
	switch status {
	case StatusPending, StatusFixed, StatusIgnored:
	default:
		return nil, fmt.Errorf("invalid status: %q", status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Status = status
			s.items[i].UpdatedAt = time.Now()
			d := s.items[i]
			return &d, nil
		}
	}
	return nil, fmt.Errorf("drift not found: %s", id)
}
