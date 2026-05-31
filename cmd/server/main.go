package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var store db.RunStore

	if cfg.DatabaseURL != "" {
		pgStore, err := db.NewStore(ctx, cfg.DatabaseURL)
		if err != nil {
			slog.Warn("db not available, using in-memory store", "error", err)
			store = db.NewMemoryStore()
		} else {
			defer pgStore.Close()
			if err := db.RunMigrations(ctx, cfg.DatabaseURL); err != nil {
				slog.Warn("migrations failed", "error", err)
			}
			store = pgStore
		}
	} else {
		store = db.NewMemoryStore()
	}

	srv := api.NewServer(cfg, store)

	// Background scheduler: polls due schedules every 60s and enqueues runs
	go runScheduler(srv.Schedules(), store)

	addr := ":" + cfg.AppPort
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func runScheduler(schedStore *schedule.Store, runStore db.RunStore) {
	slog.Info("scheduler started", "interval", "60s")
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("scheduler panic recovered", "error", r)
				}
			}()
			processDueSchedules(schedStore, runStore)
		}()
		time.Sleep(60 * time.Second)
	}
}

func processDueSchedules(schedStore *schedule.Store, runStore db.RunStore) {
	now := time.Now()
	due := schedStore.GetDue(now)
	if len(due) == 0 {
		return
	}

	slog.Info("scheduler: found due schedules", "count", len(due))
	ctx := context.Background()

	for _, sch := range due {
		run := &agent.TestRun{
			ID:           uuid.New().String(),
			ProjectPath:  sch.ProjectPath,
			Requirements: sch.Requirements,
			Mode:         sch.Mode,
			State:        agent.StateIdle,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := runStore.CreateRun(ctx, run); err != nil {
			slog.Error("scheduler: failed to create run", "schedule", sch.Name, "error", err)
			continue
		}

		// Update schedule with last run info and compute next run
		schedStore.Update(sch.ID, func(s *schedule.Schedule) {
			s.LastRunAt = &now
			s.LastRunID = run.ID
			s.LastRunStatus = string(run.State)
			s.NextRunAt = schedule.CalcNextRun(s.Frequency, s.CronExpr, now)
		})

		slog.Info("scheduler: enqueued run", "schedule", sch.Name, "run_id", run.ID, "next_run", schedule.CalcNextRun(sch.Frequency, sch.CronExpr, now).Format(time.RFC3339))
	}
}
