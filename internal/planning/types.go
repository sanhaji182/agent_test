// Package planning stores the reusable test-design lifecycle: draft plans,
// approved test cases, test lists, and review-gated change proposals.
//
// It mirrors the persistence convention used by internal/project and
// internal/db: a Store interface with an in-memory implementation for
// local/fallback use and a PostgreSQL-backed implementation selected when a
// database pool is available. Schema authority is internal/db/migrations
// (004_test_planning_review, 005_test_lists, 008_change_proposals).
package planning

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DraftCase is a single proposed test case inside a draft plan. It is stored
// as JSON inside test_plan_drafts.cases and is not a standalone table row.
type DraftCase struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Feature    string   `json:"feature"`
	Priority   string   `json:"priority"`
	Enabled    bool     `json:"enabled"`
	Steps      []string `json:"steps"`
	Assertions []string `json:"assertions"`
	Tags       []string `json:"tags"`
	Confidence float64  `json:"confidence"`
}

// DraftPlan is a generated, editable set of proposed cases for a project. It
// maps to the test_plan_drafts table.
type DraftPlan struct {
	ID        string      `json:"id"`
	ProjectID string      `json:"project_id"`
	Status    string      `json:"status"`
	Cases     []DraftCase `json:"cases"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// TestCase is an approved, versioned reusable test case. It maps to the
// test_cases table.
type TestCase struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	PlanID     string    `json:"plan_id,omitempty"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	Feature    string    `json:"feature"`
	Priority   string    `json:"priority"`
	Steps      []string  `json:"steps"`
	Assertions []string  `json:"assertions"`
	Tags       []string  `json:"tags"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TestList is a named grouping of approved test-case IDs used for execution
// and scheduling. It maps to the test_lists table. Membership is stored as a
// JSON array of IDs rather than a junction table (schema-faithful).
type TestList struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ProjectID   string    `json:"project_id,omitempty"`
	Tags        []string  `json:"tags"`
	TestCaseIDs []string  `json:"test_case_ids"`
	Pinned      bool      `json:"pinned"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChangeProposal is a review-gated refinement of an approved test case. It
// maps to the change_proposals table. Original and Proposed embed a full
// TestCase snapshot as JSON.
type ChangeProposal struct {
	ID            string     `json:"id"`
	TestCaseID    string     `json:"test_case_id"`
	Status        string     `json:"status"`
	Prompt        string     `json:"prompt"`
	Rationale     string     `json:"rationale"`
	Original      TestCase   `json:"original"`
	Proposed      TestCase   `json:"proposed"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	Reviewer      string     `json:"reviewer,omitempty"`
	ReviewComment string     `json:"review_comment,omitempty"`
}

// Store is the persistence boundary for the reusable test-design lifecycle.
// A "" filter argument to ListTestCases/ListTestLists/ListChangeProposals
// means "all rows". Get* returns pgx.ErrNoRows when the row is absent.
type Store interface {
	CreateDraft(context.Context, *DraftPlan) error
	GetDraft(context.Context, string) (*DraftPlan, error)
	UpdateDraft(context.Context, *DraftPlan) error

	CreateTestCases(context.Context, []*TestCase) error
	ListTestCases(context.Context, string) ([]*TestCase, error)
	GetTestCase(context.Context, string) (*TestCase, error)
	UpdateTestCase(context.Context, *TestCase) error

	CreateTestList(context.Context, *TestList) error
	ListTestLists(context.Context, string) ([]*TestList, error)
	GetTestList(context.Context, string) (*TestList, error)

	CreateChangeProposal(context.Context, *ChangeProposal) error
	GetChangeProposal(context.Context, string) (*ChangeProposal, error)
	ListChangeProposals(context.Context, string) ([]*ChangeProposal, error)
	UpdateChangeProposal(context.Context, *ChangeProposal) error
}

// --- shared preparation helpers (ID/default/timestamp assignment) ---

func ensureCaseIDs(cases []DraftCase) {
	for i := range cases {
		if cases[i].ID == "" {
			cases[i].ID = uuid.New().String()
		}
	}
}

func prepareDraftCreate(p *DraftPlan) {
	now := time.Now()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Status == "" {
		p.Status = "draft"
	}
	if p.Cases == nil {
		p.Cases = []DraftCase{}
	}
	ensureCaseIDs(p.Cases)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
}

func prepareTestCaseCreate(tc *TestCase) {
	now := time.Now()
	if tc.ID == "" {
		tc.ID = uuid.New().String()
	}
	if tc.Type == "" {
		tc.Type = "ui"
	}
	if tc.Priority == "" {
		tc.Priority = "medium"
	}
	if tc.Version == 0 {
		tc.Version = 1
	}
	if tc.CreatedAt.IsZero() {
		tc.CreatedAt = now
	}
	if tc.UpdatedAt.IsZero() {
		tc.UpdatedAt = now
	}
}

func prepareTestListCreate(l *TestList) {
	now := time.Now()
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.Tags == nil {
		l.Tags = []string{}
	}
	if l.TestCaseIDs == nil {
		l.TestCaseIDs = []string{}
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = now
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = now
	}
}

func prepareProposalCreate(p *ChangeProposal) {
	now := time.Now()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Status == "" {
		p.Status = "pending"
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
}

// nullableUUID returns nil for empty strings so that nullable UUID FK columns
// receive NULL rather than an invalid empty UUID literal.
func nullableUUID(s string) any {
	if s == "" {
		return nil
	}
	return s
}
