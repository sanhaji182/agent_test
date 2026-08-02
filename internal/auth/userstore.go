package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// User merepresentasikan satu user dashboard (login email+password).
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // tidak pernah di-expose ke JSON
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserStore mengelola user dashboard dengan persistence PostgreSQL.
type UserStore struct {
	mu      sync.RWMutex
	users   map[string]*User // keyed by ID
	byEmail map[string]*User // keyed by lowercase email
	dbPool  *pgxpool.Pool
}

// NewUserStore membuat store user in-memory kosong.
func NewUserStore() *UserStore {
	return &UserStore{
		users:   make(map[string]*User),
		byEmail: make(map[string]*User),
	}
}

// EnableDB mengaktifkan persistence PostgreSQL dan memuat user yang tersimpan.
func (s *UserStore) EnableDB(pool *pgxpool.Pool) {
	s.dbPool = pool
	if err := s.loadDB(); err != nil {
		slog.Warn("userstore: gagal memuat user dari DB", "error", err)
	}
}

// Create membuat user baru dengan password ter-hash bcrypt.
func (s *UserStore) Create(email, password, name string, role Role) (*User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, errors.New("email is required")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}
	if !ValidRoles[role] {
		return nil, errors.New("invalid role: " + string(role))
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	s.mu.Lock()
	if _, exists := s.byEmail[email]; exists {
		s.mu.Unlock()
		return nil, errors.New("email already registered")
	}
	s.users[user.ID] = user
	s.byEmail[email] = user
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.persistDB(user); err != nil {
			slog.Warn("userstore: gagal persist create", "error", err)
		}
	}
	return user, nil
}

// Authenticate memverifikasi email+password. Return user jika valid & aktif.
func (s *UserStore) Authenticate(email, password string) (*User, error) {
	email = normalizeEmail(email)
	s.mu.RLock()
	user, ok := s.byEmail[email]
	s.mu.RUnlock()
	if !ok || !user.IsActive {
		return nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}
	return user, nil
}

// List mengembalikan semua user (password hash tidak ikut, via json:"-").
func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]User, 0, len(s.users))
	for _, u := range s.users {
		result = append(result, *u)
	}
	return result
}

// Get mengembalikan user by ID.
func (s *UserStore) Get(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

// Update memperbarui name/role/active. newPassword kosong = password tidak diubah.
func (s *UserStore) Update(id, name string, role Role, isActive bool, newPassword string) (*User, error) {
	s.mu.Lock()
	user, ok := s.users[id]
	if !ok {
		s.mu.Unlock()
		return nil, errors.New("user not found")
	}
	if name != "" {
		user.Name = name
	}
	if ValidRoles[role] {
		user.Role = role
	}
	user.IsActive = isActive
	if newPassword != "" {
		if len(newPassword) < 6 {
			s.mu.Unlock()
			return nil, errors.New("password must be at least 6 characters")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			s.mu.Unlock()
			return nil, err
		}
		user.PasswordHash = string(hash)
	}
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.persistDB(user); err != nil {
			slog.Warn("userstore: gagal persist update", "error", err)
		}
	}
	cp := *user
	return &cp, nil
}

// Delete menghapus user.
func (s *UserStore) Delete(id string) bool {
	s.mu.Lock()
	user, ok := s.users[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	delete(s.users, id)
	delete(s.byEmail, user.Email)
	s.mu.Unlock()

	if s.dbPool != nil {
		if err := s.deleteDB(id); err != nil {
			slog.Warn("userstore: gagal persist delete", "error", err)
		}
	}
	return true
}

// Count mengembalikan jumlah user.
func (s *UserStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// SeedDefaultAdmin membuat admin default jika belum ada user sama sekali,
// supaya sistem bisa bootstrap login email+password (first-run setup).
func (s *UserStore) SeedDefaultAdmin(email, password, name string) *User {
	if s.Count() > 0 {
		return nil
	}
	if password == "" {
		password = "admin123"
	}
	if name == "" {
		name = "Administrator"
	}
	user, err := s.Create(email, password, name, RoleAdmin)
	if err != nil {
		slog.Warn("userstore: gagal seed admin default", "error", err)
		return nil
	}
	slog.Info("userstore: admin default dibuat (first-run)", "email", email)
	return user
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ─── PostgreSQL persistence ─────────────────────────────────────────────

func (s *UserStore) persistDB(u *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, is_active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			role = EXCLUDED.role,
			is_active = EXCLUDED.is_active`,
		u.ID, u.Email, u.PasswordHash, u.Name, string(u.Role), u.IsActive, u.CreatedAt)
	return err
}

func (s *UserStore) deleteDB(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (s *UserStore) loadDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, email, password_hash, name, role, is_active, created_at FROM users`)
	if err != nil {
		return err
	}
	defer rows.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	for rows.Next() {
		var u User
		var role string
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &role, &u.IsActive, &u.CreatedAt); err != nil {
			return err
		}
		u.Role = Role(role)
		s.users[u.ID] = &u
		s.byEmail[u.Email] = &u
	}
	return rows.Err()
}
