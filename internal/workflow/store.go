// Package workflow menyediakan human-in-the-loop review/approval dan suite management.
package workflow

import (
	"sync"
	"time"

	"github.com/google/uuid"
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
}

func NewReviewStore() *ReviewStore {
	return &ReviewStore{reviews: make(map[string]*Review), byRun: make(map[string][]string)}
}

func (s *ReviewStore) Create(r *Review) *Review {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.Status = Pending
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	s.reviews[r.ID] = r
	s.byRun[r.RunID] = append(s.byRun[r.RunID], r.ID)
	return r
}

func (s *ReviewStore) Get(id string) (*Review, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.reviews[id]
	return r, ok
}

func (s *ReviewStore) ByRun(runID string) []*Review {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Review
	for _, id := range s.byRun[runID] {
		if r, ok := s.reviews[id]; ok {
			result = append(result, r)
		}
	}
	return result
}

func (s *ReviewStore) Approve(id, reviewer, comment string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.reviews[id]
	if !ok {
		return false
	}
	r.Status = Approved
	r.Reviewer = reviewer
	r.Comment = comment
	r.UpdatedAt = time.Now()
	return true
}

func (s *ReviewStore) Reject(id, reviewer, comment string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.reviews[id]
	if !ok {
		return false
	}
	r.Status = Rejected
	r.Reviewer = reviewer
	r.Comment = comment
	r.UpdatedAt = time.Now()
	return true
}

// --- Suite/Tag Management ---

type Suite struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ProjectID   string   `json:"project_id,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Tags        []string `json:"tags"`
	Pinned      bool     `json:"pinned"`
	RunIDs      []string `json:"run_ids"`
	CreatedAt   time.Time `json:"created_at"`
}

type SuiteStore struct {
	mu     sync.RWMutex
	suites map[string]*Suite
	order  []string
}

func NewSuiteStore() *SuiteStore {
	return &SuiteStore{suites: make(map[string]*Suite)}
}

func (s *SuiteStore) Create(suite *Suite) *Suite {
	s.mu.Lock()
	defer s.mu.Unlock()
	if suite.ID == "" {
		suite.ID = uuid.New().String()
	}
	suite.CreatedAt = time.Now()
	s.suites[suite.ID] = suite
	s.order = append([]string{suite.ID}, s.order...)
	return suite
}

func (s *SuiteStore) Get(id string) (*Suite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	suite, ok := s.suites[id]
	return suite, ok
}

func (s *SuiteStore) List() []*Suite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Suite
	for _, id := range s.order {
		if suite, ok := s.suites[id]; ok {
			result = append(result, suite)
		}
	}
	return result
}

func (s *SuiteStore) ByTag(tag string) []*Suite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Suite
	for _, suite := range s.suites {
		for _, t := range suite.Tags {
			if t == tag {
				result = append(result, suite)
				break
			}
		}
	}
	return result
}

func (s *SuiteStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suites[id]; !ok {
		return false
	}
	delete(s.suites, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}
