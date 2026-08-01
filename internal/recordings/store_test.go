package recordings

import (
	"testing"
	"time"
)

func TestStore_AddAssignsSequentialIDs(t *testing.T) {
	s := NewStore()
	r1 := s.Add(Recording{RunID: "run-1", StepName: "a"})
	r2 := s.Add(Recording{RunID: "run-1", StepName: "b"})
	if r1.ID != "run-1-rec-1" || r2.ID != "run-1-rec-2" {
		t.Fatalf("expected sequential IDs, got %q %q", r1.ID, r2.ID)
	}
}

func TestStore_AddPreservesExplicitID(t *testing.T) {
	s := NewStore()
	r := s.Add(Recording{ID: "custom-id", RunID: "run-1"})
	if r.ID != "custom-id" {
		t.Fatalf("explicit ID overwritten: %q", r.ID)
	}
}

func TestStore_ByRunFilters(t *testing.T) {
	s := NewStore()
	s.Add(Recording{RunID: "run-1", StepName: "a"})
	s.Add(Recording{RunID: "run-2", StepName: "b"})
	s.Add(Recording{RunID: "run-1", StepName: "c"})

	got := s.ByRun("run-1")
	if len(got) != 2 {
		t.Fatalf("expected 2 for run-1, got %d", len(got))
	}
	if s.ByRun("missing") != nil {
		t.Fatal("expected nil for unknown run")
	}
}

func TestStore_AllReturnsCopy(t *testing.T) {
	s := NewStore()
	s.Add(Recording{RunID: "run-1", Status: "captured"})
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}
	// Mutating the returned slice must not affect the store's backing array.
	all[0].Status = "hacked"
	fresh := s.All()
	if fresh[0].Status != "captured" {
		t.Fatalf("All leaked shared backing state: %q", fresh[0].Status)
	}
}

func TestItoa(t *testing.T) {
	cases := map[int64]string{0: "0", 1: "1", 9: "9", 10: "10", 42: "42", 100: "100", 123456789: "123456789"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Fatalf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestStore_CreateSession(t *testing.T) {
	tests := []struct {
		name string
		sess Session
		want Session
	}{
		{
			name: "generates id and timestamps",
			sess: Session{Name: "login flow", ProjectPath: "/app", BaseURL: "http://localhost:3000"},
			want: Session{Name: "login flow", ProjectPath: "/app", BaseURL: "http://localhost:3000", Status: "recording"},
		},
		{
			name: "preserves explicit id and status",
			sess: Session{ID: "sess-abc", Name: "checkout", Status: "paused"},
			want: Session{ID: "sess-abc", Name: "checkout", Status: "paused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore()
			got := s.CreateSession(tt.sess)
			if tt.want.ID != "" && got.ID != tt.want.ID {
				t.Fatalf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Fatalf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Status != tt.want.Status {
				t.Fatalf("Status = %q, want %q", got.Status, tt.want.Status)
			}
			if got.CreatedAt.IsZero() {
				t.Fatal("CreatedAt is zero")
			}
			if got.UpdatedAt.IsZero() {
				t.Fatal("UpdatedAt is zero")
			}
		})
	}
}

func TestStore_GetSession(t *testing.T) {
	s := NewStore()
	created := s.CreateSession(Session{Name: "login"})

	got, ok := s.GetSession(created.ID)
	if !ok {
		t.Fatal("expected session to be found")
	}
	if got.Name != "login" {
		t.Fatalf("Name = %q, want %q", got.Name, "login")
	}

	_, ok = s.GetSession("missing")
	if ok {
		t.Fatal("expected missing session to not be found")
	}
}

func TestStore_ListSessions(t *testing.T) {
	s := NewStore()
	s1 := s.CreateSession(Session{Name: "first"})
	time.Sleep(10 * time.Millisecond)
	s2 := s.CreateSession(Session{Name: "second"})

	list := s.ListSessions()
	if len(list) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(list))
	}
	if list[0].ID != s2.ID || list[1].ID != s1.ID {
		t.Fatalf("expected descending order, got %v", list)
	}
}

func TestStore_UpdateSessionStatus(t *testing.T) {
	s := NewStore()
	created := s.CreateSession(Session{Name: "login"})

	got, ok := s.UpdateSessionStatus(created.ID, "completed")
	if !ok {
		t.Fatal("expected update to succeed")
	}
	if got.Status != "completed" {
		t.Fatalf("Status = %q, want %q", got.Status, "completed")
	}
	if got.UpdatedAt.Before(created.UpdatedAt) {
		t.Fatal("UpdatedAt should not be before the original")
	}
	fresh, ok := s.GetSession(created.ID)
	if !ok || fresh.Status != "completed" {
		t.Fatal("expected persisted status update")
	}

	_, ok = s.UpdateSessionStatus("missing", "completed")
	if ok {
		t.Fatal("expected update for missing session to fail")
	}
}

func TestStore_DeleteSession(t *testing.T) {
	s := NewStore()
	created := s.CreateSession(Session{Name: "login"})
	s.AddEvent(Event{SessionID: created.ID, EventType: EventClick})

	if ok := s.DeleteSession(created.ID); !ok {
		t.Fatal("expected delete to succeed")
	}
	if _, ok := s.GetSession(created.ID); ok {
		t.Fatal("expected session to be deleted")
	}
	if evs := s.GetEventsBySession(created.ID); evs != nil {
		t.Fatalf("expected events to be deleted, got %d", len(evs))
	}

	if ok := s.DeleteSession("missing"); ok {
		t.Fatal("expected delete for missing session to fail")
	}
}

func TestStore_AddEvent(t *testing.T) {
	s := NewStore()
	sess := s.CreateSession(Session{Name: "login"})

	ev1 := s.AddEvent(Event{SessionID: sess.ID, EventType: EventClick, Selector: "#btn"})
	ev2 := s.AddEvent(Event{SessionID: sess.ID, EventType: EventFill, Selector: "#email", Value: "a@b.com"})

	if ev1.ID == "" || ev2.ID == "" {
		t.Fatal("expected event IDs to be generated")
	}
	if ev1.SequenceOrder != 1 {
		t.Fatalf("expected first event sequence 1, got %d", ev1.SequenceOrder)
	}
	if ev2.SequenceOrder != 2 {
		t.Fatalf("expected second event sequence 2, got %d", ev2.SequenceOrder)
	}
}

func TestStore_AddEventOrphansNotAssignedToSession(t *testing.T) {
	s := NewStore()
	ev := s.AddEvent(Event{SessionID: "missing", EventType: EventClick})
	if ev.SequenceOrder != 0 {
		t.Fatalf("expected orphan event sequence 0, got %d", ev.SequenceOrder)
	}
}

func TestStore_GetEventsBySession(t *testing.T) {
	s := NewStore()
	sess := s.CreateSession(Session{Name: "login"})

	s.AddEvent(Event{SessionID: sess.ID, EventType: EventNavigate, URL: "/login", SequenceOrder: 1})
	s.AddEvent(Event{SessionID: sess.ID, EventType: EventClick, Selector: "#btn", SequenceOrder: 2})

	evs := s.GetEventsBySession(sess.ID)
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
	if evs[0].EventType != EventNavigate {
		t.Fatalf("expected first event navigate, got %q", evs[0].EventType)
	}
	if evs[1].EventType != EventClick {
		t.Fatalf("expected second event click, got %q", evs[1].EventType)
	}

	if evs := s.GetEventsBySession("missing"); evs != nil {
		t.Fatalf("expected nil for missing session, got %d", len(evs))
	}
}

func TestStore_GetSessionWithEvents(t *testing.T) {
	s := NewStore()
	sess := s.CreateSession(Session{Name: "login"})
	s.AddEvent(Event{SessionID: sess.ID, EventType: EventClick})

	gotSess, evs, ok := s.GetSessionWithEvents(sess.ID)
	if !ok {
		t.Fatal("expected session with events to be found")
	}
	if gotSess.Name != "login" {
		t.Fatalf("Name = %q, want %q", gotSess.Name, "login")
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}

	_, _, ok = s.GetSessionWithEvents("missing")
	if ok {
		t.Fatal("expected missing session to not be found")
	}
}

func TestStore_EventIsolation(t *testing.T) {
	s := NewStore()
	sess := s.CreateSession(Session{Name: "login"})
	s.AddEvent(Event{SessionID: sess.ID, EventType: EventClick})

	evs := s.GetEventsBySession(sess.ID)
	evs[0].EventType = EventFill

	fresh := s.GetEventsBySession(sess.ID)
	if fresh[0].EventType != EventClick {
		t.Fatalf("GetEventsBySession leaked shared state: %q", fresh[0].EventType)
	}
}

func TestEventTypes(t *testing.T) {
	events := []Event{
		{EventType: EventClick},
		{EventType: EventFill},
		{EventType: EventClick},
		{EventType: EventAssertVisible},
	}
	got := EventTypes(events)
	want := []EventType{EventClick, EventFill, EventAssertVisible}
	if len(got) != len(want) {
		t.Fatalf("expected %d types, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("type[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEvent_MetadataString(t *testing.T) {
	ev := Event{Metadata: map[string]interface{}{"key": "value"}}
	if got := ev.MetadataString(); got != `{"key":"value"}` {
		t.Fatalf("MetadataString = %q, want %q", got, `{"key":"value"}`)
	}

	empty := Event{}
	if got := empty.MetadataString(); got != "" {
		t.Fatalf("empty MetadataString = %q, want empty", got)
	}
}
