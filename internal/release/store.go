package release

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Release struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	ProjectID string    `json:"project_id"`
	Status    string    `json:"status"` // "active", "completed", "cancelled"
	RunIDs    []string  `json:"run_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Summary struct {
	ReleaseID    string  `json:"release_id"`
	TotalRuns    int     `json:"total_runs"`
	PassedRuns   int     `json:"passed_runs"`
	FailedRuns   int     `json:"failed_runs"`
	PassRate     float64 `json:"pass_rate"`
	TotalTests   int     `json:"total_tests"`
	TotalPassed  int     `json:"total_passed"`
	TotalFailed  int     `json:"total_failed"`
	LatestStatus string  `json:"latest_status"`
}

type Store struct {
	mu       sync.RWMutex
	releases map[string]*Release
	order    []string
	dbPool   *pgxpool.Pool
}

func NewStore() *Store {
	return &Store{releases: make(map[string]*Release)}
}

func (s *Store) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

func (s *Store) Create(r *Release) *Release {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = now
	}
	if r.Status == "" {
		r.Status = "active"
	}
	stored := *r
	s.releases[r.ID] = &stored
	s.order = append([]string{r.ID}, s.order...)
	if s.dbPool != nil {
		if err := s.persistReleaseDB(&stored); err != nil {
			slog.Warn("release persistence failed", "release_id", stored.ID, "error", err)
		}
	}
	return &stored
}

func (s *Store) Get(id string) (*Release, bool) {
	if s.dbPool != nil {
		if rel, err := s.getReleaseDB(id); err == nil {
			return rel, true
		} else {
			slog.Warn("release DB read failed, falling back to memory", "release_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.releases[id]
	if !ok {
		return nil, false
	}
	copied := *r
	return &copied, true
}

func (s *Store) List() []*Release {
	if s.dbPool != nil {
		if rels, err := s.listReleasesDB(); err == nil {
			return rels
		} else {
			slog.Warn("releases DB list failed, falling back to memory", "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Release
	for _, id := range s.order {
		if r, ok := s.releases[id]; ok {
			copied := *r
			result = append(result, &copied)
		}
	}
	return result
}

func (s *Store) Update(id string, fn func(*Release)) bool {
	s.mu.Lock()
	r, ok := s.releases[id]
	if ok {
		fn(r)
		r.UpdatedAt = time.Now()
		copied := *r
		s.mu.Unlock()
		if s.dbPool != nil {
			if err := s.persistReleaseDB(&copied); err != nil {
				slog.Warn("release update persistence failed", "release_id", id, "error", err)
			}
		}
		return true
	}
	s.mu.Unlock()

	if s.dbPool == nil {
		return false
	}
	rel, err := s.getReleaseDB(id)
	if err != nil {
		return false
	}
	fn(rel)
	rel.UpdatedAt = time.Now()
	if err := s.persistReleaseDB(rel); err != nil {
		slog.Warn("release update persistence failed", "release_id", id, "error", err)
		return false
	}
	s.rememberRelease(rel)
	return true
}

func (s *Store) rememberRelease(rel *Release) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := *rel
	if _, ok := s.releases[rel.ID]; !ok {
		s.order = append([]string{rel.ID}, s.order...)
	}
	s.releases[rel.ID] = &stored
}

// Summarize computes aggregated metrics for a release given its runs
func Summarize(rel *Release, runs []*agent.TestRun) *Summary {
	sum := &Summary{ReleaseID: rel.ID, TotalRuns: len(runs)}
	for _, r := range runs {
		if r.State == agent.StateDone {
			sum.PassedRuns++
		} else if r.State == agent.StateFailed {
			sum.FailedRuns++
		}
		if r.RunResult != nil {
			sum.TotalTests += r.RunResult.Total
			sum.TotalPassed += r.RunResult.Passed
			sum.TotalFailed += r.RunResult.Failed
		}
	}
	if sum.TotalRuns > 0 {
		sum.PassRate = float64(sum.PassedRuns) / float64(sum.TotalRuns)
	}
	if len(runs) > 0 {
		sum.LatestStatus = string(runs[0].State)
	}
	return sum
}
