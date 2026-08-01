package db_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/release"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresWorkflowPersistenceSmoke(t *testing.T) {
	databaseURL := os.Getenv("GOTEST_POSTGRES_SMOKE_URL")
	if databaseURL == "" {
		t.Skip("set GOTEST_POSTGRES_SMOKE_URL to run live PostgreSQL migration/persistence smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	assertTableExists(t, ctx, pool, "releases")
	assertTableExists(t, ctx, pool, "reviews")
	assertTableExists(t, ctx, pool, "suites")
	assertMigrationRecorded(t, ctx, pool, "012_releases_reviews_suites.sql")

	releaseStore := release.NewStore()
	releaseStore.EnableDB(pool)
	reviewStore := workflow.NewReviewStore()
	reviewStore.EnableDB(pool)
	suiteStore := workflow.NewSuiteStore()
	suiteStore.EnableDB(pool)

	releaseID := "smoke-release-" + time.Now().UTC().Format("20060102150405.000000000")
	reviewID := "smoke-review-" + time.Now().UTC().Format("20060102150405.000000000")
	suiteID := "smoke-suite-" + time.Now().UTC().Format("20060102150405.000000000")

	createdRelease := releaseStore.Create(&release.Release{
		ID:        releaseID,
		Name:      "Smoke Release",
		Version:   "v1.2.3",
		ProjectID: "project-smoke",
		RunIDs:    []string{"run-1", "run-2"},
	})
	if !releaseStore.Update(createdRelease.ID, func(rel *release.Release) {
		rel.Status = "completed"
		rel.RunIDs = append(rel.RunIDs, "run-3")
	}) {
		t.Fatalf("update persisted release")
	}

	createdReview := reviewStore.Create(&workflow.Review{
		ID:    reviewID,
		RunID: "run-1",
		Type:  "test_plan",
	})
	if !reviewStore.Approve(createdReview.ID, "smoke-reviewer", "approved in smoke test") {
		t.Fatalf("approve persisted review")
	}

	createdSuite := suiteStore.Create(&workflow.Suite{
		ID:          suiteID,
		Name:        "Smoke Suite",
		ProjectID:   "project-smoke",
		Environment: "staging",
		Tags:        []string{"smoke", "regression"},
		Pinned:      true,
		RunIDs:      []string{"run-1"},
	})

	// Use fresh stores to prove data is read from PostgreSQL rather than the original memory maps.
	freshReleaseStore := release.NewStore()
	freshReleaseStore.EnableDB(pool)
	persistedRelease, ok := freshReleaseStore.Get(createdRelease.ID)
	if !ok {
		t.Fatalf("get persisted release")
	}
	if persistedRelease.Status != "completed" || persistedRelease.Version != "v1.2.3" || !reflect.DeepEqual(persistedRelease.RunIDs, []string{"run-1", "run-2", "run-3"}) {
		t.Fatalf("unexpected persisted release: %+v", persistedRelease)
	}

	freshReviewStore := workflow.NewReviewStore()
	freshReviewStore.EnableDB(pool)
	persistedReview, ok := freshReviewStore.Get(createdReview.ID)
	if !ok {
		t.Fatalf("get persisted review")
	}
	if persistedReview.Status != workflow.Approved || persistedReview.Reviewer != "smoke-reviewer" || persistedReview.Comment != "approved in smoke test" {
		t.Fatalf("unexpected persisted review: %+v", persistedReview)
	}
	reviewsByRun := freshReviewStore.ByRun("run-1")
	if len(reviewsByRun) == 0 {
		t.Fatalf("expected persisted review to be listed by run")
	}

	freshSuiteStore := workflow.NewSuiteStore()
	freshSuiteStore.EnableDB(pool)
	persistedSuite, ok := freshSuiteStore.Get(createdSuite.ID)
	if !ok {
		t.Fatalf("get persisted suite")
	}
	if persistedSuite.Name != "Smoke Suite" || !persistedSuite.Pinned || !reflect.DeepEqual(persistedSuite.Tags, []string{"smoke", "regression"}) || !reflect.DeepEqual(persistedSuite.RunIDs, []string{"run-1"}) {
		t.Fatalf("unexpected persisted suite: %+v", persistedSuite)
	}
	if got := freshSuiteStore.ByTag("smoke"); len(got) == 0 {
		t.Fatalf("expected persisted suite to be listed by tag")
	}
	if !freshSuiteStore.Delete(createdSuite.ID) {
		t.Fatalf("delete persisted suite")
	}
	if _, ok := freshSuiteStore.Get(createdSuite.ID); ok {
		t.Fatalf("suite should be deleted from PostgreSQL")
	}
}

func assertTableExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	)`, table).Scan(&exists); err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	if !exists {
		t.Fatalf("expected table %s to exist", table)
	}
}

func assertMigrationRecorded(t *testing.T, ctx context.Context, pool *pgxpool.Pool, version string) {
	t.Helper()
	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists); err != nil {
		t.Fatalf("check migration %s: %v", version, err)
	}
	if !exists {
		t.Fatalf("expected migration %s to be recorded", version)
	}
}
