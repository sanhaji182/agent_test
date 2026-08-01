package intelligence

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestAnalyzeTestQuality_FlakyTest(t *testing.T) {
	runs := []*agent.TestRun{
		{
			ID:        "r1",
			State:     agent.StateDone,
			TestFiles: []agent.TestFile{{Name: "login test", Content: "test()"}},
			RunResult: &agent.RunResult{
				DurationMs: 5000,
				Failures: []agent.Failure{
					{Test: "login test", Message: "timeout"},
				},
			},
		},
		{
			ID:        "r2",
			State:     agent.StateDone,
			TestFiles: []agent.TestFile{{Name: "login test", Content: "test()"}},
			RunResult: &agent.RunResult{DurationMs: 3000},
		},
		{
			ID:        "r3",
			State:     agent.StateDone,
			TestFiles: []agent.TestFile{{Name: "login test", Content: "test()"}},
			RunResult: &agent.RunResult{DurationMs: 4000},
		},
	}

	qualities := AnalyzeTestQuality(runs)
	if len(qualities) == 0 {
		t.Fatal("expected at least one test quality result")
	}

	login := findQuality("login test", qualities)
	if login == nil {
		t.Fatal("expected login test in results")
	}
	if login.Reliability > 0.8 {
		t.Errorf("expected reliability < 0.8 for flaky test, got %.2f", login.Reliability)
	}
	if len(login.Issues) == 0 {
		t.Error("expected issues for flaky test")
	}
}

func TestAnalyzeTestQuality_AllPassing(t *testing.T) {
	runs := []*agent.TestRun{
		{
			ID:        "r1",
			State:     agent.StateDone,
			TestFiles: []agent.TestFile{{Name: "checkout test", Content: "test()"}},
			RunResult: &agent.RunResult{DurationMs: 2000},
		},
	}

	qualities := AnalyzeTestQuality(runs)
	checkout := findQuality("checkout test", qualities)
	if checkout == nil {
		t.Fatal("expected checkout test in results")
	}
	if checkout.QualityScore < 0.8 {
		t.Errorf("expected high quality for all-passing test, got %.2f", checkout.QualityScore)
	}
}

func TestAnalyzeTestQuality_SlowTest(t *testing.T) {
	runs := []*agent.TestRun{
		{
			ID:        "r1",
			State:     agent.StateDone,
			TestFiles: []agent.TestFile{{Name: "slow test", Content: "test()"}},
			RunResult: &agent.RunResult{DurationMs: 65000},
		},
	}

	qualities := AnalyzeTestQuality(runs)
	slow := findQuality("slow test", qualities)
	if slow == nil {
		t.Fatal("expected slow test in results")
	}
	if slow.Performance >= 0.5 {
		t.Errorf("expected low performance for slow test, got %.2f", slow.Performance)
	}
}

func TestDetectRedundancy(t *testing.T) {
	tests := []TestQuality{
		{Name: "test login with valid credentials", QualityScore: 0.9},
		{Name: "test login with valid credentials edge case", QualityScore: 0.85},
		{Name: "test logout", QualityScore: 0.9},
		{Name: "test checkout flow", QualityScore: 0.8},
		{Name: "test checkout flow with discount", QualityScore: 0.75},
	}

	groups := DetectRedundancy(tests)
	if len(groups) < 2 {
		t.Errorf("expected at least 2 redundancy groups, got %d", len(groups))
	}
}

func TestDetectRedundancy_SingleTest(t *testing.T) {
	tests := []TestQuality{
		{Name: "test unique feature", QualityScore: 0.9},
		{Name: "test another unique", QualityScore: 0.8},
	}

	groups := DetectRedundancy(tests)
	if len(groups) != 0 {
		t.Errorf("expected 0 redundancy groups for unique tests, got %d", len(groups))
	}
}

func TestStringSimilarity(t *testing.T) {
	sim := stringSimilarity("test login with valid credentials", "test login with invalid credentials")
	if sim < 0.5 {
		t.Errorf("expected high similarity, got %.2f", sim)
	}

	sim2 := stringSimilarity("test login", "test logout")
	if sim2 > 0.8 {
		t.Errorf("expected lower similarity for different actions, got %.2f", sim2)
	}
}

func findQuality(name string, qualities []TestQuality) *TestQuality {
	for i := range qualities {
		if qualities[i].Name == name {
			return &qualities[i]
		}
	}
	return nil
}
