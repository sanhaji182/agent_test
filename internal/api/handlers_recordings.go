package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
)

// handleCreateRecordingSession creates a new recording session.
func (s *Server) handleCreateRecordingSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string                 `json:"name"`
		ProjectPath string                 `json:"project_path"`
		BaseURL     string                 `json:"base_url"`
		Metadata    map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.ProjectPath == "" {
		writeJSONError(w, http.StatusBadRequest, "project_path is required")
		return
	}
	if body.BaseURL == "" {
		writeJSONError(w, http.StatusBadRequest, "base_url is required")
		return
	}
	sess := s.recordings.CreateSession(recordings.Session{
		Name:        body.Name,
		ProjectPath: body.ProjectPath,
		BaseURL:     body.BaseURL,
		Metadata:    body.Metadata,
	})
	writeJSON(w, http.StatusCreated, sess)
}

// handleListRecordingSessions returns all recording sessions.
func (s *Server) handleListRecordingSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.recordings.ListSessions()
	if sessions == nil {
		sessions = []recordings.Session{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

// handleGetRecordingSession returns a session with its events.
func (s *Server) handleGetRecordingSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, events, ok := s.recordings.GetSessionWithEvents(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": sess,
		"events":  events,
	})
}

// handleAddRecordingEvent adds an event to a recording session.
func (s *Server) handleAddRecordingEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := s.recordings.GetSession(id); !ok {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	var body struct {
		EventType     string                 `json:"event_type"`
		Selector      string                 `json:"selector,omitempty"`
		Value         string                 `json:"value,omitempty"`
		URL           string                 `json:"url,omitempty"`
		Metadata      map[string]interface{} `json:"metadata,omitempty"`
		SequenceOrder *int                   `json:"sequence_order,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.EventType == "" {
		writeJSONError(w, http.StatusBadRequest, "event_type is required")
		return
	}
	ev := s.recordings.AddEvent(recordings.Event{
		SessionID: id,
		EventType: recordings.EventType(body.EventType),
		Selector:  body.Selector,
		Value:     body.Value,
		URL:       body.URL,
		Metadata:  body.Metadata,
	})
	writeJSON(w, http.StatusCreated, ev)
}

// handleListRecordingEvents returns all events for a session.
func (s *Server) handleListRecordingEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := s.recordings.GetSession(id); !ok {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	events := s.recordings.GetEventsBySession(id)
	if events == nil {
		events = []recordings.Event{}
	}
	writeJSON(w, http.StatusOK, events)
}

// handleGenerateTestFromRecording generates a test from a recording session.
func (s *Server) handleGenerateTestFromRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, events, ok := s.recordings.GetSessionWithEvents(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	if len(events) == 0 {
		writeJSONError(w, http.StatusBadRequest, "no events in session")
		return
	}
	// Generate a basic Playwright TypeScript skeleton from recorded events
	code := generatePlaywrightSkeleton(sess, events)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"test_code": code,
		"language":  "typescript",
		"framework": "playwright",
	})
}

func generatePlaywrightSkeleton(sess *recordings.Session, events []recordings.Event) string {
	var b strings.Builder
	b.WriteString("import { test, expect } from '@playwright/test';\n\n")
	fmt.Fprintf(&b, "test('%s', async ({ page }) => {\n", sess.Name)
	fmt.Fprintf(&b, "  // Navigate to base URL\n")
	fmt.Fprintf(&b, "  await page.goto('%s');\n\n", sess.BaseURL)
	for _, ev := range events {
		switch ev.EventType {
		case recordings.EventClick:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  // Click on %s\n", ev.Selector)
				fmt.Fprintf(&b, "  await page.click('%s');\n", sanitizeForTS(ev.Selector))
			}
		case recordings.EventFill:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  // Fill %s\n", ev.Selector)
				fmt.Fprintf(&b, "  await page.fill('%s', '%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventNavigate:
			if ev.URL != "" {
				fmt.Fprintf(&b, "  // Navigate to %s\n", ev.URL)
				fmt.Fprintf(&b, "  await page.goto('%s');\n", sanitizeForTS(ev.URL))
			}
		case recordings.EventSelect:
			if ev.Selector != "" && ev.Value != "" {
				fmt.Fprintf(&b, "  // Select '%s' from %s\n", ev.Value, ev.Selector)
				fmt.Fprintf(&b, "  await page.selectOption('%s', '%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventHover:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  // Hover over %s\n", ev.Selector)
				fmt.Fprintf(&b, "  await page.hover('%s');\n", sanitizeForTS(ev.Selector))
			}
		case recordings.EventPress:
			if ev.Value != "" {
				fmt.Fprintf(&b, "  // Press %s\n", ev.Value)
				fmt.Fprintf(&b, "  await page.keyboard.press('%s');\n", sanitizeForTS(ev.Value))
			}
		case recordings.EventScroll:
			fmt.Fprintf(&b, "  // Scroll\n")
			fmt.Fprintf(&b, "  await page.evaluate(() => window.scrollBy(0, 300));\n")
		case recordings.EventAssertText:
			if ev.Selector != "" && ev.Value != "" {
				fmt.Fprintf(&b, "  // Assert text of %s\n", ev.Selector)
				fmt.Fprintf(&b, "  await expect(page.locator('%s')).toContainText('%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventAssertVisible:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  // Assert %s is visible\n", ev.Selector)
				fmt.Fprintf(&b, "  await expect(page.locator('%s')).toBeVisible();\n", sanitizeForTS(ev.Selector))
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("});\n")
	return b.String()
}

func sanitizeForTS(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// handleDeleteRecordingSession deletes a recording session and its events.
func (s *Server) handleDeleteRecordingSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.recordings.DeleteSession(id) {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateRecordingSession updates a recording session's name, status, or metadata.
func (s *Server) handleUpdateRecordingSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := s.recordings.GetSession(id); !ok {
		writeJSONError(w, http.StatusNotFound, "session not found")
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	s.recordings.UpdateSession(id, func(sess *recordings.Session) {
		if name, ok := safeString(body, "name"); ok {
			sess.Name = name
		}
		if st, ok := safeString(body, "status"); ok {
			sess.Status = st
		}
		if md, ok := body["metadata"].(map[string]interface{}); ok {
			sess.Metadata = md
		}
	})
	updated, _ := s.recordings.GetSession(id)
	writeJSON(w, http.StatusOK, updated)
}
