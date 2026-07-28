package project

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMemoryStore_CreateAppliesDefaults(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	p := &Project{Name: "shop"}
	if err := s.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// prepareProject contract.
	if p.ID == "" || p.TestType != "ui" || p.Environment != "default" || p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		t.Fatalf("create defaults not applied: %+v", p)
	}

	// Explicit values must be preserved.
	p2 := &Project{Name: "api-proj", TestType: "api", Environment: "staging"}
	if err := s.Create(ctx, p2); err != nil {
		t.Fatalf("Create p2: %v", err)
	}
	if p2.TestType != "api" || p2.Environment != "staging" {
		t.Fatalf("explicit values overwritten: %+v", p2)
	}
}

func TestMemoryStore_GetAndUpdate(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	p := &Project{Name: "shop"}
	s.Create(ctx, p)

	got, err := s.Get(ctx, p.ID)
	if err != nil || got.Name != "shop" {
		t.Fatalf("Get: %v / %+v", err, got)
	}

	got.Name = "renamed"
	before := got.UpdatedAt
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	fresh, _ := s.Get(ctx, p.ID)
	if fresh.Name != "renamed" {
		t.Fatalf("update not persisted: %+v", fresh)
	}
	if !fresh.UpdatedAt.After(before) && !fresh.UpdatedAt.Equal(before) {
		t.Fatalf("UpdatedAt went backwards: %v -> %v", before, fresh.UpdatedAt)
	}
}

func TestMemoryStore_NotFoundPaths(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if _, err := s.Get(ctx, "missing"); err != pgx.ErrNoRows {
		t.Fatalf("Get missing: expected pgx.ErrNoRows, got %v", err)
	}
	if err := s.Update(ctx, &Project{ID: "missing"}); err != pgx.ErrNoRows {
		t.Fatalf("Update missing: expected pgx.ErrNoRows, got %v", err)
	}
}

func TestMemoryStore_ListPaginationAndOrder(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	for _, name := range []string{"p1", "p2", "p3"} {
		if err := s.Create(ctx, &Project{Name: name}); err != nil {
			t.Fatalf("Create(%s): %v", name, err)
		}
	}

	// Newest first.
	all, err := s.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 || all[0].Name != "p3" || all[2].Name != "p1" {
		t.Fatalf("expected [p3 p2 p1], got %v", []string{all[0].Name, all[1].Name, all[2].Name})
	}

	// Window.
	page, _ := s.List(ctx, 1, 1)
	if len(page) != 1 || page[0].Name != "p2" {
		t.Fatalf("expected [p2], got %+v", page)
	}

	// Past the end: empty slice, never nil (handlers encode this as JSON []).
	empty, err := s.List(ctx, 10, 99)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("expected empty non-nil slice past end, got %v / %v", empty, err)
	}
}
