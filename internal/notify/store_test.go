package notify

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestStore_AddAssignsIDAndTimestamp(t *testing.T) {
	s := NewStore()
	n1 := s.Add(Notification{RunID: "r1", Type: "failure", Message: "boom"})
	n2 := s.Add(Notification{RunID: "r2", Type: "flake", Message: "flaky"})
	if n1.ID != "notif-1" || n2.ID != "notif-2" {
		t.Fatalf("expected sequential IDs, got %q %q", n1.ID, n2.ID)
	}
	if n1.CreatedAt.IsZero() {
		t.Fatal("CreatedAt not set")
	}
}

func TestStore_ListReturnsCopy(t *testing.T) {
	s := NewStore()
	s.Add(Notification{RunID: "r1", Message: "m"})
	list := s.List()
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}
	// Mutating the returned slice must not affect the store.
	list[0].Message = "hacked"
	fresh := s.List()
	if fresh[0].Message != "m" {
		t.Fatalf("List leaked shared backing state: %q", fresh[0].Message)
	}
}

func TestStore_ByRunFilters(t *testing.T) {
	s := NewStore()
	s.Add(Notification{RunID: "r1", Message: "a"})
	s.Add(Notification{RunID: "r2", Message: "b"})
	s.Add(Notification{RunID: "r1", Message: "c"})

	got := s.ByRun("r1")
	if len(got) != 2 {
		t.Fatalf("expected 2 for r1, got %d", len(got))
	}
	if s.ByRun("missing") != nil {
		t.Fatal("expected nil for unknown run")
	}
}

func TestDeliverWebhook_EmptyURLNoop(t *testing.T) {
	if err := DeliverWebhook("", map[string]string{"x": "y"}); err != nil {
		t.Fatalf("empty URL should be a noop, got %v", err)
	}
}

func TestDeliverWebhook_PostsPayload(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := DeliverWebhook(srv.URL, map[string]string{"type": "failure"}); err != nil {
		t.Fatalf("DeliverWebhook: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 webhook hit, got %d", hits)
	}
}

func TestDeliverWebhook_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	if err := DeliverWebhook(srv.URL, nil); err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestTriggerFailure_MarksDeliveredOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewStore()
	s.TriggerFailure("r1", "sched-1", srv.URL, "it broke")

	list := s.ByRun("r1")
	if len(list) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(list))
	}
	if list[0].Type != "failure" || !list[0].Delivered {
		t.Fatalf("expected delivered failure notification, got %+v", list[0])
	}
}

func TestTriggerFailure_NotDeliveredOnWebhookError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	s := NewStore()
	s.TriggerFailure("r1", "", srv.URL, "boom")
	list := s.ByRun("r1")
	if len(list) != 1 || list[0].Delivered {
		t.Fatalf("expected undelivered notification on webhook error, got %+v", list)
	}
}

func TestTriggerFailure_NoWebhookStillRecords(t *testing.T) {
	s := NewStore()
	s.TriggerFailure("r1", "", "", "boom")
	list := s.ByRun("r1")
	if len(list) != 1 {
		t.Fatalf("expected notification recorded without webhook, got %d", len(list))
	}
	if list[0].Delivered {
		t.Fatal("no webhook means not delivered")
	}
}
