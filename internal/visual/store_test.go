package visual_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/visual"
)

func TestAdd_SameURL(t *testing.T) {
	s := visual.NewStore()
	a := s.Add(visual.Artifact{
		RunID:       "run-1",
		StepName:    "login",
		BaselineURL: "/img/a.png",
		CurrentURL:  "/img/a.png",
	})
	if a.SimilarityScore != 1.0 {
		t.Fatalf("expected 1.0 for identical URLs, got %f", a.SimilarityScore)
	}
	if !a.Passed {
		t.Fatal("expected passed=true")
	}
}

func TestAdd_DifferentURL(t *testing.T) {
	s := visual.NewStore()
	a := s.Add(visual.Artifact{
		RunID:       "run-1",
		StepName:    "checkout",
		BaselineURL: "/img/baseline.png",
		CurrentURL:  "/img/current.png",
	})
	if a.SimilarityScore < 0.7 || a.SimilarityScore >= 1.0 {
		t.Fatalf("expected score in [0.7, 1.0), got %f", a.SimilarityScore)
	}
}

func TestByRun(t *testing.T) {
	s := visual.NewStore()
	s.Add(visual.Artifact{RunID: "run-1", StepName: "a", BaselineURL: "/a", CurrentURL: "/b"})
	s.Add(visual.Artifact{RunID: "run-1", StepName: "b", BaselineURL: "/c", CurrentURL: "/d"})
	s.Add(visual.Artifact{RunID: "run-2", StepName: "x", BaselineURL: "/e", CurrentURL: "/f"})

	arts := s.ByRun("run-1")
	if len(arts) != 2 {
		t.Fatalf("expected 2 artifacts for run-1, got %d", len(arts))
	}
}
