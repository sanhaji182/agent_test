package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
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

	addr := ":" + cfg.AppPort
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
