package schedule_test

import (
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/schedule"
)

func TestCalcNextRun_Daily(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	next := schedule.CalcNextRun(schedule.Daily, "", now)
	expected := now.Add(24 * time.Hour)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCalcNextRun_Weekly(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	next := schedule.CalcNextRun(schedule.Weekly, "", now)
	expected := now.Add(7 * 24 * time.Hour)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCalcNextRun_Cron(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	// Every day at 3am
	next := schedule.CalcNextRun(schedule.Cron, "0 3 * * *", now)
	// Should be next day at 3am since we're past 3am today
	expected := time.Date(2026, 6, 2, 3, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCalcNextRun_CronEvery5Min(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 2, 0, 0, time.UTC)
	next := schedule.CalcNextRun(schedule.Cron, "*/5 * * * *", now)
	expected := time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestCalcNextRun_InvalidCron(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	// Invalid cron should fallback to daily
	next := schedule.CalcNextRun(schedule.Cron, "invalid", now)
	expected := now.Add(24 * time.Hour)
	if !next.Equal(expected) {
		t.Fatalf("expected daily fallback %v, got %v", expected, next)
	}
}

func TestGetDue(t *testing.T) {
	s := schedule.NewStore()
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	s.Create(&schedule.Schedule{Name: "due", Enabled: true, NextRunAt: past})
	s.Create(&schedule.Schedule{Name: "not-due", Enabled: true, NextRunAt: future})
	s.Create(&schedule.Schedule{Name: "disabled", Enabled: false, NextRunAt: past})

	due := s.GetDue(time.Now())
	if len(due) != 1 {
		t.Fatalf("expected 1 due schedule, got %d", len(due))
	}
	if due[0].Name != "due" {
		t.Fatalf("expected 'due', got %s", due[0].Name)
	}
}

func TestClaimNextDue_ClaimsAtomically(t *testing.T) {
	s := schedule.NewStore()
	now := time.Now()
	past := now.Add(-1 * time.Hour)

	s.Create(&schedule.Schedule{
		ID: "sched-1", Name: "first-due", Enabled: true,
		NextRunAt: past, Frequency: schedule.Daily,
	})
	s.Create(&schedule.Schedule{
		ID: "sched-2", Name: "second-due", Enabled: true,
		NextRunAt: past.Add(30 * time.Minute), Frequency: schedule.Weekly,
	})

	// First claim gets the earliest-due schedule
	claimed := s.ClaimNextDue(now, "worker-1")
	if claimed == nil {
		t.Fatal("expected a claimed schedule, got nil")
	}
	if claimed.Name != "first-due" {
		t.Fatalf("expected 'first-due', got %s", claimed.Name)
	}

	// next_run_at should have been advanced as part of the atomic claim
	if !claimed.NextRunAt.After(now) {
		t.Fatal("expected next_run_at to be advanced after claim")
	}

	// Second claim gets the next-due schedule
	claimed2 := s.ClaimNextDue(now, "worker-2")
	if claimed2 == nil {
		t.Fatal("expected a second claimed schedule, got nil")
	}
	if claimed2.Name != "second-due" {
		t.Fatalf("expected 'second-due', got %s", claimed2.Name)
	}
}

func TestClaimNextDue_NoDueReturnsNil(t *testing.T) {
	s := schedule.NewStore()
	future := time.Now().Add(1 * time.Hour)

	s.Create(&schedule.Schedule{
		ID: "future-only", Name: "future", Enabled: true,
		NextRunAt: future, Frequency: schedule.Daily,
	})

	claimed := s.ClaimNextDue(time.Now(), "worker-1")
	if claimed != nil {
		t.Fatalf("expected nil claim, got %s", claimed.Name)
	}
}

func TestClaimNextDue_SkipsDisabled(t *testing.T) {
	s := schedule.NewStore()
	past := time.Now().Add(-1 * time.Hour)
	now := time.Now()

	s.Create(&schedule.Schedule{
		ID: "disabled", Name: "disabled-sched", Enabled: false,
		NextRunAt: past, Frequency: schedule.Daily,
	})

	claimed := s.ClaimNextDue(now, "worker-1")
	if claimed != nil {
		t.Fatalf("expected nil claim for disabled schedule, got %s", claimed.Name)
	}
}
