package webhook

import "testing"

func TestRegistrationStore_Create(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{
		RepositoryURL: "https://github.com/owner/repo",
		GithubToken:   "ghp_test",
	})
	if r.ID == "" {
		t.Fatal("expected ID to be assigned")
	}
	if r.Status != "active" {
		t.Errorf("expected active status, got %q", r.Status)
	}
}

func TestRegistrationStore_List(t *testing.T) {
	s := NewRegistrationStore()
	s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})
	s.Create(WebhookRegistration{RepositoryURL: "https://github.com/c/d"})

	list := s.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 registrations, got %d", len(list))
	}
}

func TestRegistrationStore_Get(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})

	got, ok := s.Get(r.ID)
	if !ok {
		t.Fatal("expected to find registration")
	}
	if got.RepositoryURL != "https://github.com/a/b" {
		t.Errorf("wrong repository URL: %s", got.RepositoryURL)
	}
}

func TestRegistrationStore_GetNotFound(t *testing.T) {
	s := NewRegistrationStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRegistrationStore_UpdateStatus(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})

	updated, err := s.UpdateStatus(r.ID, "inactive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != "inactive" {
		t.Errorf("expected inactive, got %s", updated.Status)
	}
}

func TestRegistrationStore_UpdateStatusInvalid(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})

	_, err := s.UpdateStatus(r.ID, "bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestRegistrationStore_Delete(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})

	if !s.Delete(r.ID) {
		t.Fatal("expected delete to succeed")
	}
	if _, ok := s.Get(r.ID); ok {
		t.Fatal("expected registration to be gone")
	}
}

func TestRegistrationStore_DeleteNotFound(t *testing.T) {
	s := NewRegistrationStore()
	if s.Delete("nonexistent") {
		t.Fatal("expected delete to return false for nonexistent")
	}
}

func TestRegistrationStore_UpdateLastSync(t *testing.T) {
	s := NewRegistrationStore()
	r := s.Create(WebhookRegistration{RepositoryURL: "https://github.com/a/b"})

	if !s.UpdateLastSync(r.ID) {
		t.Fatal("expected UpdateLastSync to succeed")
	}
	got, _ := s.Get(r.ID)
	if got.LastSyncAt == nil {
		t.Fatal("expected LastSyncAt to be set")
	}
}
