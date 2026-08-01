package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// APIKeyEntry represents a single API key with its associated role and metadata.
type APIKeyEntry struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`    // plain key (only returned on creation)
	KeyHash   string    `json:"-"`      // SHA-256 hash for storage
	Label     string    `json:"label"`  // human-readable name (e.g., "John's Key", "CI Pipeline")
	Role      Role      `json:"role"`   // admin, reviewer, viewer, api_client
	Active    bool      `json:"active"` // can be disabled without deleting
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by,omitempty"`
}

// KeyStore manages multiple API keys with role assignments.
type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]*APIKeyEntry // keyed by key hash
	byID map[string]*APIKeyEntry // keyed by entry ID
}

// NewKeyStore creates a new in-memory key store.
func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]*APIKeyEntry),
		byID: make(map[string]*APIKeyEntry),
	}
}

// Validate checks whether an API key is valid and returns the associated role and label.
// Returns the role and label if valid, or empty strings if not found/inactive.
func (ks *KeyStore) Validate(key string) (Role, string, string, bool) {
	hash := hashKey(key)

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	entry, ok := ks.keys[hash]
	if !ok || !entry.Active {
		return "", "", "", false
	}
	return entry.Role, entry.Label, entry.ID, true
}

// Create generates a new API key with the given label and role.
// Returns the full entry including the plain key (only time it's visible).
func (ks *KeyStore) Create(label string, role Role, createdBy string) (*APIKeyEntry, error) {
	if label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if !ValidRoles[role] {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	plain, _, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	hash := hashKey(plain)

	entry := &APIKeyEntry{
		ID:        uuid.New().String(),
		Key:       plain,
		KeyHash:   hash,
		Label:     label,
		Role:      role,
		Active:    true,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	ks.mu.Lock()
	ks.keys[hash] = entry
	ks.byID[entry.ID] = entry
	ks.mu.Unlock()

	return entry, nil
}

// List returns all API key entries (without plain keys).
func (ks *KeyStore) List() []APIKeyEntry {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	result := make([]APIKeyEntry, 0, len(ks.byID))
	for _, entry := range ks.byID {
		// Don't expose the plain key
		safe := *entry
		safe.Key = ""
		result = append(result, safe)
	}
	return result
}

// Get returns a specific entry by ID (without plain key).
func (ks *KeyStore) Get(id string) (*APIKeyEntry, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	entry, ok := ks.byID[id]
	if !ok {
		return nil, false
	}
	safe := *entry
	safe.Key = ""
	return &safe, true
}

// Revoke disables (or enables) a key by ID.
func (ks *KeyStore) Revoke(id string, active bool) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	entry, ok := ks.byID[id]
	if !ok {
		return false
	}
	entry.Active = active
	return true
}

// Delete removes a key permanently.
func (ks *KeyStore) Delete(id string) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	entry, ok := ks.byID[id]
	if !ok {
		return false
	}
	delete(ks.keys, entry.KeyHash)
	delete(ks.byID, id)
	return true
}

// Count returns the number of active keys.
func (ks *KeyStore) Count() int {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return len(ks.byID)
}

// SeedDefaultKey creates a default admin key if no keys exist.
// This ensures backward compatibility — the first startup creates an admin key.
func (ks *KeyStore) SeedDefaultKey() *APIKeyEntry {
	if ks.Count() > 0 {
		return nil
	}
	entry, err := ks.Create("Default Admin Key", RoleAdmin, "system")
	if err != nil {
		return nil
	}
	return entry
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
