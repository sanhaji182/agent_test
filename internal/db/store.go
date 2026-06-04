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

// Pool mengembalikan pgxpool.Pool untuk digunakan oleh store lainnya
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// CreateRun menyimpan test run baru ke database
func (s *Store) CreateRun(ctx context.Context, run *agent.TestRun) error {
	featureMap, _ := json.Marshal(run.FeatureMap)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO test_runs (
			id, project_id, state, mode, project_path, requirements,
			test_type, test_case_id, test_list_id, prd, api_docs, auth_type,
			credentials, focus_hints, skip_hints, feature_map, created_at, updated_at
		)
		VALUES ($1, NULL, $2, $3, $4, $5, $6, nullif($7, '')::uuid, nullif($8, '')::uuid, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		run.ID, string(run.State), run.Mode, run.ProjectPath, run.Requirements,
		run.TestType, run.TestCaseID, run.TestListID, run.PRD, run.APIDocs,
		run.AuthType, run.Credentials, run.FocusHints, run.SkipHints,
		featureMap, run.CreatedAt, run.UpdatedAt)
	return err
}

// UpdateRun memperbarui data run yang sudah ada
func (s *Store) UpdateRun(ctx context.Context, run *agent.TestRun) error {
	testPlan, _ := json.Marshal(run.TestPlan)
	testFiles, _ := json.Marshal(run.TestFiles)
	runResult, _ := json.Marshal(run.RunResult)
	featureMap, _ := json.Marshal(run.FeatureMap)

	_, err := s.pool.Exec(ctx, `
		UPDATE test_runs SET
			state = $2, code_analysis = $3, test_plan = $4,
			test_files = $5, run_result = $6, fix_attempts = $7,
			error_msg = $8, updated_at = $9, finished_at = $10,
			video_url = $11, video_status = $12, video_duration = $13,
			video_size = $14, video_failure_marker_at = $15,
			mode = $16, project_path = $17, requirements = $18,
			test_type = $19, test_case_id = nullif($20, '')::uuid,
			test_list_id = nullif($21, '')::uuid, prd = $22, api_docs = $23,
			auth_type = $24, credentials = $25, focus_hints = $26,
			skip_hints = $27, feature_map = $28
		WHERE id = $1`,
		run.ID, string(run.State), run.CodeAnalysis,
		testPlan, testFiles, runResult,
		run.FixAttempts, run.Error, time.Now(), run.FinishedAt,
		run.VideoURL, run.VideoStatus, run.VideoDuration,
		run.VideoSize, run.VideoFailureMarkerAt,
		run.Mode, run.ProjectPath, run.Requirements, run.TestType,
		run.TestCaseID, run.TestListID, run.PRD, run.APIDocs, run.AuthType,
		run.Credentials, run.FocusHints, run.SkipHints, featureMap)
	return err
}

// GetRun mengambil detail test run berdasarkan ID
func (s *Store) GetRun(ctx context.Context, id string) (*agent.TestRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, state, mode, project_path, requirements, test_type,
			COALESCE(test_case_id::text, ''), COALESCE(test_list_id::text, ''),
			prd, api_docs, auth_type, credentials, focus_hints, skip_hints, feature_map,
			code_analysis, test_plan, test_files,
			run_result, fix_attempts, error_msg, created_at, updated_at, finished_at,
			video_url, video_status, video_duration, video_size, video_failure_marker_at
		FROM test_runs WHERE id = $1`, id)

	var run agent.TestRun
	var state string
	var testPlan, testFiles, runResult, featureMap []byte
	var codeAnalysis, errorMsg, videoURL, videoStatus *string
	var mode, projectPath, testType, prd, apiDocs, authType, credentials, focusHints, skipHints *string
	var testCaseID, testListID string
	var videoDuration, videoFailureMarkerAt *float64
	var videoSize *int64
	var finishedAt *time.Time

	err := row.Scan(&run.ID, &state, &mode, &projectPath, &run.Requirements,
		&testType, &testCaseID, &testListID, &prd, &apiDocs, &authType,
		&credentials, &focusHints, &skipHints, &featureMap, &codeAnalysis, &testPlan, &testFiles,
		&runResult, &run.FixAttempts,
		&errorMsg, &run.CreatedAt, &run.UpdatedAt, &finishedAt,
		&videoURL, &videoStatus, &videoDuration, &videoSize, &videoFailureMarkerAt)
	if err != nil {
		return nil, err
	}

	run.State = agent.State(state)
	if mode != nil {
		run.Mode = *mode
	}
	if projectPath != nil {
		run.ProjectPath = *projectPath
	}
	if testType != nil {
		run.TestType = *testType
	}
	run.TestCaseID = testCaseID
	run.TestListID = testListID
	if prd != nil {
		run.PRD = *prd
	}
	if apiDocs != nil {
		run.APIDocs = *apiDocs
	}
	if authType != nil {
		run.AuthType = *authType
	}
	if credentials != nil {
		run.Credentials = *credentials
	}
	if focusHints != nil {
		run.FocusHints = *focusHints
	}
	if skipHints != nil {
		run.SkipHints = *skipHints
	}
	if codeAnalysis != nil {
		run.CodeAnalysis = *codeAnalysis
	}
	if errorMsg != nil {
		run.Error = *errorMsg
	}
	if videoURL != nil {
		run.VideoURL = *videoURL
	}
	if videoStatus != nil {
		run.VideoStatus = *videoStatus
	}
	if videoDuration != nil {
		run.VideoDuration = *videoDuration
	}
	if videoSize != nil {
		run.VideoSize = *videoSize
	}
	if videoFailureMarkerAt != nil {
		run.VideoFailureMarkerAt = *videoFailureMarkerAt
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
	if len(featureMap) > 0 {
		json.Unmarshal(featureMap, &run.FeatureMap)
	}

	return &run, nil
}

// ListRuns menampilkan daftar test run dengan pagination
func (s *Store) ListRuns(ctx context.Context, limit, offset int) ([]*agent.TestRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, state, mode, project_path, requirements, test_type,
			COALESCE(test_case_id::text, ''), COALESCE(test_list_id::text, ''), run_result,
			fix_attempts, video_url, video_status, video_duration,
			created_at, updated_at, finished_at
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
		var mode, projectPath, testType, videoURL, videoStatus *string
		var testCaseID, testListID string
		var runResult []byte
		var videoDuration *float64
		if err := rows.Scan(&run.ID, &state, &mode, &projectPath, &run.Requirements,
			&testType, &testCaseID, &testListID, &runResult, &run.FixAttempts, &videoURL, &videoStatus,
			&videoDuration, &run.CreatedAt, &run.UpdatedAt, &finishedAt); err != nil {
			return nil, err
		}
		run.State = agent.State(state)
		if mode != nil {
			run.Mode = *mode
		}
		if projectPath != nil {
			run.ProjectPath = *projectPath
		}
		if testType != nil {
			run.TestType = *testType
		}
		run.TestCaseID = testCaseID
		run.TestListID = testListID
		if len(runResult) > 0 {
			json.Unmarshal(runResult, &run.RunResult)
		}
		if videoURL != nil {
			run.VideoURL = *videoURL
		}
		if videoStatus != nil {
			run.VideoStatus = *videoStatus
		}
		if videoDuration != nil {
			run.VideoDuration = *videoDuration
		}
		run.FinishedAt = finishedAt
		runs = append(runs, &run)
	}
	return runs, nil
}
