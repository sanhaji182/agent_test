package drift

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Generated-test statuses.
const (
	GenStatusGenerated = "generated"
	GenStatusReviewed  = "reviewed"
	GenStatusRejected  = "rejected"
)

// GeneratedTest is a test synthesized by the LLM to resolve a drift.
type GeneratedTest struct {
	ID              string    `json:"id"`
	DriftID         string    `json:"drift_id"`
	TestName        string    `json:"test_name"`
	TestCode        string    `json:"test_code"`
	ConfidenceScore int       `json:"confidence_score"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GeneratedTestStore is an in-memory store for LLM-generated drift tests.
type GeneratedTestStore struct {
	mu      sync.RWMutex
	items   []GeneratedTest
	counter int64
	dbPool  *pgxpool.Pool
}

func NewGeneratedTestStore() *GeneratedTestStore { return &GeneratedTestStore{} }

// EnableDB enables PostgreSQL persistence for generated drift tests.
func (s *GeneratedTestStore) EnableDB(pool *pgxpool.Pool) { s.dbPool = pool }

func (s *GeneratedTestStore) Add(t GeneratedTest) *GeneratedTest {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.Status == "" {
		t.Status = GenStatusGenerated
	}
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	s.items = append(s.items, t)
	if s.dbPool != nil {
		if err := s.persistGeneratedTestDB(&t); err != nil {
			slog.Warn("generated drift test persistence failed", "generated_test_id", t.ID, "error", err)
		}
	}
	return &t
}

// ByDrift returns generated tests for a drift, newest first.
func (s *GeneratedTestStore) ByDrift(driftID string) []GeneratedTest {
	if s.dbPool != nil {
		if list, err := s.generatedTestsByDriftDB(driftID); err == nil {
			return list
		} else {
			slog.Warn("generated drift tests DB list failed, falling back to memory", "drift_id", driftID, "error", err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []GeneratedTest{}
	for i := len(s.items) - 1; i >= 0; i-- {
		if s.items[i].DriftID == driftID {
			result = append(result, s.items[i])
		}
	}
	return result
}

// UpdateStatus transitions a generated test to generated, reviewed, or rejected.
func (s *GeneratedTestStore) UpdateStatus(id, status string) (*GeneratedTest, error) {
	switch status {
	case GenStatusGenerated, GenStatusReviewed, GenStatusRejected:
	default:
		return nil, fmt.Errorf("invalid status: %q", status)
	}
	if s.dbPool != nil {
		if gt, err := s.getGeneratedTestDB(id); err == nil {
			gt.Status = status
			gt.UpdatedAt = time.Now()
			if err := s.persistGeneratedTestDB(gt); err != nil {
				slog.Warn("generated drift test status persistence failed", "generated_test_id", id, "error", err)
			}
			return gt, nil
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Status = status
			s.items[i].UpdatedAt = time.Now()
			t := s.items[i]
			return &t, nil
		}
	}
	return nil, fmt.Errorf("generated test not found: %s", id)
}

// Generator synthesizes tests for drifts via an LLM.
type Generator struct {
	llm   ai.Client
	store *GeneratedTestStore
}

func NewGenerator(llm ai.Client, store *GeneratedTestStore) *Generator {
	return &Generator{llm: llm, store: store}
}

// GenerateForDrift asks the LLM to produce a test that resolves the drift and
// stores it. removed_test drifts are not auto-generatable (the source is gone).
func (g *Generator) GenerateForDrift(ctx context.Context, d Drift) (*GeneratedTest, error) {
	if g.llm == nil {
		return nil, fmt.Errorf("no LLM client configured")
	}
	if d.Type == TypeRemovedTest {
		return nil, fmt.Errorf("cannot auto-generate a test for a removed_test drift")
	}

	prompt := buildGenerationPrompt(d)
	out, err := g.llm.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm generation failed: %w", err)
	}

	code := extractCodeBlock(out)
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("llm returned empty test code")
	}

	gt := g.store.Add(GeneratedTest{
		DriftID:         d.ID,
		TestName:        testNameFor(d.FilePath),
		TestCode:        code,
		ConfidenceScore: extractConfidence(out),
	})
	return gt, nil
}

func buildGenerationPrompt(d Drift) string {
	var b strings.Builder
	b.WriteString("You are a senior test engineer. A code/test drift was detected.\n\n")
	fmt.Fprintf(&b, "Drift type: %s\n", d.Type)
	fmt.Fprintf(&b, "Source file: %s\n", d.FilePath)
	fmt.Fprintf(&b, "Repository: %s\n", d.Repository)
	fmt.Fprintf(&b, "Description: %s\n\n", d.Description)
	b.WriteString("Write a single, self-contained automated test file that covers the source file above, ")
	b.WriteString("using the idiomatic test framework for its language (Go testing, Jest/Vitest for JS/TS, pytest for Python, PHPUnit for PHP).\n")
	b.WriteString("Return ONLY the test code inside one fenced code block, then on a final line output 'Confidence: N' where N is 0-100.\n")
	return b.String()
}

var codeBlockRe = regexp.MustCompile("(?s)```[a-zA-Z0-9_+-]*\\n(.*?)```")

// extractCodeBlock returns the first fenced code block, or the trimmed input
// if no fence is present.
func extractCodeBlock(s string) string {
	if m := codeBlockRe.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(s)
}

var confidenceRe = regexp.MustCompile(`(?i)confidence[:\s]+(\d{1,3})`)

func extractConfidence(s string) int {
	m := confidenceRe.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}

func testNameFor(filePath string) string {
	return "test for " + filePath
}
