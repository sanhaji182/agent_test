package recordings

import (
	"context"
	"encoding/json"
	"time"
)

func recordingDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *Store) persistSessionDB(sess *Session) error {
	ctx, cancel := recordingDBContext()
	defer cancel()
	metadata, _ := json.Marshal(sess.Metadata)
	if metadata == nil {
		metadata = []byte("{}")
	}
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO recording_sessions (id, name, project_path, base_url, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			project_path = EXCLUDED.project_path,
			base_url = EXCLUDED.base_url,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at`,
		sess.ID, sess.Name, sess.ProjectPath, sess.BaseURL, sess.Status, metadata, sess.CreatedAt, sess.UpdatedAt)
	return err
}

func (s *Store) getSessionDB(id string) (*Session, error) {
	ctx, cancel := recordingDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, name, project_path, base_url, status, metadata, created_at, updated_at
		FROM recording_sessions WHERE id = $1`, id)
	return scanSession(row)
}

func (s *Store) listSessionsDB() ([]Session, error) {
	ctx, cancel := recordingDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, name, project_path, base_url, status, metadata, created_at, updated_at
		FROM recording_sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return sessions, err
		}
		sessions = append(sessions, *sess)
	}
	return sessions, rows.Err()
}

func scanSession(row interface{ Scan(dest ...any) error }) (*Session, error) {
	var sess Session
	var metadata []byte
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProjectPath, &sess.BaseURL, &sess.Status, &metadata, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
		return nil, err
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &sess.Metadata)
	}
	return &sess, nil
}

func (s *Store) deleteSessionDB(id string) error {
	ctx, cancel := recordingDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `DELETE FROM recording_sessions WHERE id = $1`, id)
	return err
}

func (s *Store) persistEventDB(ev *Event) error {
	ctx, cancel := recordingDBContext()
	defer cancel()
	metadata, _ := json.Marshal(ev.Metadata)
	if metadata == nil {
		metadata = []byte("{}")
	}
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO recorded_events (id, session_id, event_type, selector, value, url, timestamp, metadata, sequence_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			event_type = EXCLUDED.event_type,
			selector = EXCLUDED.selector,
			value = EXCLUDED.value,
			url = EXCLUDED.url,
			timestamp = EXCLUDED.timestamp,
			metadata = EXCLUDED.metadata,
			sequence_order = EXCLUDED.sequence_order`,
		ev.ID, ev.SessionID, string(ev.EventType), ev.Selector, ev.Value, ev.URL, ev.Timestamp, metadata, ev.SequenceOrder, ev.CreatedAt)
	return err
}

func (s *Store) getEventsBySessionDB(sessionID string) ([]Event, error) {
	ctx, cancel := recordingDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, session_id, event_type, selector, value, url, timestamp, metadata, sequence_order, created_at
		FROM recorded_events WHERE session_id = $1 ORDER BY sequence_order ASC, created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var ev Event
		var eventType string
		var metadata []byte
		if err := rows.Scan(&ev.ID, &ev.SessionID, &eventType, &ev.Selector, &ev.Value, &ev.URL, &ev.Timestamp, &metadata, &ev.SequenceOrder, &ev.CreatedAt); err != nil {
			return events, err
		}
		ev.EventType = EventType(eventType)
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &ev.Metadata)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
