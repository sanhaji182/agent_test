# GoTest Agent — Master Build Blueprint
> Version: 1.0 | Author: AI Agent Planning Document
> Target: Solo Developer + AI Agent (Cursor/Claude Code)
> Stack: Go + Steel Browser (self-host) + LangGraph Sidecar + Vision Model + Braintrust Evals

---

# SECTION 1 — VISION & POSITIONING

## 1.1 Problem Statement

Software testing is the most neglected part of modern development, especially for:
- Solo developers and small teams building with AI-generated code
- Teams that don't have dedicated QA engineers
- Developers in emerging markets who can't afford expensive testing SaaS ($99+/month)
- PHP/CI4 + Laravel projects that have ZERO AI testing agent support today

## 1.2 Solution

GoTest Agent is a self-hostable AI testing agent that:
1. Reads your codebase and understands what to test
2. Generates a comprehensive test plan via LLM
3. Writes and executes Playwright + PHPUnit tests automatically
4. Auto-fixes failing tests in an agentic loop (max 3 attempts)
5. Reports results with actionable insights
6. Integrates directly into your IDE via MCP (Cursor, VS Code, Windsurf)

## 1.3 Ideal Customer Profile (ICP)

Primary:
- Solo developers / small teams (1-5 people) building web apps
- Using Next.js, CodeIgniter 4, Laravel, Express, Hono
- Already using AI coding tools (Cursor, Claude Code)
- Budget-conscious: paying TestSprite $49-99/month feels wasteful

Secondary:
- Indonesian/Asian dev teams building SaaS products
- Agencies that need to test client projects repeatedly
- Developers who build with "vibe coding" and need confidence in output

## 1.4 Competitive Positioning

| Product     | Price     | Self-host | PHP support | Open Source |
|-------------|-----------|-----------|-------------|-------------|
| TestSprite  | $49+/mo   | ❌        | ❌          | ❌          |
| Qodo        | $19+/mo   | ❌        | Partial     | ❌          |
| GoTest Agent| $0 self-host / $19+/mo cloud | ✅ | ✅ | ✅ |

## 1.5 Pricing Hypothesis

- Tier 0: Self-host (free, open source, community)
- Tier 1: Cloud Starter — $19/month — 500 test runs/month, 1 project
- Tier 2: Cloud Pro — $49/month — unlimited runs, 5 projects, priority queue
- Tier 3: Team — $99/month — unlimited, team seats, GitHub integration, SLA

---

# SECTION 2 — FUNCTIONAL REQUIREMENTS

## 2.1 Core Features (MVP - Phase 1-2)

FR-001: Codebase Analysis
- Agent must detect language (PHP, JS/TS, Go, Python)
- Agent must detect framework (CI4, Laravel, Next.js, Express, Hono, Chi)
- Agent must extract: routes, controllers, models, API endpoints
- Agent must build a project summary for LLM context

FR-002: Test Plan Generation
- Agent generates test plan from codebase analysis + user requirements
- Test plan includes: scenarios, priority (high/medium/low), steps
- Test plan is persisted to DB and versioned
- User can view/edit test plan before execution

FR-003: Test Script Generation
- Agent generates Playwright TypeScript test files per scenario
- Files follow page object model pattern
- For PHP projects: generate PHPUnit test files alongside Playwright
- Generated files are persisted and downloadable

FR-004: Test Execution
- Tests run in Steel Browser sandbox (self-hosted)
- Execution is async (queued via Asynq)
- Realtime progress streamed via SSE (Server-Sent Events)
- Execution timeout: configurable, default 5 minutes per run
- Parallel execution: up to 5 runs simultaneously (configurable)

FR-005: Auto-Fix Loop
- On test failure, agent analyzes error + screenshot
- LLM suggests fixes to test scripts
- Fixed scripts re-executed automatically
- Max fix attempts: 3 (configurable)
- Each attempt logged and stored

FR-006: Result Reporting
- Summary: total/passed/failed/skipped
- Per-test detail: steps executed, error message, screenshot
- LLM-generated human-readable summary
- Export: JSON, HTML report
- GitHub commit status integration (Phase 3)

FR-007: MCP Server (IDE Integration)
- Tools exposed: run_tests, generate_test_plan, analyze_project, get_run_status
- Works with Cursor, VS Code (Copilot), Windsurf
- Single binary, stdio transport
- No auth required for local mode

FR-008: HTTP API
- REST API for all operations
- API key authentication
- Webhook support for CI/CD triggers
- OpenAPI spec auto-generated

## 2.2 Enhanced Features (Phase 3-4)

FR-009: Visual Regression Testing
- Baseline screenshot capture on first run
- Diff comparison on subsequent runs
- GPT-4o Vision judges visual similarity (0-100% score)
- Configurable threshold (default: 95%)
- Highlight changed regions in diff image

FR-010: Self-Healing Tests
- On selector failure, agent captures screenshot
- Vision model identifies the target element visually
- New selector generated based on visual + DOM context
- Test file updated automatically

FR-011: Multi-Agent Loop (LangGraph Sidecar)
- Planner agent: breaks requirements into test scenarios
- Writer agent: generates test scripts per scenario
- Critic agent: reviews generated scripts for quality
- Executor agent: runs tests and captures context
- Fixer agent: analyzes failures and patches scripts

FR-012: Eval Pipeline (Braintrust)
- Score quality of: test plan, generated scripts, fix suggestions
- Track regression across LLM model upgrades
- CI gate: block deploy if eval score drops >5%

## 2.3 Non-Functional Requirements

NFR-001: Performance
- Test plan generation: < 30 seconds
- Test execution start: < 10 seconds after queuing
- API response time: < 200ms for non-LLM endpoints
- Concurrent runs: minimum 5 simultaneous

NFR-002: Reliability
- Uptime target: 99.5% (self-hosted)
- Failed runs auto-retry: 3 times with exponential backoff
- Data persistence: all runs stored indefinitely (user can purge)

NFR-003: Security
- API key auth for all HTTP endpoints
- Steel Browser network isolation per session
- No code execution outside Docker sandbox
- Secrets never logged or stored in plain text

NFR-004: Portability
- Single docker-compose.yml to run everything
- Works on any Linux VPS with 4GB+ RAM
- ARM64 support (for Apple Silicon dev machines)

---

# SECTION 3 — USER FLOWS

## 3.1 Flow A: IDE Usage (MCP)
1. Developer opens Cursor with project folder
2. Types: "Test my project" or drags folder to chat
3. MCP tool `run_tests` invoked automatically
4. Agent analyzes codebase → shows test plan in chat
5. Agent writes tests → executes in Steel Browser sandbox
6. Results stream back to Cursor chat
7. If failures: agent shows what failed + auto-fix attempt
8. Final summary: "X tests passed, Y failed, Z fixed automatically"

## 3.2 Flow B: Dashboard Usage
1. User opens GoTest Agent web dashboard
2. Clicks "New Run" → enters project path or uploads project zip
3. Optionally enters requirements in natural language
4. Clicks "Run" → sees realtime progress via SSE
5. Views results: passed/failed breakdown, screenshots, logs
6. Downloads HTML report
7. Shares report link (public URL, 30-day expiry)

## 3.3 Flow C: CI/CD Integration
1. Developer pushes code to GitHub
2. GitHub webhook triggers POST /api/v1/webhooks/github
3. Agent queues test run automatically
4. On completion, posts GitHub commit status (✅ or ❌)
5. Optional: post summary to Slack/Telegram

## 3.4 Flow D: Self-Healing (Phase 3)
1. Test fails with "Element not found: #submit-btn"
2. Agent captures screenshot of current page state
3. Vision model analyzes screenshot
4. Identifies: "The submit button now has class .btn-submit"
5. Agent patches test file with new selector
6. Re-runs test — passes
7. Updated test file committed to repo (optional)

---

# SECTION 4 — END-TO-END ARCHITECTURE

## 4.1 System Map

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                            │
│  Cursor/VS Code IDE    Next.js Dashboard    GitHub/CI Webhook   │
└────────┬───────────────────────┬───────────────────┬───────────┘
         │ MCP (stdio)           │ HTTP REST          │ Webhook POST
         ▼                       ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GATEWAY LAYER (Go)                         │
│    cmd/mcp — MCP Server      cmd/server — HTTP API (Chi)        │
│    - stdio transport         - /api/v1/*                        │
│    - 4 tools exposed         - API key auth middleware          │
│    - direct agent call       - SSE /api/v1/runs/:id/stream      │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ORCHESTRATION LAYER (Go)                     │
│                   internal/agent — State Machine                │
│                                                                 │
│   SIMPLE RUNS: Go state machine (direct)                        │
│   idle → analyzing → planning → writing → running → fixing → done │
│                                                                 │
│   COMPLEX RUNS: LangGraph Sidecar (Python HTTP service)         │
│   POST http://langgraph-sidecar:8000/agent/run                  │
│   Multi-agent: planner → writer → critic → executor → fixer     │
└──────────┬───────────────────────────────┬──────────────────────┘
           ▼                               ▼
┌─────────────────────┐      ┌───────────────────────────────────┐
│   QUEUE LAYER       │      │         LLM LAYER                 │
│   Asynq + Redis     │      │  Primary: Anthropic Claude        │
│   - test:run jobs   │      │  Fallback: OpenAI GPT-4o          │
│   - 5 workers       │      │  Vision: GPT-4o Vision (Phase 3)  │
│   - retry 3x        │      │  Gateway: LiteLLM (self-hosted)   │
└──────────┬──────────┘      │  Evals: Braintrust SDK            │
           ▼                 └───────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                    EXECUTION LAYER                              │
│                                                                 │
│   Steel Browser (self-hosted)      Docker Playwright Runner     │
│   - Session API: :3000             - Playwright TypeScript      │
│   - UI: :5173                      - Connects to Steel via CDP  │
│   - CDP endpoint per session       - PHPUnit for PHP projects   │
│   - 10 max concurrent sessions     - Isolated per run           │
└──────────────────────────────┬──────────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    PERSISTENCE LAYER                            │
│   PostgreSQL          Redis              Local Filesystem       │
│   - test_runs         - Job queue        - Screenshots          │
│   - test_files        - Session cache    - HTML reports         │
│   - run_results       - Rate limits      - Generated test files │
│   - projects                                                    │
└─────────────────────────────────────────────────────────────────┘
```

## 4.2 Service Ports Reference

| Service             | Port  | Protocol  |
|---------------------|-------|-----------|
| GoTest API          | 8080  | HTTP      |
| MCP Server          | stdio | stdio     |
| Steel Browser API   | 3000  | HTTP/WS   |
| Steel Browser UI    | 5173  | HTTP      |
| LangGraph Sidecar   | 8000  | HTTP      |
| LiteLLM Gateway     | 4000  | HTTP      |
| PostgreSQL          | 5432  | TCP       |
| Redis               | 6379  | TCP       |
| Langfuse            | 3001  | HTTP      |
| Braintrust (local)  | 3002  | HTTP      |

---

# SECTION 5 — DETAILED SERVICE DESIGN

## 5.1 Go Backend Services

### 5.1.1 cmd/server — HTTP API

Routes:
  POST   /api/v1/runs                   → queue test run
  GET    /api/v1/runs                   → list runs (paginated)
  GET    /api/v1/runs/:id               → get run detail
  GET    /api/v1/runs/:id/stream        → SSE realtime progress
  DELETE /api/v1/runs/:id               → cancel run
  GET    /api/v1/runs/:id/report        → HTML report
  POST   /api/v1/projects               → register project
  GET    /api/v1/projects               → list projects
  POST   /api/v1/webhooks/github        → GitHub webhook handler

Middleware stack (in order):
  1. RequestID (inject X-Request-ID)
  2. Logger (structured slog)
  3. Recoverer (panic handler)
  4. CORS
  5. APIKeyAuth (check X-Api-Key header against DB)
  6. RateLimiter (10 req/sec per API key, Redis-backed)

### 5.1.2 cmd/mcp — MCP Server

Transport: stdio (for local IDE use)
Tools:
  run_tests(project_path, requirements?) → starts run, returns run_id + summary
  generate_test_plan(project_path) → returns JSON test plan
  analyze_project(project_path) → returns codebase analysis
  get_run_status(run_id) → returns current state + result

### 5.1.3 internal/agent — State Machine

States:
  idle → analyzing → plan_generated → writing_tests →
  running → fixing (loop max 3) → done | failed

Transitions:
  Each state has:
    - Enter function (what to do)
    - Exit condition (when to move forward)
    - Error handler (what to do on failure)

Context object (TestRun):
  ID, ProjectPath, Requirements
  State, CodeAnalysis, TestPlan
  TestFiles, RunResult, FixAttempts
  Screenshots, VisualDiffs (Phase 3)
  CreatedAt, UpdatedAt, FinishedAt

### 5.1.4 internal/llm — LLM Clients

Interfaces:
  type Analyzer interface { AnalyzeCodebase(ctx, path) (string, error) }
  type Planner interface { GenerateTestPlan(ctx, analysis, req) (*TestPlan, error) }
  type Writer interface { GenerateTestScripts(ctx, plan, analysis) ([]TestFile, error) }
  type Fixer interface { SuggestFixes(ctx, failures, files) ([]TestFile, error) }
  type VisionAnalyzer interface { AnalyzeScreenshot(ctx, imgB64, prompt) (string, error) }

Implementations:
  AnthropicClient — implements all interfaces using anthropic-sdk-go
  OpenAIClient — fallback implementation
  LiteLLMClient — routes through self-hosted LiteLLM gateway

### 5.1.5 internal/runner — Test Executors

Interface:
  type TestRunner interface {
    Run(ctx, testFiles []TestFile, config RunConfig) (*RunResult, error)
  }

RunConfig:
  ProjectURL string    // target URL to test
  BaseURL    string    // e.g., http://localhost:3000
  Timeout    time.Duration
  SteelSessionID string

Implementations:
  SteelPlaywrightRunner — connects Playwright to Steel Browser via CDP
  DockerPlaywrightRunner — fallback: standalone Docker container
  PHPUnitRunner — for PHP projects (runs phpunit in Docker)

### 5.1.6 internal/steel — Steel Browser Client

Responsibilities:
  - Create browser sessions via Steel API
  - Return CDP WebSocket URL for Playwright to connect
  - Capture screenshots via Steel session API
  - Cleanup sessions after run completes
  - Pool management: max 10 concurrent sessions

Key methods:
  CreateSession(ctx) (*Session, error)
  GetCDPURL(sessionID string) string
  Screenshot(ctx, sessionID, fullPage bool) ([]byte, error)
  DestroySession(ctx, sessionID string) error
  ListSessions(ctx) ([]*Session, error)

Steel API base: http://localhost:3000/v1

### 5.1.7 internal/vision — Visual Analysis (Phase 3)

Responsibilities:
  - Take screenshot before/after test step
  - Compare screenshots for visual regression
  - Use GPT-4o Vision to identify UI elements
  - Generate selector suggestions from visual context

Key methods:
  CaptureBaseline(ctx, runID, stepName string) error
  CompareWithBaseline(ctx, runID, stepName string) (*DiffResult, error)
  IdentifyElement(ctx, screenshot []byte, description string) (*ElementHint, error)

DiffResult:
  SimilarityScore float64   // 0.0 - 1.0
  DiffImageB64    string
  ChangedRegions  []Region
  Passed          bool      // score >= threshold

### 5.1.8 internal/evals — Braintrust Integration

Responsibilities:
  - Log LLM inputs/outputs for eval datasets
  - Run automated scorers after each LLM call
  - Track experiment snapshots
  - CI integration: fail if score drops

Scorers to implement:
  TestPlanQualityScorer   → coverage, specificity, actionability (0-100)
  TestScriptValidityScorer → syntactically valid TS, has assertions (pass/fail)
  FixSuccessScorer        → did the fix actually resolve the failure (pass/fail)
  OverallRunScorer        → ratio of issues caught vs issues total (0-100)

### 5.1.9 internal/db — Database Layer

Tables:
  projects      — id, name, path, config, created_at
  test_runs     — id, project_id, state, code_analysis, test_plan,
                  test_files, run_result, fix_attempts, error_msg,
                  screenshots, created_at, updated_at, finished_at
  api_keys      — id, key_hash, name, user_id, created_at, last_used_at
  users         — id, email, password_hash, plan, created_at (Phase 4)

Use pgx/v5 pool for all queries.
Use sqlc to generate type-safe Go code from SQL.
Use golang-migrate for schema versioning.

### 5.1.10 internal/queue — Job Queue

Job types:
  tests:run     → main test execution job
  tests:fix     → fix loop job (spawned by tests:run)
  report:generate → async HTML report generation

Worker config:
  Concurrency: 5
  RetryLimit: 3
  RetryDelays: 10s, 30s, 60s (exponential)
  Timeout per job: 10 minutes

---

# SECTION 6 — AGENT WORKFLOW DESIGN

## 6.1 Simple Agent (Go State Machine)

Used for: standard test runs, MCP tool calls

```
[IDLE]
  ↓ trigger: TestRun created
[ANALYZING]
  - Call llm.AnalyzeCodebase(projectPath)
  - Detect language + framework
  - Extract routes/endpoints/controllers
  - Store analysis in TestRun.CodeAnalysis
  ↓ success
[PLAN_GENERATED]
  - Call llm.GenerateTestPlan(analysis, requirements)
  - Validate plan has >= 1 scenario
  - Store in TestRun.TestPlan
  - Persist to DB
  ↓ success
[WRITING_TESTS]
  - For each scenario in TestPlan:
    - Call llm.GenerateTestScripts(scenario, analysis)
    - Validate: file is valid TypeScript / PHP
    - Store in TestRun.TestFiles
  - Persist to DB
  ↓ success
[RUNNING]
  - Create Steel Browser session
  - Write test files to temp directory
  - Connect Playwright to Steel CDP endpoint
  - Execute tests with timeout
  - Capture screenshots on failure
  - Parse results into RunResult
  - Destroy Steel session
  - Persist RunResult to DB
  ↓ if failed AND fix_attempts < 3
[FIXING]
  - Send failures + screenshots to LLM
  - LLM suggests patched test files
  - Update TestRun.TestFiles
  - Increment FixAttempts
  - → back to [RUNNING]
  ↓ if passed OR fix_attempts >= 3
[DONE]
  - Generate HTML report
  - Send notifications (if configured)
  - Update DB final state
```

## 6.2 Complex Agent (LangGraph Sidecar)

Used for: large projects, multi-agent critic review, deep analysis

Trigger: RunConfig.UseAdvancedAgent = true

HTTP Call from Go to LangGraph:
  POST http://langgraph-sidecar:8000/agent/run
  Body: { project_path, requirements, run_id, webhook_url }
  Response: { job_id }  (async)
  Webhook: POST to webhook_url on completion

LangGraph Python agents:

  PlannerAgent:
    input: codebase analysis
    task: break into atomic test scenarios with priority
    output: TestPlan JSON

  WriterAgent:
    input: single scenario + codebase context
    task: write Playwright test file for this scenario
    output: TestFile

  CriticAgent:
    input: generated TestFile
    task: review for: assertions, selectors quality, edge cases
    output: review_score (0-100) + improvement_suggestions

  RewriterAgent (if critic score < 70):
    input: TestFile + critic suggestions
    task: rewrite test file incorporating suggestions
    output: improved TestFile

  ExecutorAgent:
    input: TestFiles
    task: trigger execution via Go API, monitor progress
    output: RunResult

  FixerAgent:
    input: RunResult failures + screenshots
    task: analyze root cause, patch test files
    output: patched TestFiles

## 6.3 Prompt Contracts

All prompts are stored in internal/prompts/ as .txt template files.
Go loads them with embed.FS.

Files:
  analyze_codebase.txt
  generate_test_plan.txt
  generate_playwright_test.txt
  generate_phpunit_test.txt
  fix_failing_test.txt
  analyze_screenshot.txt
  generate_selector_from_visual.txt
  critic_review_test.txt

Each prompt template uses {{.Variable}} syntax.
Prompts are versioned — changing a prompt = new version in DB.
Braintrust logs prompt version with every LLM call.

---

# SECTION 7 — STEEL BROWSER SELF-HOST DEPLOYMENT

## 7.1 System Requirements

Minimum VPS spec for Steel self-host:
  RAM: 4GB (8GB recommended for 10 concurrent sessions)
  CPU: 2 vCPU (4 recommended)
  Disk: 20GB SSD
  OS: Ubuntu 22.04 / Debian 12
  Docker: 24.x+

## 7.2 Docker Compose Config for Steel

Add to docker-compose.yml:

```yaml
  steel-browser:
    image: ghcr.io/steel-dev/steel-browser:latest
    ports:
      - "3000:3000"     # Steel API
      - "5173:5173"     # Steel UI (optional, close in prod)
    environment:
      - STEEL_API_PORT=3000
      - MAX_CONCURRENT_SESSIONS=10
      - SESSION_TIMEOUT=300000     # 5 minutes in ms
      - BLOCK_ADS=true
      - STEEL_API_KEY=your-steel-api-key  # optional auth
    volumes:
      - steel_data:/data
      - /dev/shm:/dev/shm          # REQUIRED: shared memory for Chrome
    shm_size: '2gb'                # REQUIRED: prevent Chrome crashes
    cap_add:
      - SYS_ADMIN                  # REQUIRED: for Chrome sandbox
    restart: unless-stopped

volumes:
  steel_data:
```

## 7.3 Steel Session Lifecycle

1. GoTest Agent calls: POST http://steel:3000/v1/sessions
   Response: { sessionId, cdpUrl, seleniumUrl }

2. Playwright connects: chromium.connectOverCDP(cdpUrl)

3. Tests run using this browser instance

4. On completion/timeout: DELETE http://steel:3000/v1/sessions/:id

5. Steel cleans up all browser resources

## 7.4 Security Hardening for Steel

- DO NOT expose port 3000 to public internet
- Access only via internal Docker network
- Close port 5173 (UI) in production
- Set STEEL_API_KEY env var and pass from Go service
- Set SESSION_TIMEOUT to prevent zombie sessions (300000ms = 5min)
- Monitor: GET http://steel:3000/v1/sessions — alert if > 8 active

## 7.5 Steel Client in Go (internal/steel/client.go)

```go
type Client struct {
    baseURL string
    apiKey  string
    http    *http.Client
}

type Session struct {
    ID         string `json:"sessionId"`
    CDPURL     string `json:"cdpUrl"`
    SeleniumURL string `json:"seleniumUrl"`
    CreatedAt  string `json:"createdAt"`
}

func (c *Client) CreateSession(ctx context.Context) (*Session, error)
func (c *Client) GetSession(ctx context.Context, id string) (*Session, error)
func (c *Client) DestroySession(ctx context.Context, id string) error
func (c *Client) Screenshot(ctx context.Context, id string, fullPage bool) ([]byte, error)
func (c *Client) ListSessions(ctx context.Context) ([]*Session, error)
```

---

# SECTION 8 — LANGGRAPH SIDECAR DESIGN (Python)

## 8.1 Overview

Language: Python 3.12
Framework: LangGraph + LangChain
Transport: FastAPI HTTP server on :8000
Communication: Go calls LangGraph via HTTP, LangGraph calls Go API for execution

## 8.2 Directory Structure

```
sidecar/
├── main.py                 ← FastAPI app + LangGraph server
├── agents/
│   ├── planner.py          ← PlannerAgent
│   ├── writer.py           ← WriterAgent
│   ├── critic.py           ← CriticAgent
│   ├── executor.py         ← ExecutorAgent
│   └── fixer.py            ← FixerAgent
├── graph.py                ← LangGraph state graph definition
├── state.py                ← AgentState TypedDict
├── tools.py                ← Tool definitions (call Go API)
├── requirements.txt
└── Dockerfile
```

## 8.3 Agent State

```python
class AgentState(TypedDict):
    run_id: str
    project_path: str
    requirements: str
    code_analysis: str
    test_plan: dict
    test_files: list[dict]
    run_result: dict
    fix_attempts: int
    max_fixes: int
    critic_scores: list[float]
    screenshots: list[str]
    next_step: str
    error: Optional[str]
```

## 8.4 Graph Definition

```python
graph = StateGraph(AgentState)
graph.add_node("planner", planner_node)
graph.add_node("writer", writer_node)
graph.add_node("critic", critic_node)
graph.add_node("rewriter", rewriter_node)
graph.add_node("executor", executor_node)
graph.add_node("fixer", fixer_node)

graph.set_entry_point("planner")
graph.add_edge("planner", "writer")
graph.add_conditional_edges("critic", route_critic,
    {"pass": "executor", "fail": "rewriter"})
graph.add_edge("rewriter", "critic")
graph.add_edge("writer", "critic")
graph.add_conditional_edges("executor", route_executor,
    {"passed": END, "failed": "fixer", "max_fixes": END})
graph.add_edge("fixer", "executor")
```

## 8.5 FastAPI Endpoints

```
POST /agent/run    → start async LangGraph run
GET  /agent/:id    → get run status
GET  /health       → health check
```

## 8.6 Docker Compose for Sidecar

```yaml
  langgraph-sidecar:
    build: ./sidecar
    ports:
      - "8000:8000"
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - GOTEST_API_URL=http://app:8080
      - GOTEST_API_KEY=${API_KEY}
    depends_on:
      - app
    restart: unless-stopped
```

---

# SECTION 9 — BRAINTRUST EVAL FRAMEWORK

## 9.1 Overview

Braintrust is used to:
1. Measure quality of every LLM call
2. Track quality over time (experiment snapshots)
3. Run CI gates before deploying new prompts
4. Build dataset from production traces

Self-hosted option: use Braintrust OSS or log to braintrust.dev (free tier).

## 9.2 What Gets Evaluated

### Eval 1: Test Plan Quality
Input: codebase analysis + requirements
Output: generated test plan
Scorer: LLM judge (1-10 scale)
  - Coverage: does it cover main user flows?
  - Specificity: are steps concrete and actionable?
  - Priority accuracy: are high-risk areas marked high?
Pass threshold: >= 7/10

### Eval 2: Test Script Validity
Input: test scenario + generated .spec.ts file
Output: pass/fail
Scorer: static analysis
  - Is the file valid TypeScript?
  - Does it import @playwright/test?
  - Does it have at least 1 expect() assertion?
  - Does it have at least 1 await page.goto()?
Pass threshold: all checks pass

### Eval 3: Fix Success Rate
Input: (failure details, original test, fixed test, re-run result)
Output: pass/fail
Scorer: binary (did the re-run pass after fix?)
Target: >= 60% fix success rate

### Eval 4: Overall Run Quality
Input: (total tests, passed, failed, auto-fixed)
Output: quality score
Scorer: (passed + auto_fixed) / total
Target: >= 80%

## 9.3 Integration in Go

```go
// internal/evals/braintrust.go
type EvalLogger struct {
    client *braintrust.Client
    projectID string
}

func (e *EvalLogger) LogTestPlanGeneration(ctx context.Context,
    input TestPlanInput, output *TestPlan, score float64) error

func (e *EvalLogger) LogScriptGeneration(ctx context.Context,
    input ScriptInput, output TestFile, passed bool) error

func (e *EvalLogger) LogFixAttempt(ctx context.Context,
    input FixInput, output []TestFile, succeeded bool) error
```

## 9.4 CI Gate (GitHub Actions)

```yaml
- name: Run Evals
  run: |
    go run ./cmd/eval --experiment-name "ci-$GITHUB_SHA"
    # Fail if score drops more than 5% from baseline
```

---

# SECTION 10 — DATA MODEL & API CONTRACT

## 10.1 Database Schema

```sql
-- 001_init.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE projects (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(255) NOT NULL,
    path         TEXT,
    language     VARCHAR(50),
    framework    VARCHAR(50),
    config       JSONB DEFAULT '{}',
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE test_runs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID REFERENCES projects(id),
    state           VARCHAR(50) NOT NULL DEFAULT 'idle',
    mode            VARCHAR(20) DEFAULT 'simple',  -- simple | advanced
    requirements    TEXT,
    code_analysis   TEXT,
    test_plan       JSONB,
    test_files      JSONB,
    run_result      JSONB,
    screenshots     JSONB,
    fix_attempts    INT DEFAULT 0,
    error_msg       TEXT,
    duration_ms     INT,
    llm_tokens_used INT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash     VARCHAR(255) UNIQUE NOT NULL,
    name         VARCHAR(100),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE TABLE visual_baselines (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id   UUID REFERENCES projects(id),
    step_name    VARCHAR(255) NOT NULL,
    image_path   TEXT NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_runs_state ON test_runs(state);
CREATE INDEX idx_runs_project ON test_runs(project_id);
CREATE INDEX idx_runs_created ON test_runs(created_at DESC);
```

## 10.2 REST API Contract

### POST /api/v1/runs
Request:
```json
{
  "project_path": "/absolute/path/to/project",
  "requirements": "Test all checkout flows",
  "project_url": "http://localhost:3000",
  "mode": "simple",
  "config": {
    "max_fixes": 3,
    "timeout_seconds": 300,
    "use_visual_regression": false
  }
}
```
Response 202:
```json
{
  "run_id": "uuid",
  "state": "queued",
  "created_at": "2026-05-31T10:00:00Z"
}
```

### GET /api/v1/runs/:id
Response 200:
```json
{
  "id": "uuid",
  "state": "done",
  "test_plan": { "summary": "...", "scenarios": [...] },
  "run_result": {
    "passed": 8,
    "failed": 1,
    "total": 9,
    "fix_attempts": 1,
    "duration_ms": 45000,
    "failures": [{"test": "...", "message": "...", "screenshot_url": "..."}]
  },
  "created_at": "...",
  "finished_at": "..."
}
```

### GET /api/v1/runs/:id/stream (SSE)
Events:
```
event: state_change
data: {"state": "analyzing", "message": "Reading codebase..."}

event: test_plan
data: {"plan": {...}}

event: test_result
data: {"test": "Login flow", "status": "passed"}

event: done
data: {"summary": "9/9 tests passed"}
```

---

# SECTION 11 — DELIVERY ROADMAP

## Phase 1 — Foundation (Week 1-2)
Goal: Go backend compiles and MCP server works in Cursor

Tasks:
- [ ] Fix go.mod module name
- [ ] go mod tidy + all dependencies installed
- [ ] Rewrite llm_anthropic.go with official anthropic-sdk-go
- [ ] Fix import paths in all files
- [ ] Fix mcp/server.go (add runner import + tool contracts)
- [ ] Fix runner/docker.go (parse Playwright JSON reporter properly)
- [ ] Create internal/steel/client.go
- [ ] Replace Docker runner with Steel runner in agent
- [ ] make build compiles without errors
- [ ] MCP server works: type "analyze project" in Cursor → response

Acceptance criteria:
- `go build ./...` zero errors
- `make mcp` runs and Cursor can call `analyze_project` tool
- `make dev` starts HTTP server on :8080
- `curl /health` returns 200

## Phase 2 — End-to-End MVP (Week 3-4)
Goal: Full run from MCP → test execution → results

Tasks:
- [ ] Steel Browser running in docker-compose
- [ ] SteelPlaywrightRunner connects to Steel CDP endpoint
- [ ] Full agent loop works: analyze → plan → write → run
- [ ] Fix loop works (up to 3 attempts)
- [ ] Results persisted to PostgreSQL
- [ ] SSE streaming endpoint working
- [ ] GET /api/v1/runs/:id returns full result
- [ ] Basic HTML report generation

Acceptance criteria:
- Run a test on a real Next.js or CI4 project
- Agent successfully creates and runs at least 3 Playwright tests
- Results visible in DB and via API
- MCP tool `run_tests` returns structured summary in Cursor

## Phase 3 — Visual & Quality (Week 5-6)
Goal: Visual regression + eval pipeline

Tasks:
- [ ] GPT-4o Vision client (internal/vision)
- [ ] Screenshot capture on test failure
- [ ] Baseline screenshot management
- [ ] Visual diff using Steel screenshot API + Vision model
- [ ] Braintrust eval logging integrated
- [ ] Test plan quality scorer running
- [ ] Script validity scorer running

Acceptance criteria:
- Failed test captures screenshot automatically
- Visual diff detects UI changes (>5% = flag)
- Braintrust dashboard shows eval scores for each run

## Phase 4 — LangGraph + Advanced Agent (Week 7-8)
Goal: Multi-agent sophisticated loop

Tasks:
- [ ] sidecar/ Python project created
- [ ] LangGraph graph defined with 5 agents
- [ ] FastAPI server running on :8000
- [ ] Go agent calls sidecar for mode=advanced runs
- [ ] Critic agent reviewing generated tests
- [ ] Self-healing selector fix using vision

Acceptance criteria:
- mode=advanced run shows critic scores in results
- Self-healing test recovers from selector failure
- Sidecar runs in docker-compose alongside Go

## Phase 5 — SaaS Layer (Week 9-10)
Goal: Multi-user, billing, public dashboard

Tasks:
- [ ] User auth (JWT, registration, login)
- [ ] API key management per user
- [ ] Next.js dashboard frontend
- [ ] Stripe billing integration
- [ ] GitHub webhook integration
- [ ] Report sharing (public URLs)
- [ ] Langfuse observability for LLM costs

Acceptance criteria:
- New user can register, get API key, run tests
- Dashboard shows run history with pass/fail chart
- Stripe checkout working (test mode)
- GitHub push triggers test run via webhook

---

# SECTION 12 — AI AGENT EXECUTION INSTRUCTIONS

## How to Use This Document

Give this entire file to your AI agent (Cursor/Claude Code) and say:
"Execute this build plan starting from Phase 1, Section 12.1"

The agent should:
1. Read Section 12.2 for current file state
2. Read Section 12.3 for common error fixes
3. Execute one phase at a time
4. Verify acceptance criteria before moving to next phase
5. Update this document's task checkboxes as it completes work

## 12.1 Execution Order

PHASE 1 (Foundation):
1. Fix go.mod module name (replace "yourusername" with actual GitHub username)
2. Run: go get github.com/anthropics/anthropic-sdk-go@latest
3. Run: go mod tidy
4. Rewrite internal/agent/llm_anthropic.go (see spec in Section 5.1.4)
5. Create internal/steel/client.go (see spec in Section 7.5)
6. Fix internal/mcp/server.go imports
7. Fix internal/runner/docker.go JSON parsing
8. Update internal/runner to add SteelPlaywrightRunner
9. Run: go build ./... — fix any remaining errors
10. Run: make build — verify both binaries compile
11. Test: add MCP to Cursor config, test analyze_project tool

PHASE 2 (E2E):
1. Update docker-compose.yml with Steel Browser service (see Section 7.2)
2. Create internal/db/store.go implementing RunStore interface
3. Create internal/db/migrate.go for auto-migration on startup
4. Update internal/queue/worker.go to persist results to DB
5. Add SSE endpoint to internal/api/server.go
6. Implement HTML report generator (internal/report/html.go)
7. Run: docker-compose up -d
8. Run: make docker-playwright (if using fallback)
9. Run: make dev
10. Test end-to-end with a real project

PHASE 3 (Vision + Evals):
1. Create internal/vision/client.go
2. Update agent to capture screenshots on failure
3. Create internal/evals/braintrust.go
4. Integrate eval logging into each LLM call in llm_anthropic.go
5. Test: run eval suite with make eval

PHASE 4 (LangGraph Sidecar):
1. Create sidecar/ directory
2. Create sidecar/requirements.txt with: langgraph, langchain-anthropic, fastapi, uvicorn
3. Create sidecar/state.py, graph.py, main.py, agents/*.py
4. Add sidecar to docker-compose.yml
5. Update internal/agent/agent.go to check RunConfig.Mode
6. If mode=advanced: POST to http://langgraph-sidecar:8000/agent/run

## 12.2 Current File State (after scaffold)

Files that exist and are CORRECT (no changes needed):
  ✅ internal/config/config.go
  ✅ internal/db/migrations/001_init.sql (needs expansion per Section 10.1)
  ✅ docker-compose.yml (needs Steel + sidecar added)
  ✅ Makefile (needs eval target added)
  ✅ Dockerfile
  ✅ deployments/docker/Dockerfile.playwright
  ✅ .env.example (needs new vars added)
  ✅ cmd/server/main.go
  ✅ cmd/mcp/main.go
  ✅ internal/agent/agent.go

Files that NEED REWRITE / FIX:
  ❌ internal/agent/llm_anthropic.go  → rewrite with anthropic-sdk-go
  ❌ internal/mcp/server.go           → fix missing runner import
  ❌ internal/runner/docker.go        → fix JSON parsing

Files that need to be CREATED:
  ➕ internal/steel/client.go
  ➕ internal/db/store.go
  ➕ internal/db/migrate.go
  ➕ internal/vision/client.go         (Phase 3)
  ➕ internal/evals/braintrust.go      (Phase 3)
  ➕ internal/report/html.go
  ➕ internal/prompts/*.txt (all prompt templates)
  ➕ sidecar/ (entire Python sidecar)  (Phase 4)
  ➕ frontend/ (Next.js dashboard)     (Phase 5)

## 12.3 Common Errors & Fixes

ERROR: cannot find module providing anthropic-sdk-go
FIX: go get github.com/anthropics/anthropic-sdk-go@latest

ERROR: undefined: runner in mcp/server.go
FIX: add import "github.com/YOUR_MODULE/internal/runner"

ERROR: Claude response JSON wrapped in markdown
FIX: Strip markers before parsing:
  s = strings.TrimPrefix(s, "```json\n")
  s = strings.TrimPrefix(s, "```\n")
  s = strings.TrimSuffix(s, "```")
  s = strings.TrimSpace(s)

ERROR: Steel Browser container crashes (Chrome)
FIX: Add to docker-compose: shm_size: '2gb' AND /dev/shm:/dev/shm volume

ERROR: Playwright cannot connect to Steel CDP
FIX: Steel CDP URL format: ws://steel:3000/v1/sessions/{id}/cdp
  Use: chromium.connectOverCDP(cdpUrl) in Playwright

ERROR: Port 3000 conflict with Next.js
FIX: Change Steel port to 3010 in docker-compose and steel client

ERROR: asynq worker cannot connect to Redis
FIX: Ensure Redis is running: docker-compose up -d redis
  Verify REDIS_URL=redis:6379 (not localhost:6379 inside Docker)

ERROR: pgx connection refused
FIX: DATABASE_URL must use service name: postgres://postgres:password@postgres:5432/db
  Not localhost inside Docker network

ERROR: LangGraph sidecar cannot call Go API
FIX: GOTEST_API_URL must be http://app:8080 (Docker service name, not localhost)

## 12.4 Environment Variables Reference (.env.example full version)

# Application
APP_ENV=development
APP_PORT=8080
API_KEY=your-secret-api-key-here

# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/gotest_agent?sslmode=disable

# Redis
REDIS_URL=localhost:6379

# LLM - Primary
ANTHROPIC_API_KEY=sk-ant-xxxx
LLM_PROVIDER=anthropic
LLM_MODEL=claude-sonnet-4-5

# LLM - Vision (Phase 3)
OPENAI_API_KEY=sk-xxxx
VISION_MODEL=gpt-4o

# LiteLLM Gateway (optional)
LITELLM_URL=http://localhost:4000
LITELLM_API_KEY=sk-litellm-xxxx

# Steel Browser
STEEL_API_URL=http://localhost:3000
STEEL_API_KEY=your-steel-key
STEEL_MAX_SESSIONS=10

# LangGraph Sidecar (Phase 4)
LANGGRAPH_URL=http://localhost:8000

# Braintrust Evals (Phase 3)
BRAINTRUST_API_KEY=your-braintrust-key

# Langfuse Observability
LANGFUSE_SECRET_KEY=sk-lf-xxxx
LANGFUSE_PUBLIC_KEY=pk-lf-xxxx
LANGFUSE_HOST=http://localhost:3001

# Storage
SCREENSHOTS_PATH=./data/screenshots
REPORTS_PATH=./data/reports

# Agent Config
MAX_FIX_ATTEMPTS=3
DEFAULT_TIMEOUT_SECONDS=300
ENABLE_VISUAL_REGRESSION=false
ENABLE_ADVANCED_AGENT=false
