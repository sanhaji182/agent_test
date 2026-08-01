// Package drift detects drift between source code and its tests (Phase 3).
package drift

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	dbPool  *pgxpool.Pool
}

func NewStore() *Store { return &Store{} }

// EnableDB enables PostgreSQL persistence for drifts.
func (s *Store) EnableDB(pool *pgxpool.Pool) { s.dbPool = pool }

func (s *Store) Add(d Drift) *Drift {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	if d.Status == "" {
		d.Status = StatusPending
	}
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now
	s.items = append(s.items, d)
	if s.dbPool != nil {
		if err := s.persistDriftDB(&d); err != nil {
			slog.Warn("drift persistence failed", "drift_id", d.ID, "error", err)
		}
	}
	return &d
}

// List returns drifts, optionally filtered by repository, type, and status
// (empty string matches all).
func (s *Store) List(repository, driftType, status string) []Drift {
	if s.dbPool != nil {
		if list, err := s.listDriftsDB(repository, driftType, status); err == nil {
			return list
		} else {
			slog.Warn("drifts DB list failed, falling back to memory", "error", err)
		}
	}
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
	if s.dbPool != nil {
		if ok, err := s.hasPendingDB(repository, driftType, filePath); err == nil {
			return ok
		} else {
			slog.Warn("drift pending DB check failed, falling back to memory", "error", err)
		}
	}
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

// Get returns a single drift by ID.
func (s *Store) Get(id string) (*Drift, bool) {
	if s.dbPool != nil {
		if d, err := s.getDriftDB(id); err == nil {
			return d, true
		} else {
			slog.Warn("drift DB read failed, falling back to memory", "drift_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.items {
		if s.items[i].ID == id {
			d := s.items[i]
			return &d, true
		}
	}
	return nil, false
}

// UpdateStatus transitions a drift to pending, fixed, or ignored.
func (s *Store) UpdateStatus(id, status string) (*Drift, error) {
	switch status {
	case StatusPending, StatusFixed, StatusIgnored:
	default:
		return nil, fmt.Errorf("invalid status: %q", status)
	}
	if s.dbPool != nil {
		if d, err := s.getDriftDB(id); err == nil {
			d.Status = status
			d.UpdatedAt = time.Now()
			if err := s.persistDriftDB(d); err != nil {
				slog.Warn("drift status persistence failed", "drift_id", id, "error", err)
			}
			return d, nil
		}
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
