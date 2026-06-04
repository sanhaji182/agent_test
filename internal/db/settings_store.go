package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingsStore handles persistent key-value settings
type SettingsStore struct {
	pool *pgxpool.Pool
}

func NewSettingsStore(pool *pgxpool.Pool) *SettingsStore {
	return &SettingsStore{pool: pool}
}

// GetAll returns all settings as a map
func (s *SettingsStore) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := s.pool.Query(ctx, "SELECT key, value FROM settings ORDER BY key")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, nil
}

// Get returns a single setting value
func (s *SettingsStore) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, "SELECT value FROM settings WHERE key = $1", key).Scan(&value)
	return value, err
}

// Set upserts a single setting
func (s *SettingsStore) Set(ctx context.Context, key, value string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3`,
		key, value, time.Now())
	return err
}

// SetMany upserts multiple settings at once
func (s *SettingsStore) SetMany(ctx context.Context, settings map[string]string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for k, v := range settings {
		_, err := tx.Exec(ctx, `
			INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, $3)
			ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3`,
			k, v, time.Now())
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
