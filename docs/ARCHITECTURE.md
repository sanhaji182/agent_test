# GoTest Agent Architecture

## Overview

GoTest Agent adalah platform autonomous testing yang menggunakan AI untuk menganalisis codebase, generate test, menjalankan test, dan auto-fix test yang gagal. Sistem ini dibangun dengan arsitektur modular yang memungkinkan ekstensi dan scaling yang mudah.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│                   Next.js + React + Tailwind                 │
│                  http://localhost:3001                        │
└────────────────────┬────────────────────────────────────────┘
                     │ REST API
                     │
┌────────────────────▼────────────────────────────────────────┐
│                      API Layer (Go)                          │
│                   Chi Router :8080                            │
│  ┌──────────────┬──────────────┬──────────────────────────┐│
│  │   Auth       │   Rate       │   CORS                   ││
│  │   Middleware │   Limiter    │   Middleware             ││
│  └──────────────┴──────────────┴──────────────────────────┘│
│  ┌──────────────────────────────────────────────────────┐  │
│  │              REST API Handlers                        │  │
│  │  /runs, /schedules, /webhooks, /metrics, /reviews   │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
   ┌─────────┐  ┌─────────┐  ┌─────────────┐
   │ Parser  │  │   AI    │  │  Executor   │
   │ Service │  │ Service │  │   Service   │
   └────┬────┘  └────┬────┘  └──────┬──────┘
        │            │              │
        │            │              │
        ▼            ▼              ▼
┌──────────────────────────────────────────────────────────┐
│                  Core Services                            │
│  ┌──────────────┬──────────────┬──────────────────────┐ │
│  │   Agent      │   Synthesis  │   Confidence         │ │
│  │   Service    │   Service    │   Scorer             │ │
│  └──────────────┴──────────────┴──────────────────────┘ │
└────────────────────┬─────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐
   │PostgreSQL│  │  Redis  │  │ Steel   │
   │  :5432  │  │  :6379  │  │ Browser │
   └─────────┘  └─────────┘  └─────────┘
```

## Component Architecture

### API Layer

#### Server (`internal/api/server.go`)
- **Port**: 8080
- **Router**: Chi Router v5
- **Middleware Stack**:
  - Request ID (untuk tracing)
  - Logger (request logging)
  - Recoverer (panic recovery)
  - Rate Limiter (100 req/min/IP)
  - CORS (configurable)

#### Handlers
| Handler File | Responsibility |
|--------------|----------------|
| `handlers_runs.go` | Test run management (create, list, get, cancel, delete) |
| `handlers_schedules.go` | Schedule management (CRUD) |
| `handlers_webhooks.go` | GitHub webhook handling |
| `handlers_metrics.go` | Metrics and insights |
| `handlers_reviews.go` | Review workflow |
| `handlers_monitoring.go` | System monitoring |
| `handlers_generation.go` | Test generation |
| `handlers_analysis.go` | Codebase analysis |
| `handlers_health.go` | Health check |

### Parser Layer

#### Parser Registry (`internal/parser/`)
Multi-language parser yang mendukung:
- **JavaScript/TypeScript**: Express, Next.js, NestJS
- **Go**: Chi, Gin, Echo, Fiber
- **PHP**: Laravel, Symfony
- **Python**: Django, Flask, FastAPI
- **Ruby**: Rails (planned)
- **Java**: Spring Boot (planned)
- **C#**: ASP.NET Core (planned)
- **Rust**: Actix, Rocket (planned)

#### Parser Interface
```go
type Parser interface {
    Parse(ctx context.Context, rootDir string) (*Codebase, error)
    SupportedLanguages() []string
    DetectFramework(rootDir string) (string, error)
}
```

#### Codebase Structure
```go
type Codebase struct {
    Language   string    // "javascript", "go", "php", "python"
    Framework  string    // "express", "chi", "laravel", "django"
    Routes     []Route   // Extracted routes
    Models     []Model   // Extracted models
    Handlers   []Handler // Extracted handlers
    AnalyzedAt time.Time
    FileCount  int
}
```

### AI Layer

#### AI Client (`internal/ai/client.go`)
Multi-provider LLM client dengan circuit breaker:
- **Supported Providers**: Anthropic, OpenAI, Google Gemini, DeepSeek, Mistral, Groq, OpenRouter, Local
- **Resilience**: Exponential backoff (1s → 2s → 4s, max 8s)
- **Circuit Breaker**: Opens after 5 failures, recovers after 30s

#### Synthesis Service (`internal/ai/synthesis.go`)
```go
type SynthesisService struct {
    client Client
}

func (s *SynthesisService) GenerateTestPlan(ctx context.Context, codebase *Codebase) (*TestPlan, error)
```

**Workflow**:
1. Analyze codebase (routes, models, handlers)
2. Generate test plan dengan AI
3. Score confidence untuk setiap test case
4. Return structured test plan

#### Confidence Scorer (`internal/ai/confidence_scorer.go`)
```go
type ConfidenceScorer struct{}

func (cs *ConfidenceScorer) ScoreTestCase(tc *TestCase, codebase *Codebase) int
```

**Scoring Factors**:
- Route coverage
- Model field coverage
- Handler parameter coverage
- Test scenario completeness

### Executor Layer

#### Agent Service (`internal/agent/agent.go`)
Core test execution engine:
```go
type Agent struct {
    llm            LLM
    runner         Runner
    maxFixAttempts int
    screenshotter  ScreenshotCapturer
    exec           *execution.Context
    store          RunPersistence
}
```

**Execution Flow**:
```
1. Analyze codebase
2. Generate test plan
3. Generate test scripts
4. Execute tests
5. If failures:
   - Capture screenshots
   - AI-powered self-healing (up to 3 attempts)
   - Re-execute healed tests
6. Generate report
```

#### Playwright Runner (`internal/agent/playwright_runner.go`)
```go
type PlaywrightRunner struct {
    browser       string    // chromium, firefox, webkit
    viewport      string    // desktop, mobile
    parallel      bool
    videoDir      string
    screenshotDir string
}
```

**Supported Actions**:
- Navigation: `goto`, `back`, `forward`, `reload`
- Interaction: `click`, `fill`, `select`, `hover`, `press`
- Verification: `assert_visible`, `assert_hidden`, `assert_text`, `assert_url`, `assert_title`
- Utilities: `wait`, `scroll`, `screenshot`

#### Self-Healing
```go
func (r *PlaywrightRunner) selfHeal(action *Action, error error, attempt int) (*Action, error)
```

**Healing Strategy**:
1. Analyze error message
2. Use AI to suggest selector fix
3. Apply heuristic selector updates
4. Retry action with healed selector
5. Max 3 attempts per action

### Storage Layer

#### PostgreSQL (`internal/db/`)
```go
type Store struct {
    pool *pgxpool.Pool
}
```

**Tables**:
- `runs` - Test run records
- `test_plans` - Generated test plans (JSONB)
- `test_files` - Generated test files
- `schedules` - Scheduled test runs
- `reviews` - Review workflow
- `webhook_registrations` - GitHub webhook registrations
- `drifts` - Drift detection records

#### Redis (`internal/queue/`)
```go
type Queue struct {
    client *asynq.Client
}
```

**Queues**:
- `runs:execute` - Test execution queue
- `runs:heal` - Self-healing queue
- `schedules:trigger` - Schedule trigger queue

### External Services

#### Steel Browser (`internal/steel/`)
Cloud browser automation:
```go
type Client struct {
    baseURL string
    apiKey  string
}

func (c *Client) CreateSession(ctx context.Context) (*Session, error)
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error)
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error
```

#### LangGraph Sidecar (`internal/sidecar/`)
Advanced AI processing:
```go
type SidecarClient struct {
    baseURL string
}

func (c *SidecarClient) AnalyzeCodebase(ctx context.Context, codebase *Codebase) (*Analysis, error)
func (c *SidecarClient) GenerateTests(ctx context.Context, plan *TestPlan) ([]TestFile, error)
```

## Data Flow

### Test Execution Flow

```
1. User Request
   POST /api/v1/runs
   {project_path, requirements}

2. API Handler
   └─> Create Run (PostgreSQL)
   └─> Queue Execution (Redis)

3. Execution Worker
   └─> Parse Codebase
   └─> Generate Test Plan (AI)
   └─> Generate Test Scripts (AI)
   └─> Execute Tests (Playwright)
   
4. Self-Healing Loop (if failures)
   └─> Capture Screenshots
   └─> AI Analyze Failure
   └─> Generate Fix
   └─> Re-execute (max 3 attempts)

5. Report Generation
   └─> Generate HTML Report
   └─> Store Results (PostgreSQL)
   └─> Send Notifications (Webhook)

6. Response
   └─> Update Run Status
   └─> Return Results
```

### Drift Detection Flow

```
1. GitHub Webhook
   POST /api/v1/webhooks/github
   {push event}

2. Webhook Handler
   └─> Verify Signature
   └─> Parse Push Event
   └─> Extract Changed Files

3. Drift Detection
   └─> Filter Relevant Files
   └─> Parse Changed Files
   └─> Compare with Existing Tests
   └─> Detect Drift

4. Auto-Generation
   └─> Generate Tests for New Code
   └─> Update Existing Tests
   └─> Create Drift Records

5. Notification
   └─> Alert Team (Slack/Teams)
   └─> Update Run Status
```

## Technology Stack

### Backend
- **Language**: Go 1.26.4
- **Framework**: Chi Router v5.2.1
- **Database**: PostgreSQL 16.14
- **Queue**: Redis + Asynq
- **Browser**: Playwright (Steel Browser)
- **AI**: Multi-provider (Anthropic, OpenAI, Google, DeepSeek, etc.)
- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus

### Frontend
- **Framework**: Next.js 16.2.7
- **UI**: React 19.2.7
- **Styling**: Tailwind CSS
- **State**: React Context

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Services**: 6 containers (backend, frontend, postgres, redis, steel-browser, langgraph-sidecar)

## Security Considerations

### API Security
- API Key authentication (configurable)
- Rate limiting (100 req/min/IP)
- CORS (configurable)
- Request size limiting (10MB)

### Data Security
- PostgreSQL credentials in environment variables
- API keys in environment variables
- GitHub webhook signature verification (HMAC-SHA256)

### Browser Security
- Sandboxed browser execution (Steel Browser)
- Isolated browser sessions
- Automatic session cleanup

### Production Recommendations
1. **Change database credentials** - Use strong random passwords
2. **Rotate secrets** - JWT_SECRET, API_KEY, GITHUB_WEBHOOK_SECRET
3. **Bind ports** - Bind database/Redis to 127.0.0.1
4. **TLS termination** - Use reverse proxy (Caddy, nginx, Traefik)
5. **Rate limiting** - Implement production-grade rate limiting
6. **Docker secrets** - Use Docker secrets for sensitive values
7. **Network isolation** - Private Docker network for database/Redis

## Scalability

### Horizontal Scaling
```
Load Balancer
    ├─> Backend Instance 1
    ├─> Backend Instance 2
    └─> Backend Instance 3
         │
         ├─> PostgreSQL (Primary)
         ├─> PostgreSQL (Replica)
         └─> Redis Cluster
```

### Queue-Based Scaling
```
API Layer
    └─> Queue (Redis/Asynq)
         │
         ├─> Worker 1 (Execute)
         ├─> Worker 2 (Execute)
         ├─> Worker 3 (Heal)
         └─> Worker 4 (Schedule)
```

## Monitoring

### Metrics (Prometheus)
```
gotest_runs_total              # Total test runs
gotest_tests_total             # Total tests executed
gotest_tests_passed            # Tests passed
gotest_tests_failed            # Tests failed
gotest_tests_healed            # Tests healed
gotest_execution_duration_ms   # Execution duration
```

### Tracing (Jaeger/OpenTelemetry)
- Request tracing
- Execution timing
- Error tracking
- Performance profiling

### Logging
- Structured logging (slog)
- Request ID correlation
- Error stack traces
- Performance metrics

## Deployment

### Docker Compose (Development)
```yaml
services:
  backend:      # :8080
  frontend:     # :3001
  postgres:     # :5432
  redis:        # :6379
  steel-browser: # :3010
  langgraph-sidecar: # :8000
```

### Production Deployment
1. **Kubernetes** (recommended)
   - StatefulSet for PostgreSQL
   - StatefulSet for Redis
   - Deployment for backend
   - Deployment for frontend
   - HPA for auto-scaling

2. **Docker Swarm**
   - Stack deployment
   - Service discovery
   - Load balancing

3. **Cloud Platforms**
   - AWS ECS/EKS
   - Google Cloud Run
   - Azure Container Instances

## Future Enhancements

### Phase 2: Record & Playback
- Chrome Extension for recording user interactions
- AI-powered action extraction
- Test library management
- Test composition and reuse

### Phase 3: Continuous Sync
- GitHub webhook integration
- Drift detection
- Auto-regeneration on code changes
- Real-time notifications

### Phase 4: Enterprise Features
- SSO (SAML, OIDC)
- RBAC (admin, developer, viewer)
- Audit logging
- Advanced analytics
- SLA guarantees

## References

- [API Documentation](./API.md)
- [Phase 1 Plan](./PHASE-1-PLAN.md)
- [Phase 2 Plan](./PHASE-2-PLAN.md)
- [Phase 3 Plan](./PHASE-3-PLAN.md)
- [Phase 4 Plan](./PHASE-4-PLAN.md)
- [Setup Guide](./SETUP.md)
