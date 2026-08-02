// Package llmprofile menyediakan penyimpanan profil konfigurasi LLM
// (multi-provider). User dapat menyimpan beberapa profil provider dan
// mengaktifkan salah satunya untuk dipakai saat run test.
package llmprofile

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errNameRequired dikembalikan saat membuat profil tanpa nama.
var errNameRequired = errors.New("name is required")

// Profile menyimpan satu konfigurasi LLM provider.
type Profile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Model       string    `json:"model"`
	Temperature string    `json:"temperature"`
	MaxTokens   string    `json:"max_tokens"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store mengelola profil LLM (in-memory + opsional persistence PostgreSQL).
type Store struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
	dbPool   *pgxpool.Pool
}

// NewStore membuat store in-memory kosong.
func NewStore() *Store {
	return &Store{profiles: make(map[string]*Profile)}
}

// EnableDB mengaktifkan persistence PostgreSQL dan memuat profil yang tersimpan.
func (s *Store) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
	profiles, err := s.listDB()
	if err != nil {
		slog.Warn("llmprofile: gagal memuat profil dari DB", "error", err)
		return
	}
	s.mu.Lock()
	for i := range profiles {
		s.profiles[profiles[i].ID] = &profiles[i]
	}
	s.mu.Unlock()
}

// maskAPIKey menyamarkan API key untuk tampilan (first4...last4).
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// masked mengembalikan copy profile dengan API key tersamar.
func masked(p *Profile) Profile {
	safe := *p
	safe.APIKey = maskAPIKey(p.APIKey)
	return safe
}

// Create menyimpan profil baru. Jika IsActive, profil lain dinonaktifkan.
func (s *Store) Create(p *Profile) (*Profile, error) {
	if p.Name == "" {
		return nil, errNameRequired
	}
	now := time.Now()
	p.ID = uuid.New().String()
	p.CreatedAt = now
	p.UpdatedAt = now

	s.mu.Lock()
	if p.IsActive {
		s.deactivateAllLocked()
	}
	s.profiles[p.ID] = p
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.persistDB(p); err != nil {
			slog.Warn("llmprofile: gagal persist create", "error", err)
		}
		if p.IsActive {
			_ = s.setActiveDB(p.ID)
		}
	}
	out := masked(p)
	return &out, nil
}

// List mengembalikan semua profil (API key tersamar), urut created_at desc.
func (s *Store) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Profile, 0, len(s.profiles))
	for _, p := range s.profiles {
		result = append(result, masked(p))
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].CreatedAt.After(result[i].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// Get mengembalikan profil by ID (API key tersamar).
func (s *Store) Get(id string) (*Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, false
	}
	out := masked(p)
	return &out, true
}

// GetRaw mengembalikan profil by ID dengan API key ASLI (internal use,
// misalnya untuk test koneksi; jangan encode langsung ke response API).
func (s *Store) GetRaw(id string) (*Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

// Update memperbarui profil yang ada. APIKey kosong berarti tidak diubah.
func (s *Store) Update(id string, in *Profile) (*Profile, bool) {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return nil, false
	}
	if in.Name != "" {
		p.Name = in.Name
	}
	p.Provider = in.Provider
	p.BaseURL = in.BaseURL
	p.Model = in.Model
	p.Temperature = in.Temperature
	p.MaxTokens = in.MaxTokens
	if in.APIKey != "" {
		p.APIKey = in.APIKey
	}
	becameActive := in.IsActive && !p.IsActive
	if in.IsActive {
		s.deactivateAllLocked()
		p.IsActive = true
	} else {
		p.IsActive = false
	}
	p.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.persistDB(p); err != nil {
			slog.Warn("llmprofile: gagal persist update", "error", err)
		}
		if becameActive {
			_ = s.setActiveDB(p.ID)
		}
	}
	out := masked(p)
	return &out, true
}

// Delete menghapus profil.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	_, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	delete(s.profiles, id)
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.deleteDB(id); err != nil {
			slog.Warn("llmprofile: gagal persist delete", "error", err)
		}
	}
	return true
}

// SetActive mengaktifkan profil id dan menonaktifkan yang lain.
func (s *Store) SetActive(id string) bool {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	s.deactivateAllLocked()
	p.IsActive = true
	p.UpdatedAt = time.Now()
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.setActiveDB(id); err != nil {
			slog.Warn("llmprofile: gagal persist set-active", "error", err)
		}
	}
	return true
}

// ActiveRaw mengembalikan profil aktif dengan API key ASLI (untuk internal
// buildAgentForRun; jangan encode langsung ke response API).
func (s *Store) ActiveRaw() (*Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.profiles {
		if p.IsActive {
			cp := *p
			return &cp, true
		}
	}
	return nil, false
}

// deactivateAllLocked menonaktifkan semua profil (caller harus hold lock).
func (s *Store) deactivateAllLocked() {
	for _, p := range s.profiles {
		p.IsActive = false
	}
}

// ─── PostgreSQL persistence ─────────────────────────────────────────────

func llmProfileDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *Store) persistDB(p *Profile) error {
	ctx, cancel := llmProfileDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO llm_profiles (id, name, provider, base_url, api_key, model, temperature, max_tokens, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			provider = EXCLUDED.provider,
			base_url = EXCLUDED.base_url,
			api_key = EXCLUDED.api_key,
			model = EXCLUDED.model,
			temperature = EXCLUDED.temperature,
			max_tokens = EXCLUDED.max_tokens,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at`,
		p.ID, p.Name, p.Provider, p.BaseURL, p.APIKey, p.Model, p.Temperature, p.MaxTokens, p.IsActive, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *Store) setActiveDB(id string) error {
	ctx, cancel := llmProfileDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `UPDATE llm_profiles SET is_active = (id = $1)`, id)
	return err
}

func (s *Store) deleteDB(id string) error {
	ctx, cancel := llmProfileDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `DELETE FROM llm_profiles WHERE id = $1`, id)
	return err
}

func (s *Store) listDB() ([]Profile, error) {
	ctx, cancel := llmProfileDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, name, provider, base_url, api_key, model, temperature, max_tokens, is_active, created_at, updated_at
		FROM llm_profiles ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.Provider, &p.BaseURL, &p.APIKey, &p.Model, &p.Temperature, &p.MaxTokens, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return profiles, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}
