package db

import (
	"context"
	"sync"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// RunStore is the interface for persisting test runs.
type RunStore interface {
	CreateRun(ctx context.Context, run *agent.TestRun) error
	UpdateRun(ctx context.Context, run *agent.TestRun) error
	GetRun(ctx context.Context, id string) (*agent.TestRun, error)
	ListRuns(ctx context.Context, limit, offset int) ([]*agent.TestRun, error)
}

// MemoryStore is an in-memory RunStore for development without PostgreSQL.
type MemoryStore struct {
	mu   sync.RWMutex
	runs map[string]*agent.TestRun
	order []string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{runs: make(map[string]*agent.TestRun)}
}

func (m *MemoryStore) CreateRun(_ context.Context, run *agent.TestRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = run
	m.order = append([]string{run.ID}, m.order...)
	return nil
}

func (m *MemoryStore) UpdateRun(_ context.Context, run *agent.TestRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = run
	return nil
}

func (m *MemoryStore) GetRun(_ context.Context, id string) (*agent.TestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	run, ok := m.runs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return run, nil
}

func (m *MemoryStore) ListRuns(_ context.Context, limit, offset int) ([]*agent.TestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset >= len(m.order) {
		return nil, nil
	}
	end := offset + limit
	if end > len(m.order) {
		end = len(m.order)
	}

	var result []*agent.TestRun
	for _, id := range m.order[offset:end] {
		if run, ok := m.runs[id]; ok {
			result = append(result, run)
		}
	}
	return result, nil
}
