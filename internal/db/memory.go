package db

import (
	"context"
	"sync"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// RunStore adalah interface untuk menyimpan dan mengambil data test run.
// Diimplementasikan oleh Store (PostgreSQL) dan MemoryStore (in-memory).
type RunStore interface {
	CreateRun(ctx context.Context, run *agent.TestRun) error
	UpdateRun(ctx context.Context, run *agent.TestRun) error
	GetRun(ctx context.Context, id string) (*agent.TestRun, error)
	ListRuns(ctx context.Context, limit, offset int) ([]*agent.TestRun, error)
	DeleteRun(ctx context.Context, id string) error
}

// MemoryStore adalah implementasi RunStore di memori (tanpa database).
// Digunakan untuk development tanpa PostgreSQL.
type MemoryStore struct {
	mu    sync.RWMutex
	runs  map[string]*agent.TestRun
	order []string // Urutan ID dari terbaru ke terlama
}

// NewMemoryStore membuat store in-memory baru
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{runs: make(map[string]*agent.TestRun)}
}

// cloneRun returns a deep-enough copy of a run so readers never share mutable
// state with the execution goroutine (matches DB-store snapshot semantics).
// Scalars are copied by value; mutated slices (Screenshots, TestFiles) and the
// RunResult (whose Failures are patched by screenshot capture) are duplicated.
func cloneRun(run *agent.TestRun) *agent.TestRun {
	if run == nil {
		return nil
	}
	c := *run
	if run.Screenshots != nil {
		c.Screenshots = append([]string(nil), run.Screenshots...)
	}
	if run.TestFiles != nil {
		c.TestFiles = append([]agent.TestFile(nil), run.TestFiles...)
	}
	if run.RunResult != nil {
		rr := *run.RunResult
		if run.RunResult.Failures != nil {
			rr.Failures = append([]agent.Failure(nil), run.RunResult.Failures...)
		}
		c.RunResult = &rr
	}
	if run.FinishedAt != nil {
		t := *run.FinishedAt
		c.FinishedAt = &t
	}
	return &c
}

// CreateRun menyimpan run baru ke memori
func (m *MemoryStore) CreateRun(_ context.Context, run *agent.TestRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = cloneRun(run)
	m.order = append([]string{run.ID}, m.order...) // Terbaru di depan
	return nil
}

// UpdateRun memperbarui data run yang sudah ada
func (m *MemoryStore) UpdateRun(_ context.Context, run *agent.TestRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = cloneRun(run)
	return nil
}

// GetRun mengambil run berdasarkan ID
func (m *MemoryStore) GetRun(_ context.Context, id string) (*agent.TestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	run, ok := m.runs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneRun(run), nil
}

// ListRuns menampilkan daftar run dengan pagination
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
			result = append(result, cloneRun(run))
		}
	}
	return result, nil
}

// DeleteRun menghapus run berdasarkan ID
func (m *MemoryStore) DeleteRun(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[id]; !ok {
		return ErrNotFound
	}
	delete(m.runs, id)
	for i, oid := range m.order {
		if oid == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	return nil
}
