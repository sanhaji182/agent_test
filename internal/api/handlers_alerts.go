package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/notify"
)

// Alert adalah view API dari notification dengan severity & category yang
// dihitung server-side, supaya frontend tidak perlu tahu aturan pemetaan dan
// selalu mendapat data yang konsisten.
type Alert struct {
	ID           string    `json:"id"`
	RunID        string    `json:"run_id"`
	ScheduleID   string    `json:"schedule_id,omitempty"`
	Type         string    `json:"type"`
	Severity     string    `json:"severity"` // "critical" | "warning" | "info"
	Category     string    `json:"category"` // "failure" | "drift" | "system"
	Message      string    `json:"message"`
	Delivered    bool      `json:"delivered"`
	Acknowledged bool      `json:"acknowledged"`
	Dismissed    bool      `json:"dismissed"`
	CreatedAt    time.Time `json:"created_at"`
}

// alertSeverity memetakan tipe notification ke tingkat keparahan.
func alertSeverity(t string) string {
	switch t {
	case "failure":
		return "critical"
	case "flake", "degradation":
		return "warning"
	default:
		return "info"
	}
}

// alertCategory mengelompokkan tipe notification untuk filter UI.
func alertCategory(t string) string {
	switch t {
	case "failure":
		return "failure"
	case "flake", "degradation":
		return "drift"
	default:
		return "system"
	}
}

func toAlert(n notify.Notification) Alert {
	return Alert{
		ID:           n.ID,
		RunID:        n.RunID,
		ScheduleID:   n.ScheduleID,
		Type:         n.Type,
		Severity:     alertSeverity(n.Type),
		Category:     alertCategory(n.Type),
		Message:      n.Message,
		Delivered:    n.Delivered,
		Acknowledged: n.Acknowledged,
		Dismissed:    n.Dismissed,
		CreatedAt:    n.CreatedAt,
	}
}

// handleListAlerts mengembalikan daftar alert (notifications) dengan severity
// & category server-side. Mendukung query params:
//
//	?type=failure|flake|degradation|all  (default all, filter by type/category)
//	?limit=N                             (batasi jumlah item terbaru)
//	?include_dismissed=true|false        (default false — alert yang di-dismiss disembunyikan)
func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	qType := q.Get("type")
	if qType == "" {
		qType = "all"
	}
	limit := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	includeDismissed := q.Get("include_dismissed") == "true"

	list := s.notifs.List()
	alerts := make([]Alert, 0, len(list))
	for _, n := range list {
		if n.Dismissed && !includeDismissed {
			continue
		}
		if qType != "all" && n.Type != qType && alertCategory(n.Type) != qType {
			continue
		}
		alerts = append(alerts, toAlert(n))
	}
	if limit > 0 && len(alerts) > limit {
		alerts = alerts[:limit]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleAcknowledgeAlert menandai satu alert sebagai sudah dibaca.
func (s *Server) handleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.notifs.Acknowledge(id) {
		writeJSONError(w, http.StatusNotFound, "alert not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"acknowledged": true})
}

// handleDismissAlert menyembunyikan satu alert dari daftar aktif.
func (s *Server) handleDismissAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.notifs.Dismiss(id) {
		writeJSONError(w, http.StatusNotFound, "alert not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"dismissed": true})
}

// handleMarkAllAlertsRead menandai semua alert sebagai sudah dibaca.
func (s *Server) handleMarkAllAlertsRead(w http.ResponseWriter, r *http.Request) {
	count := s.notifs.MarkAllRead()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"updated": count})
}
