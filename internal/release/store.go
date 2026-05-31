package release

import (
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
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
}

func NewStore() *Store {
	return &Store{releases: make(map[string]*Release)}
}

func (s *Store) Create(r *Release) *Release {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	if r.Status == "" {
		r.Status = "active"
	}
	s.releases[r.ID] = r
	s.order = append([]string{r.ID}, s.order...)
	return r
}

func (s *Store) Get(id string) (*Release, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.releases[id]
	return r, ok
}

func (s *Store) List() []*Release {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Release
	for _, id := range s.order {
		if r, ok := s.releases[id]; ok {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) Update(id string, fn func(*Release)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.releases[id]
	if !ok {
		return false
	}
	fn(r)
	r.UpdatedAt = time.Now()
	return true
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
