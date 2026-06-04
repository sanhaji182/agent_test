package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var store db.RunStore
	var settingsStore *db.SettingsStore

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
			settingsStore = db.NewSettingsStore(pgStore.Pool())
		}
	} else {
		store = db.NewMemoryStore()
	}

	srv := api.NewServer(cfg, store, settingsStore)

	// Background scheduler: polls due schedules every 60s and enqueues runs
	go runScheduler(srv)

	addr := ":" + cfg.AppPort
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func runScheduler(srv *api.Server) {
	slog.Info("scheduler started", "interval", "60s")
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("scheduler panic recovered", "error", r)
				}
			}()
			processed := srv.ProcessDueSchedules(context.Background(), time.Now())
			if processed > 0 {
				slog.Info("scheduler: processed due schedules", "count", processed)
			}
		}()
		time.Sleep(60 * time.Second)
	}
}
