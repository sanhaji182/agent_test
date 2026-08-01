package drift

import (
	"context"
	"time"
)

func driftDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *Store) persistDriftDB(d *Drift) error {
	ctx, cancel := driftDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO drifts (id, repository_id, type, file_path, description, severity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			repository_id = EXCLUDED.repository_id,
			type = EXCLUDED.type,
			file_path = EXCLUDED.file_path,
			description = EXCLUDED.description,
			severity = EXCLUDED.severity,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		d.ID, d.Repository, d.Type, d.FilePath, d.Description, d.Severity, d.Status, d.CreatedAt, d.UpdatedAt)
	return err
}

func (s *Store) listDriftsDB(repository, driftType, status string) ([]Drift, error) {
	ctx, cancel := driftDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, repository_id, type, file_path, description, severity, status, created_at, updated_at
		FROM drifts
		WHERE ($1 = '' OR repository_id = $1)
		  AND ($2 = '' OR type = $2)
		  AND ($3 = '' OR status = $3)
		ORDER BY created_at DESC`, repository, driftType, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var drifts []Drift
	for rows.Next() {
		d, err := scanDrift(rows)
		if err != nil {
			return drifts, err
		}
		drifts = append(drifts, *d)
	}
	return drifts, rows.Err()
}

func (s *Store) hasPendingDB(repository, driftType, filePath string) (bool, error) {
	ctx, cancel := driftDBContext()
	defer cancel()
	var exists bool
	err := s.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM drifts
			WHERE status = $1 AND repository_id = $2 AND type = $3 AND file_path = $4
		)`, StatusPending, repository, driftType, filePath).Scan(&exists)
	return exists, err
}

func (s *Store) getDriftDB(id string) (*Drift, error) {
	ctx, cancel := driftDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, repository_id, type, file_path, description, severity, status, created_at, updated_at
		FROM drifts WHERE id = $1`, id)
	return scanDrift(row)
}

func scanDrift(row interface{ Scan(dest ...any) error }) (*Drift, error) {
	var d Drift
	if err := row.Scan(&d.ID, &d.Repository, &d.Type, &d.FilePath, &d.Description, &d.Severity, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
