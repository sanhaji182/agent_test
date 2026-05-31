// Package recordings menyediakan model dan store untuk metadata rekaman eksekusi test.
package recordings

import (
	"sync"
	"time"
)

// Recording menyimpan metadata satu rekaman (screenshot/step capture)
type Recording struct {
	ID            string    `json:"id"`
	RunID         string    `json:"run_id"`
	TestName      string    `json:"test_name"`
	StepName      string    `json:"step_name"`
	ScreenshotURL string    `json:"screenshot_url"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"` // "captured", "failed", "pending"
}

// Store menyimpan recordings di memori
type Store struct {
	mu         sync.RWMutex
	recordings []Recording
	counter    int64
}

func NewStore() *Store {
	return &Store{}
}

// Add menambahkan recording baru
func (s *Store) Add(rec Recording) *Recording {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if rec.ID == "" {
		rec.ID = rec.RunID + "-rec-" + itoa(s.counter)
	}
	s.recordings = append(s.recordings, rec)
	return &rec
}

// ByRun mengembalikan semua recordings untuk sebuah run
func (s *Store) ByRun(runID string) []Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Recording
	for _, r := range s.recordings {
		if r.RunID == runID {
			result = append(result, r)
		}
	}
	return result
}

// All mengembalikan semua recordings
func (s *Store) All() []Recording {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Recording{}, s.recordings...)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return string(buf[i+1:])
}
