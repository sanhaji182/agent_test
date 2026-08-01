package release

import (
	"context"
	"encoding/json"
	"time"
)

func releaseDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *Store) persistReleaseDB(rel *Release) error {
	ctx, cancel := releaseDBContext()
	defer cancel()
	runIDs, _ := json.Marshal(rel.RunIDs)
	runIDsJSON := string(runIDs)
	if runIDsJSON == "" || runIDsJSON == "null" {
		runIDsJSON = "[]"
	}
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO releases (id, name, version, project_id, status, run_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			version = EXCLUDED.version,
			project_id = EXCLUDED.project_id,
			status = EXCLUDED.status,
			run_ids = EXCLUDED.run_ids,
			updated_at = EXCLUDED.updated_at`,
		rel.ID, rel.Name, rel.Version, rel.ProjectID, rel.Status, runIDsJSON, rel.CreatedAt, rel.UpdatedAt)
	return err
}

func (s *Store) getReleaseDB(id string) (*Release, error) {
	ctx, cancel := releaseDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, name, version, project_id, status, run_ids, created_at, updated_at
		FROM releases WHERE id = $1`, id)
	return scanRelease(row)
}

func (s *Store) listReleasesDB() ([]*Release, error) {
	ctx, cancel := releaseDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, name, version, project_id, status, run_ids, created_at, updated_at
		FROM releases ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Release
	for rows.Next() {
		rel, err := scanRelease(rows)
		if err != nil {
			return result, err
		}
		result = append(result, rel)
	}
	return result, rows.Err()
}

func scanRelease(row interface{ Scan(dest ...any) error }) (*Release, error) {
	var rel Release
	var runIDs []byte
	if err := row.Scan(&rel.ID, &rel.Name, &rel.Version, &rel.ProjectID, &rel.Status, &runIDs, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
		return nil, err
	}
	if len(runIDs) > 0 {
		_ = json.Unmarshal(runIDs, &rel.RunIDs)
	}
	return &rel, nil
}
