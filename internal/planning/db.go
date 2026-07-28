package planning

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBStore is the PostgreSQL-backed implementation of Store. Column names and
// nullability follow internal/db/migrations/004, 005, and 008. JSON array
// columns (steps/assertions/tags/cases/test_case_ids) and the embedded
// TestCase snapshots in change_proposals are marshaled to JSONB.
type DBStore struct {
	pool *pgxpool.Pool
}

func NewDBStore(pool *pgxpool.Pool) *DBStore {
	return &DBStore{pool: pool}
}

type scanner interface {
	Scan(dest ...any) error
}

// --- Draft plans ---

func (s *DBStore) CreateDraft(ctx context.Context, p *DraftPlan) error {
	prepareDraftCreate(p)
	cases, err := json.Marshal(p.Cases)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO test_plan_drafts (id, project_id, status, cases, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID, nullableUUID(p.ProjectID), p.Status, cases, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *DBStore) GetDraft(ctx context.Context, id string) (*DraftPlan, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, COALESCE(project_id::text, ''), status, cases, created_at, updated_at
		FROM test_plan_drafts WHERE id = $1`, id)
	return scanDraft(row)
}

func (s *DBStore) UpdateDraft(ctx context.Context, p *DraftPlan) error {
	ensureCaseIDs(p.Cases)
	p.UpdatedAt = time.Now()
	cases, err := json.Marshal(p.Cases)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE test_plan_drafts SET status = $2, cases = $3, updated_at = $4
		WHERE id = $1`, p.ID, p.Status, cases, p.UpdatedAt)
	return err
}

func scanDraft(row scanner) (*DraftPlan, error) {
	var p DraftPlan
	var cases []byte
	if err := row.Scan(&p.ID, &p.ProjectID, &p.Status, &cases, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	if len(cases) > 0 {
		_ = json.Unmarshal(cases, &p.Cases)
	}
	if p.Cases == nil {
		p.Cases = []DraftCase{}
	}
	return &p, nil
}

// --- Test cases ---

func (s *DBStore) CreateTestCases(ctx context.Context, cases []*TestCase) error {
	if len(cases) == 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, tc := range cases {
		prepareTestCaseCreate(tc)
		steps, _ := json.Marshal(tc.Steps)
		assertions, _ := json.Marshal(tc.Assertions)
		tags, _ := json.Marshal(tc.Tags)
		if _, err := tx.Exec(ctx, `
			INSERT INTO test_cases (id, project_id, plan_id, title, type, feature,
				priority, steps, assertions, tags, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			tc.ID, nullableUUID(tc.ProjectID), nullableUUID(tc.PlanID), tc.Title, tc.Type,
			tc.Feature, tc.Priority, steps, assertions, tags, tc.Version, tc.CreatedAt, tc.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

const testCaseColumns = `id, COALESCE(project_id::text, ''), COALESCE(plan_id::text, ''),
	title, type, COALESCE(feature, ''), priority, steps, assertions, tags,
	version, created_at, updated_at`

func (s *DBStore) ListTestCases(ctx context.Context, projectID string) ([]*TestCase, error) {
	var rows pgx.Rows
	var err error
	if projectID == "" {
		rows, err = s.pool.Query(ctx, `SELECT `+testCaseColumns+` FROM test_cases ORDER BY created_at DESC`)
	} else {
		rows, err = s.pool.Query(ctx, `SELECT `+testCaseColumns+` FROM test_cases WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*TestCase{}
	for rows.Next() {
		tc, err := scanCase(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, tc)
	}
	return out, rows.Err()
}

func (s *DBStore) GetTestCase(ctx context.Context, id string) (*TestCase, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+testCaseColumns+` FROM test_cases WHERE id = $1`, id)
	return scanCase(row)
}

func (s *DBStore) UpdateTestCase(ctx context.Context, tc *TestCase) error {
	tc.UpdatedAt = time.Now()
	steps, _ := json.Marshal(tc.Steps)
	assertions, _ := json.Marshal(tc.Assertions)
	tags, _ := json.Marshal(tc.Tags)
	_, err := s.pool.Exec(ctx, `
		UPDATE test_cases SET title = $2, type = $3, feature = $4, priority = $5,
			steps = $6, assertions = $7, tags = $8, version = $9, updated_at = $10
		WHERE id = $1`,
		tc.ID, tc.Title, tc.Type, tc.Feature, tc.Priority, steps, assertions, tags, tc.Version, tc.UpdatedAt)
	return err
}

func scanCase(row scanner) (*TestCase, error) {
	var tc TestCase
	var steps, assertions, tags []byte
	if err := row.Scan(&tc.ID, &tc.ProjectID, &tc.PlanID, &tc.Title, &tc.Type,
		&tc.Feature, &tc.Priority, &steps, &assertions, &tags, &tc.Version,
		&tc.CreatedAt, &tc.UpdatedAt); err != nil {
		return nil, err
	}
	tc.Steps = unmarshalStrings(steps)
	tc.Assertions = unmarshalStrings(assertions)
	tc.Tags = unmarshalStrings(tags)
	return &tc, nil
}

// --- Test lists ---

func (s *DBStore) CreateTestList(ctx context.Context, l *TestList) error {
	prepareTestListCreate(l)
	tags, _ := json.Marshal(l.Tags)
	ids, _ := json.Marshal(l.TestCaseIDs)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO test_lists (id, name, project_id, tags, test_case_ids, pinned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		l.ID, l.Name, nullableUUID(l.ProjectID), tags, ids, l.Pinned, l.CreatedAt, l.UpdatedAt)
	return err
}

const testListColumns = `id, name, COALESCE(project_id::text, ''), tags, test_case_ids, pinned, created_at, updated_at`

func (s *DBStore) ListTestLists(ctx context.Context, projectID string) ([]*TestList, error) {
	var rows pgx.Rows
	var err error
	if projectID == "" {
		rows, err = s.pool.Query(ctx, `SELECT `+testListColumns+` FROM test_lists ORDER BY created_at DESC`)
	} else {
		rows, err = s.pool.Query(ctx, `SELECT `+testListColumns+` FROM test_lists WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*TestList{}
	for rows.Next() {
		l, err := scanList(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *DBStore) GetTestList(ctx context.Context, id string) (*TestList, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+testListColumns+` FROM test_lists WHERE id = $1`, id)
	return scanList(row)
}

func scanList(row scanner) (*TestList, error) {
	var l TestList
	var tags, ids []byte
	if err := row.Scan(&l.ID, &l.Name, &l.ProjectID, &tags, &ids, &l.Pinned, &l.CreatedAt, &l.UpdatedAt); err != nil {
		return nil, err
	}
	l.Tags = unmarshalStrings(tags)
	l.TestCaseIDs = unmarshalStrings(ids)
	return &l, nil
}

// --- Change proposals ---

func (s *DBStore) CreateChangeProposal(ctx context.Context, p *ChangeProposal) error {
	prepareProposalCreate(p)
	original, _ := json.Marshal(p.Original)
	proposed, _ := json.Marshal(p.Proposed)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO change_proposals (id, test_case_id, status, prompt, rationale,
			original, proposed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.TestCaseID, p.Status, p.Prompt, p.Rationale, original, proposed, p.CreatedAt, p.UpdatedAt)
	return err
}

const proposalColumns = `id, test_case_id, status, prompt, rationale, original, proposed,
	created_at, updated_at, reviewed_at, COALESCE(reviewer, ''), COALESCE(review_comment, '')`

func (s *DBStore) GetChangeProposal(ctx context.Context, id string) (*ChangeProposal, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+proposalColumns+` FROM change_proposals WHERE id = $1`, id)
	return scanProposal(row)
}

func (s *DBStore) ListChangeProposals(ctx context.Context, testCaseID string) ([]*ChangeProposal, error) {
	var rows pgx.Rows
	var err error
	if testCaseID == "" {
		rows, err = s.pool.Query(ctx, `SELECT `+proposalColumns+` FROM change_proposals ORDER BY created_at DESC`)
	} else {
		rows, err = s.pool.Query(ctx, `SELECT `+proposalColumns+` FROM change_proposals WHERE test_case_id = $1 ORDER BY created_at DESC`, testCaseID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*ChangeProposal{}
	for rows.Next() {
		p, err := scanProposal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *DBStore) UpdateChangeProposal(ctx context.Context, p *ChangeProposal) error {
	p.UpdatedAt = time.Now()
	original, _ := json.Marshal(p.Original)
	proposed, _ := json.Marshal(p.Proposed)
	_, err := s.pool.Exec(ctx, `
		UPDATE change_proposals SET status = $2, prompt = $3, rationale = $4,
			original = $5, proposed = $6, updated_at = $7, reviewed_at = $8,
			reviewer = $9, review_comment = $10
		WHERE id = $1`,
		p.ID, p.Status, p.Prompt, p.Rationale, original, proposed, p.UpdatedAt,
		p.ReviewedAt, p.Reviewer, p.ReviewComment)
	return err
}

func scanProposal(row scanner) (*ChangeProposal, error) {
	var p ChangeProposal
	var original, proposed []byte
	if err := row.Scan(&p.ID, &p.TestCaseID, &p.Status, &p.Prompt, &p.Rationale,
		&original, &proposed, &p.CreatedAt, &p.UpdatedAt, &p.ReviewedAt,
		&p.Reviewer, &p.ReviewComment); err != nil {
		return nil, err
	}
	if len(original) > 0 {
		_ = json.Unmarshal(original, &p.Original)
	}
	if len(proposed) > 0 {
		_ = json.Unmarshal(proposed, &p.Proposed)
	}
	return &p, nil
}

func unmarshalStrings(b []byte) []string {
	if len(b) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(b, &out); err != nil || out == nil {
		return []string{}
	}
	return out
}
