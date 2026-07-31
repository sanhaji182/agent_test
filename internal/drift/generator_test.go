package drift

import (
	"context"
	"errors"
	"testing"
)

type mockLLM struct {
	out string
	err error
}

func (m *mockLLM) GenerateText(ctx context.Context, prompt string) (string, error) {
	return m.out, m.err
}
func (m *mockLLM) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	return m.out, m.err
}

func TestGenerateForDriftSuccess(t *testing.T) {
	llm := &mockLLM{out: "Here is the test:\n```go\npackage pkg\n\nfunc TestFoo(t *testing.T) {}\n```\nConfidence: 87"}
	store := NewGeneratedTestStore()
	gen := NewGenerator(llm, store)

	d := Drift{ID: "drift-1", Type: TypeMissingTest, FilePath: "pkg/foo.go", Repository: "acme/app"}
	gt, err := gen.GenerateForDrift(context.Background(), d)
	if err != nil {
		t.Fatalf("GenerateForDrift: %v", err)
	}
	if gt.DriftID != "drift-1" {
		t.Errorf("expected drift-1, got %s", gt.DriftID)
	}
	if gt.ConfidenceScore != 87 {
		t.Errorf("expected confidence 87, got %d", gt.ConfidenceScore)
	}
	if gt.TestCode != "package pkg\n\nfunc TestFoo(t *testing.T) {}" {
		t.Errorf("unexpected extracted code: %q", gt.TestCode)
	}
	if gt.Status != GenStatusGenerated {
		t.Errorf("expected status generated, got %s", gt.Status)
	}
	if got := store.ByDrift("drift-1"); len(got) != 1 {
		t.Errorf("expected 1 stored test, got %d", len(got))
	}
}

func TestGenerateForDriftRemovedTestRejected(t *testing.T) {
	gen := NewGenerator(&mockLLM{out: "x"}, NewGeneratedTestStore())
	_, err := gen.GenerateForDrift(context.Background(), Drift{ID: "d", Type: TypeRemovedTest, FilePath: "a_test.go"})
	if err == nil {
		t.Fatal("expected error for removed_test drift")
	}
}

func TestGenerateForDriftLLMError(t *testing.T) {
	gen := NewGenerator(&mockLLM{err: errors.New("boom")}, NewGeneratedTestStore())
	_, err := gen.GenerateForDrift(context.Background(), Drift{ID: "d", Type: TypeMissingTest, FilePath: "a.go"})
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

func TestGenerateForDriftEmptyCode(t *testing.T) {
	gen := NewGenerator(&mockLLM{out: "   "}, NewGeneratedTestStore())
	_, err := gen.GenerateForDrift(context.Background(), Drift{ID: "d", Type: TypeMissingTest, FilePath: "a.go"})
	if err == nil {
		t.Fatal("expected error for empty test code")
	}
}

func TestGenerateForDriftNilLLM(t *testing.T) {
	gen := NewGenerator(nil, NewGeneratedTestStore())
	_, err := gen.GenerateForDrift(context.Background(), Drift{ID: "d", Type: TypeMissingTest, FilePath: "a.go"})
	if err == nil {
		t.Fatal("expected error when no LLM configured")
	}
}

func TestGeneratedTestStoreUpdateStatus(t *testing.T) {
	store := NewGeneratedTestStore()
	gt := store.Add(GeneratedTest{DriftID: "d", TestCode: "x"})
	if _, err := store.UpdateStatus(gt.ID, GenStatusReviewed); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if _, err := store.UpdateStatus(gt.ID, "bogus"); err == nil {
		t.Fatal("expected error for invalid status")
	}
	if _, err := store.UpdateStatus("nope", GenStatusReviewed); err == nil {
		t.Fatal("expected error for unknown id")
	}
}

func TestExtractConfidenceClamps(t *testing.T) {
	cases := map[string]int{
		"Confidence: 50": 50,
		"confidence 200": 100,
		"no number here": 0,
		"Confidence: 0":  0,
	}
	for in, want := range cases {
		if got := extractConfidence(in); got != want {
			t.Errorf("extractConfidence(%q) = %d, want %d", in, got, want)
		}
	}
}
