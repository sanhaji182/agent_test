package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Notification struct {
	ID         string    `json:"id"`
	RunID      string    `json:"run_id"`
	ScheduleID string    `json:"schedule_id,omitempty"`
	Type       string    `json:"type"` // "failure", "flake", "degradation"
	Message    string    `json:"message"`
	Delivered  bool      `json:"delivered"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store struct {
	mu      sync.RWMutex
	items   []Notification
	counter int64
}

func NewStore() *Store { return &Store{} }

func (s *Store) Add(n Notification) *Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	n.ID = fmt.Sprintf("notif-%d", s.counter)
	n.CreatedAt = time.Now()
	s.items = append(s.items, n)
	return &n
}

func (s *Store) List() []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Notification{}, s.items...)
}

func (s *Store) ByRun(runID string) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Notification
	for _, n := range s.items {
		if n.RunID == runID {
			result = append(result, n)
		}
	}
	return result
}

// DeliverWebhook sends a notification payload to a webhook URL
func DeliverWebhook(webhookURL string, payload map[string]string) error {
	if webhookURL == "" {
		return nil
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}

// TriggerFailure creates a notification and attempts webhook delivery
func (s *Store) TriggerFailure(runID, scheduleID, webhookURL, message string) {
	n := s.Add(Notification{
		RunID:      runID,
		ScheduleID: scheduleID,
		Type:       "failure",
		Message:    message,
	})
	if webhookURL != "" {
		err := DeliverWebhook(webhookURL, map[string]string{
			"type":    "failure",
			"run_id":  runID,
			"message": message,
		})
		if err == nil {
			s.mu.Lock()
			for i := range s.items {
				if s.items[i].ID == n.ID {
					s.items[i].Delivered = true
				}
			}
			s.mu.Unlock()
		}
	}
}
