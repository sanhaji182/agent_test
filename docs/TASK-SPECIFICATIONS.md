# Task Specifications for AI Agent Autonomous Execution

**Purpose:** Detailed specifications for each task so AI agents can execute autonomously without human intervention.

**Format:** Each task includes success criteria, dependencies, file structure, and acceptance tests.

---

## Phase 1, Sprint 1-2: Foundation

### Task 1.1: Install Tree-sitter Dependencies

**Description:** Install all tree-sitter dependencies required for multi-language parsing.

**Success Criteria:**
- All tree-sitter packages installed in go.mod
- `go build ./...` succeeds
- `go test ./internal/parser/... -v` runs (may fail on parser logic, but compiles)

**Dependencies:** None

**Commands:**
```bash
go get github.com/smacker/go-tree-sitter@latest
go get github.com/smacker/go-tree-sitter/javascript@latest
go get github.com/smacker/go-tree-sitter/golang@latest
go get github.com/smacker/go-tree-sitter/php@latest
go get github.com/smacker/go-tree-sitter/python@latest
go mod tidy
```

**Acceptance Tests:**
```bash
go build ./...
# Expected: No errors

go list -m all | grep tree-sitter
# Expected: Shows all 5 tree-sitter packages
```

**Rollback:** If installation fails, revert go.mod: `git checkout go.mod go.sum`

---

### Task 1.2: Run All Parser Tests

**Description:** Execute all parser tests to validate implementations.

**Success Criteria:**
- All tests in `internal/parser/*` pass
- Test coverage > 80% for each parser
- No compilation errors

**Dependencies:** Task 1.1 (tree-sitter dependencies)

**Commands:**
```bash
go test ./internal/parser/... -v -cover
```

**Acceptance Tests:**
- `PASS` for all test files
- Coverage report shows > 80% for:
  - `internal/parser/javascript`
  - `internal/parser/go`
  - `internal/parser/php`
  - `internal/parser/python`

**Expected Output:**
```
ok  	github.com/go-go-golems/gotest-agent/internal/parser/javascript	coverage: 85.2%
ok  	github.com/go-go-golems/gotest-agent/internal/parser/go	coverage: 82.7%
ok  	github.com/go-go-golems/gotest-agent/internal/parser/php	coverage: 88.1%
ok  	github.com/go-go-golems/gotest-agent/internal/parser/python	coverage: 84.5%
```

**Failure Handling:**
- If tests fail, read error messages and fix parser logic
- Common issues:
  - Tree-sitter API changes → update parser code to match new API
  - Missing test cases → add test cases to improve coverage
  - Parsing logic errors → debug and fix extraction logic

---

### Task 1.3: Fix Parser Test Failures

**Description:** If tests fail, debug and fix parser implementations.

**Success Criteria:**
- All tests pass after fixes
- Coverage maintained or improved

**Dependencies:** Task 1.2 (test failures identified)

**Process:**
1. Read test failure output
2. Identify root cause (parsing logic, tree-sitter API, test data)
3. Fix parser implementation
4. Re-run tests: `go test ./internal/parser/... -v`
5. Iterate until all tests pass

**Common Fixes:**
- **Tree-sitter API mismatch:** Check `github.com/smacker/go-tree-sitter` docs, update method calls
- **Missing node types:** Add handling for additional AST node types
- **Test data issues:** Update test fixtures to match real-world code patterns

---

## Phase 1, Sprint 8: AI Synthesis Layer

### Task 8.1: Design AI Prompt for Codebase Analysis

**Description:** Create prompt template for LLM to analyze parsed codebase and generate test plans.

**Success Criteria:**
- Prompt template exists in `internal/ai/prompts/codebase_analysis.txt`
- Prompt includes all relevant context (routes, models, handlers, framework)
- Prompt requests structured output (JSON format)

**Dependencies:** Task 1.2 (parser tests pass)

**File Structure:**
```
internal/ai/
├── prompts/
│   └── codebase_analysis.txt
└── synthesis.go
```

**Prompt Template:**
```
You are an expert test automation engineer. Analyze this codebase and generate a comprehensive test plan.

CODEBASE INFORMATION:
Language: {language}
Framework: {framework}

ROUTES ({route_count}):
{routes_json}

MODELS ({model_count}):
{models_json}

HANDLERS ({handler_count}):
{handlers_json}

Generate a test plan that covers:
1. Happy path scenarios for each route
2. Edge cases (invalid input, missing data, authorization errors)
3. CRUD operations for each model
4. Integration tests for complex workflows

Output format (JSON):
{
  "test_plan": [
    {
      "name": "Test name",
      "type": "unit|integration|e2e",
      "description": "What this test verifies",
      "priority": "high|medium|low",
      "estimated_time": "5m"
    }
  ]
}
```

**Acceptance Tests:**
- Prompt file exists and is well-structured
- Prompt can be rendered with real codebase data
- Output format is parseable JSON

---

### Task 8.2: Implement Codebase Synthesis Service

**Description:** Create service that combines parser output and calls LLM to generate test plans.

**Success Criteria:**
- Service exists in `internal/ai/synthesis.go`
- Service can generate test plan from parsed codebase
- Service handles LLM API errors gracefully

**Dependencies:** Task 8.1 (prompt template)

**File Structure:**
```
internal/ai/
├── synthesis.go
└── synthesis_test.go
```

**Implementation:**
```go
package ai

import (
    "context"
    "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

type SynthesisService struct {
    client *Client
}

func NewSynthesisService(client *Client) *SynthesisService {
    return &SynthesisService{client: client}
}

func (s *SynthesisService) GenerateTestPlan(ctx context.Context, codebase *types.Codebase) (*TestPlan, error) {
    // 1. Load prompt template
    // 2. Render prompt with codebase data
    // 3. Call LLM API
    // 4. Parse JSON response
    // 5. Return structured test plan
}

type TestPlan struct {
    Tests []TestCase `json:"test_plan"`
}

type TestCase struct {
    Name          string `json:"name"`
    Type          string `json:"type"`
    Description   string `json:"description"`
    Priority      string `json:"priority"`
    EstimatedTime string `json:"estimated_time"`
}
```

**Acceptance Tests:**
```go
func TestSynthesisService_GenerateTestPlan(t *testing.T) {
    // Mock LLM client
    // Create sample codebase
    // Call GenerateTestPlan
    // Verify output structure
    // Verify all routes have corresponding tests
}
```

---

### Task 8.3: Implement Test Plan to Playwright Code Generator

**Description:** Convert test plan (JSON) to executable Playwright test scripts.

**Success Criteria:**
- Generator exists in `internal/ai/playwright_generator.go`
- Generator produces valid Playwright TypeScript code
- Generated code can be executed by Playwright

**Dependencies:** Task 8.2 (synthesis service)

**File Structure:**
```
internal/ai/
├── playwright_generator.go
└── playwright_generator_test.go
```

**Implementation:**
```go
package ai

type PlaywrightGenerator struct{}

func (g *PlaywrightGenerator) GenerateCode(testPlan *TestPlan) (string, error) {
    // 1. Generate TypeScript imports
    // 2. For each test case:
    //    - Generate test function
    //    - Add setup/teardown if needed
    //    - Add assertions based on test type
    // 3. Return complete TypeScript code
}
```

**Generated Code Example:**
```typescript
import { test, expect } from '@playwright/test';

test('User can create new post', async ({ page }) => {
  // Navigate to posts page
  await page.goto('/posts');
  
  // Click create button
  await page.click('button:has-text("Create")');
  
  // Fill form
  await page.fill('input[name="title"]', 'Test Post');
  await page.fill('textarea[name="content"]', 'Test content');
  
  // Submit form
  await page.click('button[type="submit"]');
  
  // Verify post created
  await expect(page.locator('.post-title')).toHaveText('Test Post');
});
```

**Acceptance Tests:**
- Generator produces valid TypeScript syntax
- Generated code includes proper imports
- Generated code follows Playwright best practices
- Code can be linted without errors: `npx eslint generated-test.ts`

---

### Task 8.4: Implement Confidence Scoring

**Description:** Add confidence scoring to generated test plans to help users prioritize review.

**Success Criteria:**
- Each test case has confidence score (0-100)
- Scores based on:
  - Route complexity (number of parameters, middleware)
  - Model complexity (number of fields, relationships)
  - Handler complexity (lines of code, external calls)
- Low-confidence tests flagged for manual review

**Dependencies:** Task 8.2 (synthesis service)

**Implementation:**
```go
type ConfidenceScorer struct{}

func (s *ConfidenceScorer) ScoreTestCase(test *TestCase, codebase *types.Codebase) int {
    score := 100
    
    // Reduce score for complex routes
    route := findRoute(test.Name, codebase.Routes)
    if route != nil {
        if len(route.Middleware) > 3 {
            score -= 10
        }
        if countRouteParams(route.Path) > 2 {
            score -= 15
        }
    }
    
    // Reduce score for complex models
    model := findModel(test.Name, codebase.Models)
    if model != nil {
        if len(model.Fields) > 10 {
            score -= 10
        }
        if len(model.Relationships) > 3 {
            score -= 15
        }
    }
    
    return max(0, score)
}
```

**Acceptance Tests:**
```go
func TestConfidenceScorer_ComplexRoute(t *testing.T) {
    // Create test with complex route (5 middleware, 3 params)
    // Verify score < 80
}

func TestConfidenceScorer_SimpleRoute(t *testing.T) {
    // Create test with simple route (no middleware, 0 params)
    // Verify score = 100
}
```

---

## Code Patterns & Conventions

### Error Handling Pattern

**Standard:**
```go
func (s *Service) DoSomething(ctx context.Context) error {
    result, err := s.dependency.Call()
    if err != nil {
        return fmt.Errorf("failed to call dependency: %w", err)
    }
    
    // Process result
    return nil
}
```

**With Context:**
```go
func (s *Service) DoSomething(ctx context.Context) error {
    ctx, span := tracing.Tracer.Start(ctx, "Service.DoSomething")
    defer span.End()
    
    result, err := s.dependency.Call()
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("failed to call dependency: %w", err)
    }
    
    return nil
}
```

### Testing Pattern

**Unit Test:**
```go
func TestService_DoSomething(t *testing.T) {
    // Arrange
    mockDep := &MockDependency{}
    mockDep.On("Call").Return("result", nil)
    
    service := NewService(mockDep)
    
    // Act
    err := service.DoSomething(context.Background())
    
    // Assert
    assert.NoError(t, err)
    mockDep.AssertExpectations(t)
}
```

**Table-Driven Test:**
```go
func TestParser_ExtractRoutes(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected []Route
    }{
        {
            name: "simple route",
            code: `app.get('/users', handler)`,
            expected: []Route{
                {Method: "GET", Path: "/users", Handler: "handler"},
            },
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            routes := parser.ExtractRoutes(tt.code)
            assert.Equal(t, tt.expected, routes)
        })
    }
}
```

### Naming Conventions

- **Packages:** lowercase, single word (e.g., `parser`, `ai`, `api`)
- **Files:** lowercase with underscores (e.g., `synthesis.go`, `playwright_generator.go`)
- **Functions:** PascalCase for exported, camelCase for private
- **Constants:** PascalCase (e.g., `MaxRetries`, `DefaultTimeout`)
- **Variables:** camelCase (e.g., `routeCount`, `testPlan`)
- **Tests:** `TestTypeName_MethodName` (e.g., `TestSynthesisService_GenerateTestPlan`)

---

## Next Steps

After completing these tasks:

1. **Sprint 9-10: GitHub Integration** - OAuth, clone repos, webhooks
2. **Sprint 11-12: UI/UX** - Dashboard, results view, test execution
3. **Sprint 13-14: Testing & Beta Prep** - E2E tests, performance, docs
4. **Sprint 15-16: Beta Launch** - 50 beta users, feedback, iteration

---

**Last Updated:** 2026-07-30
**Maintained By:** AI Agent
