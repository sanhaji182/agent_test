// Package project stores self-hosted project setup, source specs, and extracted feature maps.
package project

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	TestType    string            `json:"test_type"`
	BaseURL     string            `json:"base_url"`
	Environment string            `json:"environment,omitempty"`
	Spec        string            `json:"spec,omitempty"`
	APIDocs     string            `json:"api_docs,omitempty"`
	AuthType    string            `json:"auth_type,omitempty"`
	Credentials string            `json:"credentials,omitempty"`
	FocusHints  string            `json:"focus_hints,omitempty"`
	SkipHints   string            `json:"skip_hints,omitempty"`
	FeatureMap  *agent.FeatureMap `json:"feature_map,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Store interface {
	Create(context.Context, *Project) error
	Update(context.Context, *Project) error
	Get(context.Context, string) (*Project, error)
	List(context.Context, int, int) ([]*Project, error)
}

type MemoryStore struct {
	mu       sync.RWMutex
	projects map[string]*Project
	order    []string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{projects: make(map[string]*Project)}
}

func (s *MemoryStore) Create(_ context.Context, p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepareProject(p)
	s.projects[p.ID] = p
	s.order = append([]string{p.ID}, s.order...)
	return nil
}

func (s *MemoryStore) Update(_ context.Context, p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.ID]; !ok {
		return pgx.ErrNoRows
	}
	p.UpdatedAt = time.Now()
	s.projects[p.ID] = p
	return nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return p, nil
}

func (s *MemoryStore) List(_ context.Context, limit, offset int) ([]*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if offset >= len(s.order) {
		return []*Project{}, nil
	}
	end := offset + limit
	if end > len(s.order) {
		end = len(s.order)
	}
	result := make([]*Project, 0, end-offset)
	for _, id := range s.order[offset:end] {
		if p, ok := s.projects[id]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

type DBStore struct {
	pool *pgxpool.Pool
}

func NewDBStore(pool *pgxpool.Pool) *DBStore {
	return &DBStore{pool: pool}
}

func (s *DBStore) Create(ctx context.Context, p *Project) error {
	prepareProject(p)
	featureMap, _ := json.Marshal(p.FeatureMap)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO projects (
			id, name, path, test_type, base_url, environment, spec,
			api_docs, auth_type, credentials, focus_hints, skip_hints,
			feature_map, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		p.ID, p.Name, p.BaseURL, p.TestType, p.BaseURL, p.Environment, p.Spec,
		p.APIDocs, p.AuthType, p.Credentials, p.FocusHints, p.SkipHints,
		featureMap, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *DBStore) Update(ctx context.Context, p *Project) error {
	p.UpdatedAt = time.Now()
	featureMap, _ := json.Marshal(p.FeatureMap)
	_, err := s.pool.Exec(ctx, `
		UPDATE projects SET
			name = $2, path = $3, test_type = $4, base_url = $5,
			environment = $6, spec = $7, api_docs = $8, auth_type = $9,
			credentials = $10, focus_hints = $11, skip_hints = $12,
			feature_map = $13, updated_at = $14
		WHERE id = $1`,
		p.ID, p.Name, p.BaseURL, p.TestType, p.BaseURL, p.Environment,
		p.Spec, p.APIDocs, p.AuthType, p.Credentials, p.FocusHints,
		p.SkipHints, featureMap, p.UpdatedAt)
	return err
}

func (s *DBStore) Get(ctx context.Context, id string) (*Project, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, COALESCE(test_type, ''), COALESCE(base_url, path, ''),
			COALESCE(environment, ''), COALESCE(spec, ''), COALESCE(api_docs, ''),
			COALESCE(auth_type, ''), COALESCE(credentials, ''), COALESCE(focus_hints, ''),
			COALESCE(skip_hints, ''), feature_map, created_at, updated_at
		FROM projects WHERE id = $1`, id)
	return scanProject(row)
}

func (s *DBStore) List(ctx context.Context, limit, offset int) ([]*Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, COALESCE(test_type, ''), COALESCE(base_url, path, ''),
			COALESCE(environment, ''), COALESCE(spec, ''), COALESCE(api_docs, ''),
			COALESCE(auth_type, ''), COALESCE(credentials, ''), COALESCE(focus_hints, ''),
			COALESCE(skip_hints, ''), feature_map, created_at, updated_at
		FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		result = []*Project{}
	}
	return result, rows.Err()
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanProject(row scanner) (*Project, error) {
	var p Project
	var featureMap []byte
	if err := row.Scan(&p.ID, &p.Name, &p.TestType, &p.BaseURL, &p.Environment,
		&p.Spec, &p.APIDocs, &p.AuthType, &p.Credentials, &p.FocusHints,
		&p.SkipHints, &featureMap, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	if len(featureMap) > 0 {
		_ = json.Unmarshal(featureMap, &p.FeatureMap)
	}
	return &p, nil
}

func prepareProject(p *Project) {
	now := time.Now()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.TestType == "" {
		p.TestType = "ui"
	}
	if p.Environment == "" {
		p.Environment = "default"
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
}
