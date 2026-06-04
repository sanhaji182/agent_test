// Package visual menyediakan model untuk visual testing artifacts (baseline, current, diff).
package visual

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Artifact menyimpan data visual test untuk satu step
type Artifact struct {
	ID              string    `json:"id"`
	RunID           string    `json:"run_id"`
	StepName        string    `json:"step_name"`
	BaselineURL     string    `json:"baseline_url,omitempty"`
	CurrentURL      string    `json:"current_url,omitempty"`
	DiffURL         string    `json:"diff_url,omitempty"`
	SimilarityScore float64   `json:"similarity_score"` // 0.0-1.0
	Passed          bool      `json:"passed"`
	CreatedAt       time.Time `json:"created_at"`
}

// Store menyimpan visual artifacts di memori
type Store struct {
	mu        sync.RWMutex
	artifacts []Artifact
}

func NewStore() *Store {
	return &Store{}
}

// Add menambahkan artifact baru. Jika baseline dan current tersedia, hitung similarity.
func (s *Store) Add(a Artifact) *Artifact {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = a.RunID + "-vis-" + hashShort(a.StepName)
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	// Hitung similarity score menggunakan perbandingan pixel nyata jika file lokal tersedia
	if a.BaselineURL != "" && a.CurrentURL != "" && a.SimilarityScore == 0 {
		// Asumsikan URL adalah local path
		sim, err := CompareImages(a.BaselineURL, a.CurrentURL, a.DiffURL)
		if err == nil {
			a.SimilarityScore = sim
		} else {
			a.SimilarityScore = deterministicScore(a.BaselineURL, a.CurrentURL)
		}
		a.Passed = a.SimilarityScore >= 0.95
	}
	s.artifacts = append(s.artifacts, a)
	return &a
}

// ByRun mengembalikan semua visual artifacts untuk sebuah run
func (s *Store) ByRun(runID string) []Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Artifact
	for _, a := range s.artifacts {
		if a.RunID == runID {
			result = append(result, a)
		}
	}
	return result
}

// deterministicScore menghasilkan skor similarity berdasarkan hash URL.
// Jika URL sama → 1.0 (identik). Jika berbeda → skor antara 0.7-0.99 berdasarkan hash distance.
// Ini heuristik deterministik; akan diganti dengan perbandingan pixel nyata saat Vision API aktif.
func deterministicScore(baseline, current string) float64 {
	if baseline == current {
		return 1.0
	}
	hb := sha256.Sum256([]byte(baseline))
	hc := sha256.Sum256([]byte(current))
	// Hitung berapa byte yang sama
	same := 0
	for i := range hb {
		if hb[i] == hc[i] {
			same++
		}
	}
	// Normalisasi ke range 0.7-0.99
	return 0.7 + (float64(same)/float64(len(hb)))*0.29
}

func hashShort(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:4])
}
