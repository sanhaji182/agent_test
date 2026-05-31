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
