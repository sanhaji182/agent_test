// Package db menyediakan layer persistensi untuk GoTest Agent.
// Mendukung PostgreSQL (produksi) dan in-memory store (development).
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound dikembalikan ketika data tidak ditemukan
var ErrNotFound = errors.New("not found")

// Store adalah implementasi RunStore menggunakan PostgreSQL
type Store struct {
	pool *pgxpool.Pool
}

// NewStore membuat koneksi pool ke PostgreSQL
func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close menutup koneksi pool
func (s *Store) Close() {
	s.pool.Close()
}

// CreateRun menyimpan test run baru ke database
func (s *Store) CreateRun(ctx context.Context, run *agent.TestRun) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO test_runs (id, project_id, state, requirements, created_at, updated_at)
		VALUES ($1, NULL, $2, $3, $4, $5)`,
		run.ID, string(run.State), run.Requirements, run.CreatedAt, run.UpdatedAt)
	return err
}

// UpdateRun memperbarui data run yang sudah ada
func (s *Store) UpdateRun(ctx context.Context, run *agent.TestRun) error {
	testPlan, _ := json.Marshal(run.TestPlan)
	testFiles, _ := json.Marshal(run.TestFiles)
	runResult, _ := json.Marshal(run.RunResult)

	_, err := s.pool.Exec(ctx, `
		UPDATE test_runs SET
			state = $2, code_analysis = $3, test_plan = $4,
			test_files = $5, run_result = $6, fix_attempts = $7,
			error_msg = $8, updated_at = $9, finished_at = $10
		WHERE id = $1`,
		run.ID, string(run.State), run.CodeAnalysis,
		testPlan, testFiles, runResult,
		run.FixAttempts, run.Error, time.Now(), run.FinishedAt)
	return err
}

// GetRun mengambil detail test run berdasarkan ID
func (s *Store) GetRun(ctx context.Context, id string) (*agent.TestRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, state, requirements, code_analysis, test_plan, test_files,
			run_result, fix_attempts, error_msg, created_at, updated_at, finished_at
		FROM test_runs WHERE id = $1`, id)

	var run agent.TestRun
	var state string
	var testPlan, testFiles, runResult []byte
	var codeAnalysis, errorMsg *string
	var finishedAt *time.Time

	err := row.Scan(&run.ID, &state, &run.Requirements, &codeAnalysis,
		&testPlan, &testFiles, &runResult, &run.FixAttempts,
		&errorMsg, &run.CreatedAt, &run.UpdatedAt, &finishedAt)
	if err != nil {
		return nil, err
	}

	run.State = agent.State(state)
	if codeAnalysis != nil {
		run.CodeAnalysis = *codeAnalysis
	}
	if errorMsg != nil {
		run.Error = *errorMsg
	}
	run.FinishedAt = finishedAt

	// Unmarshal JSONB fields
	if len(testPlan) > 0 {
		json.Unmarshal(testPlan, &run.TestPlan)
	}
	if len(testFiles) > 0 {
		json.Unmarshal(testFiles, &run.TestFiles)
	}
	if len(runResult) > 0 {
		json.Unmarshal(runResult, &run.RunResult)
	}

	return &run, nil
}

// ListRuns menampilkan daftar test run dengan pagination
func (s *Store) ListRuns(ctx context.Context, limit, offset int) ([]*agent.TestRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, state, requirements, fix_attempts, created_at, updated_at, finished_at
		FROM test_runs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*agent.TestRun
	for rows.Next() {
		var run agent.TestRun
		var state string
		var finishedAt *time.Time
		if err := rows.Scan(&run.ID, &state, &run.Requirements, &run.FixAttempts,
			&run.CreatedAt, &run.UpdatedAt, &finishedAt); err != nil {
			return nil, err
		}
		run.State = agent.State(state)
		run.FinishedAt = finishedAt
		runs = append(runs, &run)
	}
	return runs, nil
}
