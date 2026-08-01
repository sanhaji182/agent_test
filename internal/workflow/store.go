// Package workflow menyediakan human-in-the-loop review/approval dan suite management.
package workflow

import (
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Review/Approval ---

type ReviewStatus string

const (
	Pending  ReviewStatus = "pending"
	Approved ReviewStatus = "approved"
	Rejected ReviewStatus = "rejected"
)

type Review struct {
	ID        string       `json:"id"`
	RunID     string       `json:"run_id"`
	Type      string       `json:"type"` // "test_plan", "test_scripts", "fix_suggestion"
	Status    ReviewStatus `json:"status"`
	Reviewer  string       `json:"reviewer,omitempty"`
	Comment   string       `json:"comment,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ReviewStore struct {
	mu      sync.RWMutex
	reviews map[string]*Review
	byRun   map[string][]string
	dbPool  *pgxpool.Pool
}

func NewReviewStore() *ReviewStore {
	return &ReviewStore{reviews: make(map[string]*Review), byRun: make(map[string][]string)}
}

func (s *ReviewStore) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

func (s *ReviewStore) Create(r *Review) *Review {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.Status = Pending
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = now
	}
	stored := *r
	s.reviews[r.ID] = &stored
	s.byRun[r.RunID] = append(s.byRun[r.RunID], r.ID)
	if s.dbPool != nil {
		if err := s.persistReviewDB(&stored); err != nil {
			slog.Warn("review persistence failed", "review_id", stored.ID, "error", err)
		}
	}
	return &stored
}

func (s *ReviewStore) Get(id string) (*Review, bool) {
	if s.dbPool != nil {
		if rev, err := s.getReviewDB(id); err == nil {
			return rev, true
		} else {
			slog.Warn("review DB read failed, falling back to memory", "review_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.reviews[id]
	if !ok {
		return nil, false
	}
	copied := *r
	return &copied, true
}

func (s *ReviewStore) ByRun(runID string) []*Review {
	if s.dbPool != nil {
		if revs, err := s.listReviewsByRunDB(runID); err == nil {
			return revs
		} else {
			slog.Warn("review DB list failed, falling back to memory", "run_id", runID, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Review
	for _, id := range s.byRun[runID] {
		if r, ok := s.reviews[id]; ok {
			copied := *r
			result = append(result, &copied)
		}
	}
	return result
}

func (s *ReviewStore) Approve(id, reviewer, comment string) bool {
	return s.updateStatus(id, Approved, reviewer, comment)
}

func (s *ReviewStore) Reject(id, reviewer, comment string) bool {
	return s.updateStatus(id, Rejected, reviewer, comment)
}

func (s *ReviewStore) updateStatus(id string, status ReviewStatus, reviewer, comment string) bool {
	s.mu.Lock()
	r, ok := s.reviews[id]
	if ok {
		r.Status = status
		r.Reviewer = reviewer
		r.Comment = comment
		r.UpdatedAt = time.Now()
		copied := *r
		s.mu.Unlock()
		if s.dbPool != nil {
			if err := s.persistReviewDB(&copied); err != nil {
				slog.Warn("review update persistence failed", "review_id", id, "error", err)
			}
		}
		return true
	}
	s.mu.Unlock()

	if s.dbPool == nil {
		return false
	}
	rev, err := s.getReviewDB(id)
	if err != nil {
		slog.Warn("review DB update read failed", "review_id", id, "error", err)
		return false
	}
	rev.Status = status
	rev.Reviewer = reviewer
	rev.Comment = comment
	rev.UpdatedAt = time.Now()
	if err := s.persistReviewDB(rev); err != nil {
		slog.Warn("review update persistence failed", "review_id", id, "error", err)
		return false
	}
	s.rememberReview(rev)
	return true
}

func (s *ReviewStore) rememberReview(r *Review) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := *r
	if _, ok := s.reviews[r.ID]; !ok {
		s.byRun[r.RunID] = append(s.byRun[r.RunID], r.ID)
	}
	s.reviews[r.ID] = &stored
}

// --- Suite/Tag Management ---

type Suite struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ProjectID   string    `json:"project_id,omitempty"`
	Environment string    `json:"environment,omitempty"`
	Tags        []string  `json:"tags"`
	Pinned      bool      `json:"pinned"`
	RunIDs      []string  `json:"run_ids"`
	CreatedAt   time.Time `json:"created_at"`
}

type SuiteStore struct {
	mu     sync.RWMutex
	suites map[string]*Suite
	order  []string
	dbPool *pgxpool.Pool
}

func NewSuiteStore() *SuiteStore {
	return &SuiteStore{suites: make(map[string]*Suite)}
}

func (s *SuiteStore) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
}

func (s *SuiteStore) Create(suite *Suite) *Suite {
	s.mu.Lock()
	defer s.mu.Unlock()
	if suite.ID == "" {
		suite.ID = uuid.New().String()
	}
	if suite.CreatedAt.IsZero() {
		suite.CreatedAt = time.Now()
	}
	stored := *suite
	s.suites[suite.ID] = &stored
	s.order = append([]string{suite.ID}, s.order...)
	if s.dbPool != nil {
		if err := s.persistSuiteDB(&stored); err != nil {
			slog.Warn("suite persistence failed", "suite_id", stored.ID, "error", err)
		}
	}
	return &stored
}

func (s *SuiteStore) Get(id string) (*Suite, bool) {
	if s.dbPool != nil {
		if suite, err := s.getSuiteDB(id); err == nil {
			return suite, true
		} else {
			slog.Warn("suite DB read failed, falling back to memory", "suite_id", id, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	suite, ok := s.suites[id]
	if !ok {
		return nil, false
	}
	copied := *suite
	return &copied, true
}

func (s *SuiteStore) List() []*Suite {
	if s.dbPool != nil {
		if suites, err := s.listSuitesDB(); err == nil {
			return suites
		} else {
			slog.Warn("suites DB list failed, falling back to memory", "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Suite
	for _, id := range s.order {
		if suite, ok := s.suites[id]; ok {
			copied := *suite
			result = append(result, &copied)
		}
	}
	return result
}

func (s *SuiteStore) ByTag(tag string) []*Suite {
	if s.dbPool != nil {
		if suites, err := s.listSuitesByTagDB(tag); err == nil {
			return suites
		} else {
			slog.Warn("suites DB tag query failed, falling back to memory", "tag", tag, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Suite
	for _, suite := range s.suites {
		for _, t := range suite.Tags {
			if t == tag {
				copied := *suite
				result = append(result, &copied)
				break
			}
		}
	}
	return result
}

func (s *SuiteStore) Delete(id string) bool {
	s.mu.Lock()
	_, ok := s.suites[id]
	if ok {
		delete(s.suites, id)
		for i, oid := range s.order {
			if oid == id {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
	s.mu.Unlock()

	if s.dbPool != nil {
		deleted, err := s.deleteSuiteDB(id)
		if err == nil && deleted {
			return true
		}
		if err != nil {
			slog.Warn("suite delete persistence failed", "suite_id", id, "error", err)
		}
	}
	return ok
}
