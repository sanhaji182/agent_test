package webhook

import (
	"context"
	"time"
)

func webhookDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *RegistrationStore) persistRegistrationDB(reg *WebhookRegistration) error {
	ctx, cancel := webhookDBContext()
	defer cancel()
	userID := reg.UserID
	if userID == "" {
		userID = "default"
	}
	webhookID := reg.WebhookID
	if webhookID == "" {
		webhookID = reg.ID
	}
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO webhook_registrations (id, user_id, repository_id, repository_url, webhook_id, status, last_sync_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			repository_id = EXCLUDED.repository_id,
			repository_url = EXCLUDED.repository_url,
			webhook_id = EXCLUDED.webhook_id,
			status = EXCLUDED.status,
			last_sync_at = EXCLUDED.last_sync_at,
			updated_at = EXCLUDED.updated_at`,
		reg.ID, userID, reg.RepositoryID, reg.RepositoryURL, webhookID, reg.Status, reg.LastSyncAt, reg.CreatedAt, reg.UpdatedAt)
	return err
}

func (s *RegistrationStore) listRegistrationsDB() ([]WebhookRegistration, error) {
	ctx, cancel := webhookDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, user_id, repository_id, repository_url, webhook_id, status, last_sync_at, created_at, updated_at
		FROM webhook_registrations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var regs []WebhookRegistration
	for rows.Next() {
		reg, err := scanRegistration(rows)
		if err != nil {
			return regs, err
		}
		regs = append(regs, *reg)
	}
	return regs, rows.Err()
}

func (s *RegistrationStore) getRegistrationDB(id string) (*WebhookRegistration, error) {
	ctx, cancel := webhookDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, user_id, repository_id, repository_url, webhook_id, status, last_sync_at, created_at, updated_at
		FROM webhook_registrations WHERE id = $1`, id)
	return scanRegistration(row)
}

func scanRegistration(row interface{ Scan(dest ...any) error }) (*WebhookRegistration, error) {
	var reg WebhookRegistration
	if err := row.Scan(&reg.ID, &reg.UserID, &reg.RepositoryID, &reg.RepositoryURL, &reg.WebhookID, &reg.Status, &reg.LastSyncAt, &reg.CreatedAt, &reg.UpdatedAt); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (s *RegistrationStore) deleteRegistrationDB(id string) error {
	ctx, cancel := webhookDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `DELETE FROM webhook_registrations WHERE id = $1`, id)
	return err
}
