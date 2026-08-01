package workflow

import (
	"context"
	"encoding/json"
	"time"
)

func workflowDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *ReviewStore) persistReviewDB(rev *Review) error {
	ctx, cancel := workflowDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO reviews (id, run_id, type, status, reviewer, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			run_id = EXCLUDED.run_id,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			reviewer = EXCLUDED.reviewer,
			comment = EXCLUDED.comment,
			updated_at = EXCLUDED.updated_at`,
		rev.ID, rev.RunID, rev.Type, rev.Status, rev.Reviewer, rev.Comment, rev.CreatedAt, rev.UpdatedAt)
	return err
}

func (s *ReviewStore) getReviewDB(id string) (*Review, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, run_id, type, status, reviewer, comment, created_at, updated_at
		FROM reviews WHERE id = $1`, id)
	return scanReview(row)
}

func (s *ReviewStore) listReviewsByRunDB(runID string) ([]*Review, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, run_id, type, status, reviewer, comment, created_at, updated_at
		FROM reviews WHERE run_id = $1 ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Review
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return result, err
		}
		result = append(result, rev)
	}
	return result, rows.Err()
}

func scanReview(row interface{ Scan(dest ...any) error }) (*Review, error) {
	var rev Review
	if err := row.Scan(&rev.ID, &rev.RunID, &rev.Type, &rev.Status, &rev.Reviewer, &rev.Comment, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
		return nil, err
	}
	return &rev, nil
}

func (s *SuiteStore) persistSuiteDB(suite *Suite) error {
	ctx, cancel := workflowDBContext()
	defer cancel()
	tags, _ := json.Marshal(suite.Tags)
	tagsJSON := string(tags)
	if tagsJSON == "" || tagsJSON == "null" {
		tagsJSON = "[]"
	}
	runIDs, _ := json.Marshal(suite.RunIDs)
	runIDsJSON := string(runIDs)
	if runIDsJSON == "" || runIDsJSON == "null" {
		runIDsJSON = "[]"
	}
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO suites (id, name, project_id, environment, tags, pinned, run_ids, created_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7::jsonb, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			project_id = EXCLUDED.project_id,
			environment = EXCLUDED.environment,
			tags = EXCLUDED.tags,
			pinned = EXCLUDED.pinned,
			run_ids = EXCLUDED.run_ids`,
		suite.ID, suite.Name, suite.ProjectID, suite.Environment, tagsJSON, suite.Pinned, runIDsJSON, suite.CreatedAt)
	return err
}

func (s *SuiteStore) getSuiteDB(id string) (*Suite, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, name, project_id, environment, tags, pinned, run_ids, created_at
		FROM suites WHERE id = $1`, id)
	return scanSuite(row)
}

func (s *SuiteStore) listSuitesDB() ([]*Suite, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, name, project_id, environment, tags, pinned, run_ids, created_at
		FROM suites ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Suite
	for rows.Next() {
		suite, err := scanSuite(rows)
		if err != nil {
			return result, err
		}
		result = append(result, suite)
	}
	return result, rows.Err()
}

func (s *SuiteStore) listSuitesByTagDB(tag string) ([]*Suite, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, name, project_id, environment, tags, pinned, run_ids, created_at
		FROM suites WHERE tags @> $1::jsonb ORDER BY created_at DESC`, mustJSONList(tag))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Suite
	for rows.Next() {
		suite, err := scanSuite(rows)
		if err != nil {
			return result, err
		}
		result = append(result, suite)
	}
	return result, rows.Err()
}

func scanSuite(row interface{ Scan(dest ...any) error }) (*Suite, error) {
	var suite Suite
	var tags []byte
	var runIDs []byte
	if err := row.Scan(&suite.ID, &suite.Name, &suite.ProjectID, &suite.Environment, &tags, &suite.Pinned, &runIDs, &suite.CreatedAt); err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &suite.Tags)
	}
	if len(runIDs) > 0 {
		_ = json.Unmarshal(runIDs, &suite.RunIDs)
	}
	return &suite, nil
}

func (s *SuiteStore) deleteSuiteDB(id string) (bool, error) {
	ctx, cancel := workflowDBContext()
	defer cancel()
	cmd, err := s.dbPool.Exec(ctx, `DELETE FROM suites WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func mustJSONList(value string) string {
	b, err := json.Marshal([]string{value})
	if err != nil {
		return "[]"
	}
	return string(b)
}
