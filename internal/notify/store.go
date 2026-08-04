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
	ID           string    `json:"id"`
	RunID        string    `json:"run_id"`
	ScheduleID   string    `json:"schedule_id,omitempty"`
	Type         string    `json:"type"` // "failure", "flake", "degradation"
	Message      string    `json:"message"`
	Delivered    bool      `json:"delivered"`
	Acknowledged bool      `json:"acknowledged,omitempty"` // alert sudah dibaca/di-ack oleh user
	Dismissed    bool      `json:"dismissed,omitempty"`    // alert dibuang dari daftar aktif
	CreatedAt    time.Time `json:"created_at"`
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

// find locates an item by ID and applies fn to it. Returns false when missing.
func (s *Store) find(id string, fn func(*Notification)) bool {
	for i := range s.items {
		if s.items[i].ID == id {
			fn(&s.items[i])
			return true
		}
	}
	return false
}

// Acknowledge menandai alert sebagai sudah dibaca/di-ack oleh user.
func (s *Store) Acknowledge(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.find(id, func(n *Notification) { n.Acknowledged = true })
}

// Dismiss menyembunyikan alert dari daftar aktif.
func (s *Store) Dismiss(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.find(id, func(n *Notification) { n.Dismissed = true })
}

// MarkAllRead menandai semua alert sebagai acknowledged. Mengembalikan jumlah
// alert yang baru saja ditandai (belum acknowledged sebelumnya).
func (s *Store) MarkAllRead() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for i := range s.items {
		if !s.items[i].Acknowledged {
			s.items[i].Acknowledged = true
			count++
		}
	}
	return count
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

// DeliverSlack sends a rich Slack-formatted notification via incoming webhook.
func DeliverSlack(webhookURL, runID, status, message string, passed, failed, total int) error {
	if webhookURL == "" {
		return nil
	}
	emoji := "✅"
	color := "#36a64f"
	if failed > 0 {
		emoji = "❌"
		color = "#dc3545"
	}
	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": fmt.Sprintf("%s Test Run %s", emoji, status),
				"text":  message,
				"fields": []map[string]interface{}{
					{"title": "Run ID", "value": runID, "short": true},
					{"title": "Results", "value": fmt.Sprintf("%d passed / %d failed / %d total", passed, failed, total), "short": true},
				},
				"footer": "GoTest Agent",
				"ts":     time.Now().Unix(),
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

// DeliverTeams sends a Microsoft Teams Adaptive Card notification via incoming webhook.
func DeliverTeams(webhookURL, runID, status, message string, passed, failed, total int) error {
	if webhookURL == "" {
		return nil
	}
	themeColor := "36a64f"
	if failed > 0 {
		themeColor = "dc3545"
	}
	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    fmt.Sprintf("Test Run %s: %d/%d passed", status, passed, total),
		"sections": []map[string]interface{}{
			{
				"activityTitle": fmt.Sprintf("Test Run %s", status),
				"activityText":  message,
				"facts": []map[string]string{
					{"name": "Run ID", "value": runID},
					{"name": "Passed", "value": fmt.Sprintf("%d", passed)},
					{"name": "Failed", "value": fmt.Sprintf("%d", failed)},
					{"name": "Total", "value": fmt.Sprintf("%d", total)},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("teams webhook returned %d", resp.StatusCode)
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
