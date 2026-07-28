package planning

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// MemoryStore is the in-process implementation of Store used for local
// development, tests, and the database-unavailable fallback path. It preserves
// insertion order for list operations (newest first), matching the DBStore's
// ORDER BY created_at DESC.
type MemoryStore struct {
	mu        sync.RWMutex
	drafts    map[string]*DraftPlan
	cases     map[string]*TestCase
	lists     map[string]*TestList
	proposals map[string]*ChangeProposal
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		drafts:    make(map[string]*DraftPlan),
		cases:     make(map[string]*TestCase),
		lists:     make(map[string]*TestList),
		proposals: make(map[string]*ChangeProposal),
	}
}

func (s *MemoryStore) CreateDraft(_ context.Context, p *DraftPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepareDraftCreate(p)
	s.drafts[p.ID] = cloneDraft(p)
	return nil
}

func (s *MemoryStore) GetDraft(_ context.Context, id string) (*DraftPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.drafts[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return cloneDraft(p), nil
}

func (s *MemoryStore) UpdateDraft(_ context.Context, p *DraftPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.drafts[p.ID]; !ok {
		return pgx.ErrNoRows
	}
	ensureCaseIDs(p.Cases)
	p.UpdatedAt = time.Now()
	s.drafts[p.ID] = cloneDraft(p)
	return nil
}

func (s *MemoryStore) CreateTestCases(_ context.Context, cases []*TestCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, tc := range cases {
		prepareTestCaseCreate(tc)
		s.cases[tc.ID] = cloneCase(tc)
	}
	return nil
}

func (s *MemoryStore) ListTestCases(_ context.Context, projectID string) ([]*TestCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*TestCase, 0, len(s.cases))
	for _, tc := range s.cases {
		if projectID != "" && tc.ProjectID != projectID {
			continue
		}
		out = append(out, cloneCase(tc))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) GetTestCase(_ context.Context, id string) (*TestCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tc, ok := s.cases[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return cloneCase(tc), nil
}

func (s *MemoryStore) UpdateTestCase(_ context.Context, tc *TestCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cases[tc.ID]; !ok {
		return pgx.ErrNoRows
	}
	tc.UpdatedAt = time.Now()
	s.cases[tc.ID] = cloneCase(tc)
	return nil
}

func (s *MemoryStore) CreateTestList(_ context.Context, l *TestList) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepareTestListCreate(l)
	s.lists[l.ID] = cloneList(l)
	return nil
}

func (s *MemoryStore) ListTestLists(_ context.Context, projectID string) ([]*TestList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*TestList, 0, len(s.lists))
	for _, l := range s.lists {
		if projectID != "" && l.ProjectID != projectID {
			continue
		}
		out = append(out, cloneList(l))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) GetTestList(_ context.Context, id string) (*TestList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.lists[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return cloneList(l), nil
}

func (s *MemoryStore) CreateChangeProposal(_ context.Context, p *ChangeProposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepareProposalCreate(p)
	s.proposals[p.ID] = cloneProposal(p)
	return nil
}

func (s *MemoryStore) GetChangeProposal(_ context.Context, id string) (*ChangeProposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.proposals[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return cloneProposal(p), nil
}

func (s *MemoryStore) ListChangeProposals(_ context.Context, testCaseID string) ([]*ChangeProposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ChangeProposal, 0, len(s.proposals))
	for _, p := range s.proposals {
		if testCaseID != "" && p.TestCaseID != testCaseID {
			continue
		}
		out = append(out, cloneProposal(p))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) UpdateChangeProposal(_ context.Context, p *ChangeProposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.proposals[p.ID]; !ok {
		return pgx.ErrNoRows
	}
	p.UpdatedAt = time.Now()
	s.proposals[p.ID] = cloneProposal(p)
	return nil
}

// --- deep-copy helpers keep stored values isolated from caller mutation,
// so a returned pointer cannot be aliased and modified in place. ---

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneDraft(p *DraftPlan) *DraftPlan {
	cp := *p
	cp.Cases = make([]DraftCase, len(p.Cases))
	for i, c := range p.Cases {
		c.Steps = cloneStrings(c.Steps)
		c.Assertions = cloneStrings(c.Assertions)
		c.Tags = cloneStrings(c.Tags)
		cp.Cases[i] = c
	}
	return &cp
}

func cloneCase(tc *TestCase) *TestCase {
	cp := *tc
	cp.Steps = cloneStrings(tc.Steps)
	cp.Assertions = cloneStrings(tc.Assertions)
	cp.Tags = cloneStrings(tc.Tags)
	return &cp
}

func cloneList(l *TestList) *TestList {
	cp := *l
	cp.Tags = cloneStrings(l.Tags)
	cp.TestCaseIDs = cloneStrings(l.TestCaseIDs)
	return &cp
}

func cloneProposal(p *ChangeProposal) *ChangeProposal {
	cp := *p
	cp.Original = *cloneCase(&p.Original)
	cp.Proposed = *cloneCase(&p.Proposed)
	if p.ReviewedAt != nil {
		t := *p.ReviewedAt
		cp.ReviewedAt = &t
	}
	return &cp
}
