// Package audit provides an append-only audit log for security-relevant actions.
// Every create/update/delete/approve/reject action records who did it, when,
// and what resource was affected.
package audit

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Action categorizes the type of audited event.
type Action string

const (
	ActionCreate         Action = "create"
	ActionUpdate         Action = "update"
	ActionDelete         Action = "delete"
	ActionApprove        Action = "approve"
	ActionReject         Action = "reject"
	ActionLogin          Action = "login"
	ActionLogout         Action = "logout"
	ActionExport         Action = "export"
	ActionSettingsChange Action = "settings_change"
	ActionRunStart       Action = "run_start"
	ActionRunComplete    Action = "run_complete"
	ActionDriftDetected  Action = "drift_detected"
)

// Resource is the type of entity affected.
type Resource string

const (
	ResourceRun            Resource = "run"
	ResourceRelease        Resource = "release"
	ResourceReview         Resource = "review"
	ResourceSchedule       Resource = "schedule"
	ResourceSettings       Resource = "settings"
	ResourceRecording      Resource = "recording"
	ResourceProject        Resource = "project"
	ResourceTestPlan       Resource = "test_plan"
	ResourceTestCase       Resource = "test_case"
	ResourceTestList       Resource = "test_list"
	ResourceChangeProposal Resource = "change_proposal"
)

// Entry represents a single audit log entry.
type Entry struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	ActorRole  string    `json:"actor_role"`
	Action     Action    `json:"action"`
	Resource   Resource  `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Detail     string    `json:"detail,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Store is a thread-safe append-only audit log.
type Store struct {
	mu      sync.RWMutex
	entries []Entry
	dbPool  *pgxpool.Pool
}

// NewStore creates a new in-memory audit log store.
func NewStore() *Store {
	return &Store{}
}

// EnableDB enables PostgreSQL persistence for audit entries.
func (s *Store) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

// Record creates a new audit log entry.
func (s *Store) Record(actorID, actorRole string, action Action, resource Resource, resourceID string, detail string) *Entry {
	entry := Entry{
		ID:         uuid.New().String(),
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.entries = append(s.entries, entry)
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.persistDB(entry); err != nil {
			slog.Warn("audit: failed to persist entry", "entry_id", entry.ID, "error", err)
		}
	}

	slog.Info("audit",
		"actor", actorID,
		"role", actorRole,
		"action", action,
		"resource", fmt.Sprintf("%s/%s", resource, resourceID),
		"detail", detail,
	)

	return &entry
}

// List returns recent audit entries, ordered newest first.
func (s *Store) List(limit int) []Entry {
	if s.dbPool != nil {
		if entries, listErr := s.listDB(limit); listErr == nil {
			return entries
		} else {
			slog.Warn("audit: DB list failed, falling back to memory", "error", listErr)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.entries) {
		limit = len(s.entries)
	}

	result := make([]Entry, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.entries[len(s.entries)-1-i]
	}
	return result
}

// ListByActor returns audit entries for a specific actor.
func (s *Store) ListByActor(actorID string, limit int) []Entry {
	if s.dbPool != nil {
		if entries, listErr := s.listByActorDB(actorID, limit); listErr == nil {
			return entries
		} else {
			slog.Warn("audit: DB list by actor failed, falling back to memory", "actor_id", actorID, "error", listErr)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Entry
	for i := len(s.entries) - 1; i >= 0 && len(result) < limit; i-- {
		if s.entries[i].ActorID == actorID {
			result = append(result, s.entries[i])
		}
	}
	return result
}

// ListByResource returns audit entries for a specific resource.
func (s *Store) ListByResource(resource Resource, resourceID string, limit int) []Entry {
	if s.dbPool != nil {
		if entries, listErr := s.listByResourceDB(resource, resourceID, limit); listErr == nil {
			return entries
		} else {
			slog.Warn("audit: DB list by resource failed, falling back to memory", "resource", resource, "resource_id", resourceID, "error", listErr)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Entry
	for i := len(s.entries) - 1; i >= 0 && len(result) < limit; i-- {
		if s.entries[i].Resource == resource && s.entries[i].ResourceID == resourceID {
			result = append(result, s.entries[i])
		}
	}
	return result
}

// --- PostgreSQL persistence ---

func (s *Store) persistDB(entry Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO audit_log (id, actor_id, actor_role, action, resource, resource_id, detail, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.ID, entry.ActorID, entry.ActorRole, string(entry.Action),
		string(entry.Resource), entry.ResourceID, entry.Detail, entry.CreatedAt,
	)
	return err
}

func (s *Store) listDB(limit int) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.dbPool.Query(ctx, `
		SELECT id, actor_id, actor_role, action, resource, resource_id, detail, created_at
		FROM audit_log ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var action, resource string
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ActorRole, &action, &resource, &e.ResourceID, &e.Detail, &e.CreatedAt); err != nil {
			return entries, err
		}
		e.Action = Action(action)
		e.Resource = Resource(resource)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *Store) listByActorDB(actorID string, limit int) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.dbPool.Query(ctx, `
		SELECT id, actor_id, actor_role, action, resource, resource_id, detail, created_at
		FROM audit_log WHERE actor_id = $1 ORDER BY created_at DESC LIMIT $2`, actorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditEntries(rows)
}

func (s *Store) listByResourceDB(resource Resource, resourceID string, limit int) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.dbPool.Query(ctx, `
		SELECT id, actor_id, actor_role, action, resource, resource_id, detail, created_at
		FROM audit_log WHERE resource = $1 AND resource_id = $2 ORDER BY created_at DESC LIMIT $3`,
		string(resource), resourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditEntries(rows)
}

func scanAuditEntries(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]Entry, error) {
	var entries []Entry
	for rows.Next() {
		var e Entry
		var action, resource string
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ActorRole, &action, &resource, &e.ResourceID, &e.Detail, &e.CreatedAt); err != nil {
			return entries, err
		}
		e.Action = Action(action)
		e.Resource = Resource(resource)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
