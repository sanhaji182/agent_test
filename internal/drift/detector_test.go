package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAddListUpdate(t *testing.T) {
	s := NewStore()
	d := s.Add(Drift{Repository: "acme/app", Type: TypeMissingTest, FilePath: "main.go", Severity: SeverityHigh})
	if d.ID == "" {
		t.Fatal("expected generated ID")
	}
	if d.Status != StatusPending {
		t.Fatalf("expected default status pending, got %s", d.Status)
	}

	if got := s.List("", "", ""); len(got) != 1 {
		t.Fatalf("expected 1 drift, got %d", len(got))
	}
	if got := s.List("other/repo", "", ""); len(got) != 0 {
		t.Fatalf("expected 0 drifts for other repo, got %d", len(got))
	}
	if got := s.List("acme/app", TypeMissingTest, StatusPending); len(got) != 1 {
		t.Fatalf("expected 1 filtered drift, got %d", len(got))
	}

	updated, err := s.UpdateStatus(d.ID, StatusFixed)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
	if updated.Status != StatusFixed {
		t.Fatalf("expected fixed, got %s", updated.Status)
	}
	if _, err := s.UpdateStatus(d.ID, "bogus"); err == nil {
		t.Fatal("expected error for invalid status")
	}
	if _, err := s.UpdateStatus("nope", StatusFixed); err == nil {
		t.Fatal("expected error for unknown id")
	}
}

func TestIsTestFile(t *testing.T) {
	cases := map[string]bool{
		"internal/api/server_test.go":    true,
		"internal/api/server.go":         false,
		"src/components/Button.test.tsx": true,
		"src/__tests__/Button.spec.js":   true,
		"src/components/Button.tsx":      false,
		"app/test_views.py":              true,
		"app/views_test.py":              true,
		"app/views.py":                   false,
		"tests/Unit/UserTest.php":        true,
		"app/Models/User.php":            false,
	}
	for path, want := range cases {
		if got := IsTestFile(path); got != want {
			t.Errorf("IsTestFile(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestDetectDriftMissingTest(t *testing.T) {
	store := NewStore()
	det := NewDetector(store)

	drifts := det.DetectDrift("acme/app", "", nil, []string{"internal/api/server.go"}, nil)
	if len(drifts) != 1 {
		t.Fatalf("expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Type != TypeMissingTest || drifts[0].Severity != SeverityHigh {
		t.Errorf("expected missing_test/high, got %s/%s", drifts[0].Type, drifts[0].Severity)
	}
}

func TestDetectDriftTestChangedTogether(t *testing.T) {
	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", "",
		nil, []string{"internal/api/server.go", "internal/api/server_test.go"}, nil)
	if len(drifts) != 0 {
		t.Fatalf("expected 0 drifts when test changed together, got %d", len(drifts))
	}
}

func TestDetectDriftOutdatedTest(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoDir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "pkg", "calc_test.go"), []byte("package pkg"), 0o644); err != nil {
		t.Fatal(err)
	}

	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", repoDir, nil, []string{"pkg/calc.go"}, nil)
	if len(drifts) != 1 {
		t.Fatalf("expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Type != TypeOutdatedTest || drifts[0].Severity != SeverityMedium {
		t.Errorf("expected outdated_test/medium, got %s/%s", drifts[0].Type, drifts[0].Severity)
	}
}

func TestDetectDriftRemovedTest(t *testing.T) {
	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", "", nil, nil, []string{"pkg/calc_test.go"})
	if len(drifts) != 1 {
		t.Fatalf("expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Type != TypeRemovedTest || drifts[0].Severity != SeverityHigh {
		t.Errorf("expected removed_test/high, got %s/%s", drifts[0].Type, drifts[0].Severity)
	}
}

func TestDetectDriftRemovedSourceAndTestTogether(t *testing.T) {
	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", "", nil, nil,
		[]string{"pkg/calc.go", "pkg/calc_test.go"})
	if len(drifts) != 0 {
		t.Fatalf("expected 0 drifts when source and test removed together, got %d", len(drifts))
	}
}

func TestDetectDriftOrphanedTest(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoDir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "pkg", "calc_test.go"), []byte("package pkg"), 0o644); err != nil {
		t.Fatal(err)
	}

	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", repoDir, nil, nil, []string{"pkg/calc.go"})
	if len(drifts) != 1 {
		t.Fatalf("expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Type != TypeRemovedTest || drifts[0].Severity != SeverityLow {
		t.Errorf("expected removed_test/low, got %s/%s", drifts[0].Type, drifts[0].Severity)
	}
}

func TestDetectDriftDedupsPending(t *testing.T) {
	store := NewStore()
	det := NewDetector(store)

	first := det.DetectDrift("acme/app", "", nil, []string{"internal/api/server.go"}, nil)
	if len(first) != 1 {
		t.Fatalf("expected 1 drift on first push, got %d", len(first))
	}
	second := det.DetectDrift("acme/app", "", nil, []string{"internal/api/server.go"}, nil)
	if len(second) != 0 {
		t.Fatalf("expected 0 new drifts on repeated push, got %d", len(second))
	}
	if got := store.List("acme/app", "", ""); len(got) != 1 {
		t.Fatalf("expected 1 stored drift, got %d", len(got))
	}

	if _, err := store.UpdateStatus(first[0].ID, StatusFixed); err != nil {
		t.Fatal(err)
	}
	third := det.DetectDrift("acme/app", "", nil, []string{"internal/api/server.go"}, nil)
	if len(third) != 1 {
		t.Fatalf("expected re-detection after fix, got %d", len(third))
	}
}

func TestDetectDriftIgnoresNonSourceAndTraversal(t *testing.T) {
	det := NewDetector(NewStore())
	drifts := det.DetectDrift("acme/app", "/tmp/nonexistent",
		nil, []string{"README.md", "../../etc/passwd"}, nil)
	if len(drifts) != 0 {
		t.Fatalf("expected 0 drifts for non-source files, got %d", len(drifts))
	}
}
