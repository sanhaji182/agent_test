package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/queue"
	"github.com/go-go-golems/gotest-agent/internal/tracing"
	"github.com/hibiken/asynq"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv != "development" && cfg.APIKey == "" {
		slog.Error("API_KEY is required in production mode (set APP_ENV=development to bypass)")
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize distributed tracing
	tracingShutdown, err := tracing.Init(ctx, tracing.Config{
		Enabled:           cfg.TracingEnabled,
		ExporterEndpoint:  cfg.TracingEndpoint,
		ServiceName:       "gotest-agent",
		ServiceVersion:    cfg.ServiceVersion,
		Environment:       cfg.AppEnv,
	})
	if err != nil {
		slog.Warn("failed to initialize tracing", "error", err)
	} else {
		defer tracingShutdown(ctx)
	}

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

	// Optional durable queue (QUEUE_ENABLED=true): enqueue runs to Redis/Asynq
	// and execute them via a worker in this process. Default remains in-process
	// goroutines (agent.Launch) — no Redis required.
	if cfg.QueueEnabled {
		client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
		defer client.Close()
		worker := queue.NewRunWorker(cfg.RedisURL, srv, cfg.MaxConcurrentRuns)
		if err := worker.Start(); err != nil {
			slog.Error("queue worker failed to start; using in-process execution", "error", err)
		} else {
			defer worker.Stop()
			srv.SetRunEnqueuer(func(runID string) error {
				return queue.EnqueueRunByID(client, runID)
			})
			slog.Info("durable run queue enabled", "redis", cfg.RedisURL, "concurrency", cfg.MaxConcurrentRuns)
		}
	}

	// Background scheduler: polls due schedules every 60s and enqueues runs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runScheduler(ctx, srv)

	// Failure notifier: records notifications (+ schedule webhooks) on run_failed
	go srv.StartFailureNotifier(ctx)

	addr := ":" + cfg.AppPort
	slog.Info("starting server", "addr", addr)
	hs := &http.Server{
		Addr:              addr,
		Handler:           srv,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig.String())
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := hs.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			hs.Close()
		}
		close(idleConnsClosed)
	}()

	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		cancel()
		os.Exit(1)
	}
	<-idleConnsClosed
	slog.Info("server stopped")
}

func runScheduler(ctx context.Context, srv *api.Server) {
	slog.Info("scheduler started", "interval", "60s")
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler stopped")
			return
		case <-ticker.C:
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
		}
	}
}
