package schedule

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type Frequency string

const (
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
	Cron    Frequency = "cron"
)

type Schedule struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	TestListID    string     `json:"test_list_id,omitempty"`
	Name          string     `json:"name"`
	ProjectPath   string     `json:"project_path"`
	Requirements  string     `json:"requirements"`
	Mode          string     `json:"mode"`
	Environment   string     `json:"environment"` // local, staging, production
	BaseURL       string     `json:"base_url"`    // Target URL for tests
	Frequency     Frequency  `json:"frequency"`
	CronExpr      string     `json:"cron_expr,omitempty"`
	Timezone      string     `json:"timezone"`
	Enabled       bool       `json:"enabled"`
	NextRunAt     time.Time  `json:"next_run_at"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	LastRunID     string     `json:"last_run_id,omitempty"`
	LastRunStatus string     `json:"last_run_status,omitempty"`
	NotifyOnFail  bool       `json:"notify_on_fail"`
	WebhookURL    string     `json:"webhook_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Repository interface {
	Create(*Schedule) *Schedule
	Get(string) (*Schedule, bool)
	List() []*Schedule
	Update(string, func(*Schedule)) bool
	Delete(string) bool
	GetDue(time.Time) []*Schedule
	// ClaimNextDue atomically claims the next due schedule, preventing
	// duplicate processing across concurrent workers (ADR-004, implemented 2026-07-27).
	// Returns the claimed schedule or nil if nothing is due.
	ClaimNextDue(now time.Time, claimID string) *Schedule
}

type Store struct {
	mu        sync.RWMutex
	schedules map[string]*Schedule
	order     []string
}

func NewStore() *Store {
	return &Store{schedules: make(map[string]*Schedule)}
}

func (s *Store) Create(sch *Schedule) *Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepareSchedule(sch)
	s.schedules[sch.ID] = sch
	s.order = append([]string{sch.ID}, s.order...)
	return sch
}

func (s *Store) Get(id string) (*Schedule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sch, ok := s.schedules[id]
	return sch, ok
}

func (s *Store) List() []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Schedule
	for _, id := range s.order {
		if sch, ok := s.schedules[id]; ok {
			result = append(result, sch)
		}
	}
	return result
}

func (s *Store) Update(id string, fn func(*Schedule)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sch, ok := s.schedules[id]
	if !ok {
		return false
	}
	fn(sch)
	sch.UpdatedAt = time.Now()
	return true
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return false
	}
	delete(s.schedules, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// GetDue returns schedules that are enabled and past their next_run_at
func (s *Store) GetDue(now time.Time) []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var due []*Schedule
	for _, sch := range s.schedules {
		if sch.Enabled && !sch.NextRunAt.After(now) {
			due = append(due, sch)
		}
	}
	return due
}

// ClaimNextDue atomically claims the next due schedule. For the memory store,
// this is a locked select-and-advance operation that prevents double-claiming
// within a single process (but not across processes — see DBStore for that).
func (s *DBStore) ClaimNextDue(now time.Time, _ string) *Schedule {
	advanceTo := now.Add(1 * time.Minute)
	row := s.pool.QueryRow(context.Background(), `
		WITH next AS (
			SELECT id FROM schedules
			WHERE enabled = true AND next_run_at <= $1
			ORDER BY next_run_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE schedules SET next_run_at = $2, updated_at = NOW()
		FROM next
		WHERE schedules.id = next.id
		RETURNING
			schedules.id, COALESCE(schedules.project_id::text, ''),
			COALESCE(schedules.test_list_id::text, ''), schedules.name,
			schedules.project_path, schedules.requirements, schedules.mode,
			schedules.environment, schedules.base_url, schedules.frequency,
			COALESCE(schedules.cron_expr, ''), schedules.timezone, schedules.enabled,
			schedules.next_run_at, schedules.last_run_at,
			COALESCE(schedules.last_run_id::text, ''),
			COALESCE(schedules.last_run_status, ''),
			schedules.notify_on_fail, COALESCE(schedules.webhook_url, ''),
			schedules.created_at, schedules.updated_at`,
		now, advanceTo)
	sch, err := scanSchedule(row)
	if err != nil {
		return nil
	}
	return sch
}

func (s *Store) ClaimNextDue(now time.Time, _ string) *Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()
	var best *Schedule
	for _, sch := range s.schedules {
		if sch.Enabled && !sch.NextRunAt.After(now) {
			if best == nil || sch.NextRunAt.Before(best.NextRunAt) {
				best = sch
			}
		}
	}
	if best != nil {
		best.NextRunAt = CalcNextRun(best.Frequency, best.CronExpr, now)
	}
	return best
}

type DBStore struct {
	pool *pgxpool.Pool
}

func NewDBStore(pool *pgxpool.Pool) *DBStore {
	return &DBStore{pool: pool}
}

func (s *DBStore) Create(sch *Schedule) *Schedule {
	prepareSchedule(sch)
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO schedules (
			id, project_id, test_list_id, name, project_path, requirements, mode,
			environment, base_url, frequency, cron_expr, timezone, enabled,
			next_run_at, last_run_at, last_run_id, last_run_status,
			notify_on_fail, webhook_url, created_at, updated_at
		)
		VALUES (
			$1, nullif($2, '')::uuid, nullif($3, '')::uuid, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14, $15,
			nullif($16, '')::uuid, $17, $18, $19, $20, $21
		)`,
		sch.ID, sch.ProjectID, sch.TestListID, sch.Name, sch.ProjectPath, sch.Requirements, sch.Mode,
		sch.Environment, sch.BaseURL, sch.Frequency, sch.CronExpr, sch.Timezone, sch.Enabled,
		sch.NextRunAt, sch.LastRunAt, sch.LastRunID, sch.LastRunStatus,
		sch.NotifyOnFail, sch.WebhookURL, sch.CreatedAt, sch.UpdatedAt)
	if err != nil {
		return nil
	}
	return sch
}

func (s *DBStore) Get(id string) (*Schedule, bool) {
	row := s.pool.QueryRow(context.Background(), `
		SELECT id, COALESCE(project_id::text, ''), COALESCE(test_list_id::text, ''),
			name, project_path, requirements, mode, environment, base_url, frequency,
			COALESCE(cron_expr, ''), timezone, enabled, next_run_at, last_run_at,
			COALESCE(last_run_id::text, ''), COALESCE(last_run_status, ''),
			notify_on_fail, COALESCE(webhook_url, ''), created_at, updated_at
		FROM schedules WHERE id = $1`, id)
	sch, err := scanSchedule(row)
	return sch, err == nil
}

func (s *DBStore) List() []*Schedule {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, COALESCE(project_id::text, ''), COALESCE(test_list_id::text, ''),
			name, project_path, requirements, mode, environment, base_url, frequency,
			COALESCE(cron_expr, ''), timezone, enabled, next_run_at, last_run_at,
			COALESCE(last_run_id::text, ''), COALESCE(last_run_status, ''),
			notify_on_fail, COALESCE(webhook_url, ''), created_at, updated_at
		FROM schedules ORDER BY created_at DESC`)
	if err != nil {
		return []*Schedule{}
	}
	defer rows.Close()
	var result []*Schedule
	for rows.Next() {
		sch, err := scanSchedule(rows)
		if err == nil {
			result = append(result, sch)
		}
	}
	if result == nil {
		result = []*Schedule{}
	}
	return result
}

func (s *DBStore) Update(id string, fn func(*Schedule)) bool {
	sch, ok := s.Get(id)
	if !ok {
		return false
	}
	fn(sch)
	sch.UpdatedAt = time.Now()
	_, err := s.pool.Exec(context.Background(), `
		UPDATE schedules SET
			project_id = nullif($2, '')::uuid,
			test_list_id = nullif($3, '')::uuid,
			name = $4,
			project_path = $5,
			requirements = $6,
			mode = $7,
			environment = $8,
			base_url = $9,
			frequency = $10,
			cron_expr = $11,
			timezone = $12,
			enabled = $13,
			next_run_at = $14,
			last_run_at = $15,
			last_run_id = nullif($16, '')::uuid,
			last_run_status = $17,
			notify_on_fail = $18,
			webhook_url = $19,
			updated_at = $20
		WHERE id = $1`,
		sch.ID, sch.ProjectID, sch.TestListID, sch.Name, sch.ProjectPath,
		sch.Requirements, sch.Mode, sch.Environment, sch.BaseURL, sch.Frequency,
		sch.CronExpr, sch.Timezone, sch.Enabled, sch.NextRunAt, sch.LastRunAt,
		sch.LastRunID, sch.LastRunStatus, sch.NotifyOnFail, sch.WebhookURL, sch.UpdatedAt)
	return err == nil
}

func (s *DBStore) Delete(id string) bool {
	tag, err := s.pool.Exec(context.Background(), `DELETE FROM schedules WHERE id = $1`, id)
	return err == nil && tag.RowsAffected() > 0
}

func (s *DBStore) GetDue(now time.Time) []*Schedule {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, COALESCE(project_id::text, ''), COALESCE(test_list_id::text, ''),
			name, project_path, requirements, mode, environment, base_url, frequency,
			COALESCE(cron_expr, ''), timezone, enabled, next_run_at, last_run_at,
			COALESCE(last_run_id::text, ''), COALESCE(last_run_status, ''),
			notify_on_fail, COALESCE(webhook_url, ''), created_at, updated_at
		FROM schedules
		WHERE enabled = true AND next_run_at <= $1
		ORDER BY next_run_at ASC`, now)
	if err != nil {
		return []*Schedule{}
	}
	defer rows.Close()
	var result []*Schedule
	for rows.Next() {
		sch, err := scanSchedule(rows)
		if err == nil {
			result = append(result, sch)
		}
	}
	if result == nil {
		result = []*Schedule{}
	}
	return result
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanSchedule(row scanner) (*Schedule, error) {
	var sch Schedule
	var lastRunAt *time.Time
	var frequency string
	err := row.Scan(&sch.ID, &sch.ProjectID, &sch.TestListID, &sch.Name,
		&sch.ProjectPath, &sch.Requirements, &sch.Mode, &sch.Environment,
		&sch.BaseURL, &frequency, &sch.CronExpr, &sch.Timezone, &sch.Enabled,
		&sch.NextRunAt, &lastRunAt, &sch.LastRunID, &sch.LastRunStatus,
		&sch.NotifyOnFail, &sch.WebhookURL, &sch.CreatedAt, &sch.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	sch.Frequency = Frequency(frequency)
	sch.LastRunAt = lastRunAt
	return &sch, nil
}

func prepareSchedule(sch *Schedule) {
	now := time.Now()
	if sch.ID == "" {
		sch.ID = uuid.New().String()
	}
	if sch.Frequency == "" {
		sch.Frequency = Daily
	}
	if sch.Mode == "" {
		sch.Mode = "simple"
	}
	if sch.Timezone == "" {
		sch.Timezone = "UTC"
	}
	if sch.CreatedAt.IsZero() {
		sch.CreatedAt = now
	}
	if sch.UpdatedAt.IsZero() {
		sch.UpdatedAt = now
	}
	if sch.NextRunAt.IsZero() {
		sch.NextRunAt = CalcNextRun(sch.Frequency, sch.CronExpr, now)
	}
}

// CalcNextRun calculates the next run time based on frequency or cron expression
func CalcNextRun(freq Frequency, cronExpr string, from time.Time) time.Time {
	if freq == Cron && cronExpr != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sched, err := parser.Parse(cronExpr)
		if err == nil {
			return sched.Next(from)
		}
	}
	switch freq {
	case Daily:
		return from.Add(24 * time.Hour)
	case Weekly:
		return from.Add(7 * 24 * time.Hour)
	case Monthly:
		return from.AddDate(0, 1, 0)
	default:
		return from.Add(24 * time.Hour)
	}
}
