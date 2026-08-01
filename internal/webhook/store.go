package webhook

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookRegistration stores a registered GitHub webhook for continuous sync.
type WebhookRegistration struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	RepositoryID  string     `json:"repository_id"`
	RepositoryURL string     `json:"repository_url"`
	WebhookID     string     `json:"webhook_id"`
	Status        string     `json:"status"` // "active", "inactive"
	Secret        string     `json:"secret,omitempty"`
	GithubToken   string     `json:"github_token,omitempty"`
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// RegistrationStore is an in-memory store for webhook registrations.
type RegistrationStore struct {
	mu      sync.RWMutex
	items   []WebhookRegistration
	counter int64
	dbPool  *pgxpool.Pool
}

func NewRegistrationStore() *RegistrationStore {
	return &RegistrationStore{}
}

// EnableDB enables PostgreSQL persistence for webhook registrations.
func (s *RegistrationStore) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

func (s *RegistrationStore) Create(r WebhookRegistration) *WebhookRegistration {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "active"
	}
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	s.items = append(s.items, r)
	result := s.items[len(s.items)-1]
	if s.dbPool != nil {
		if err := s.persistRegistrationDB(&result); err != nil {
			slog.Warn("webhook registration persistence failed", "webhook_id", result.ID, "error", err)
		}
	}
	return &result
}

func (s *RegistrationStore) List() []WebhookRegistration {
	if s.dbPool != nil {
		if regs, err := s.listRegistrationsDB(); err == nil {
			return regs
		} else {
			slog.Warn("webhook registrations DB list failed, falling back to memory", "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]WebhookRegistration{}, s.items...)
}

func (s *RegistrationStore) Get(id string) (*WebhookRegistration, bool) {
	if s.dbPool != nil {
		if reg, err := s.getRegistrationDB(id); err == nil {
			return reg, true
		} else {
			slog.Warn("webhook registration DB read failed, falling back to memory", "webhook_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.items {
		if s.items[i].ID == id {
			r := s.items[i]
			return &r, true
		}
	}
	return nil, false
}

func (s *RegistrationStore) UpdateStatus(id, status string) (*WebhookRegistration, error) {
	switch status {
	case "active", "inactive":
	default:
		return nil, fmt.Errorf("invalid status: %q", status)
	}
	if s.dbPool != nil {
		if reg, err := s.getRegistrationDB(id); err == nil {
			reg.Status = status
			reg.UpdatedAt = time.Now()
			if err := s.persistRegistrationDB(reg); err != nil {
				slog.Warn("webhook status persistence failed", "webhook_id", id, "error", err)
			}
			return reg, nil
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Status = status
			s.items[i].UpdatedAt = time.Now()
			r := s.items[i]
			return &r, nil
		}
	}
	return nil, fmt.Errorf("webhook registration not found: %s", id)
}

func (s *RegistrationStore) Delete(id string) bool {
	if s.dbPool != nil {
		if _, ok := s.Get(id); !ok {
			return false
		}
		if err := s.deleteRegistrationDB(id); err != nil {
			slog.Warn("webhook registration DB delete failed", "webhook_id", id, "error", err)
			return false
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return true
		}
	}
	return s.dbPool != nil
}

func (s *RegistrationStore) UpdateLastSync(id string) bool {
	now := time.Now()
	if s.dbPool != nil {
		if reg, err := s.getRegistrationDB(id); err == nil {
			reg.LastSyncAt = &now
			reg.UpdatedAt = now
			if err := s.persistRegistrationDB(reg); err != nil {
				slog.Warn("webhook last_sync persistence failed", "webhook_id", id, "error", err)
				return false
			}
			return true
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].LastSyncAt = &now
			s.items[i].UpdatedAt = now
			return true
		}
	}
	return false
}
