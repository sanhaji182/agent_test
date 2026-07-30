# GoTest Agent — Master Development Plan

**Status:** Active  
**Created:** 2026-07-30  
**Last Updated:** 2026-07-30  
**Owner:** Engineering Team

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State](#2-current-state)
3. [Vision & Roadmap](#3-vision--roadmap)
4. [Core Features](#4-core-features)
5. [Technical Architecture](#5-technical-architecture)
6. [Development Phases](#6-development-phases)
7. [Decision Log](#7-decision-log)
8. [Open Questions](#8-open-questions)
9. [Success Metrics](#9-success-metrics)
10. [Resources](#10-resources)

---

## 1. Executive Summary

### What is GoTest Agent?

GoTest Agent is an **AI-powered end-to-end testing platform** that automatically generates, executes, and maintains test suites for web applications. Unlike traditional test automation tools that require manual test writing or recording, GoTest Agent uses AI to:

1. **Analyze codebases** and understand application structure
2. **Generate comprehensive test plans** from requirements or code
3. **Create executable tests** in multiple frameworks (Playwright, Cypress, Selenium)
4. **Self-heal tests** when UI changes
5. **Continuously sync** with code changes

### Value Proposition

**For Developers:**
- Zero manual test writing — AI generates tests from code or requirements
- Self-healing tests — no more fixing broken selectors
- Multi-framework export — no vendor lock-in

**For QA Teams:**
- 10x faster test creation
- AI-powered exploratory testing
- Comprehensive coverage without manual effort

**For Organizations:**
- 80% reduction in test maintenance time
- 50% increase in test coverage
- Faster time-to-market

### Competitive Positioning

**vs Katalon:**
- ✅ AI-native (Katalon is traditional automation)
- ✅ Self-healing (Katalon requires manual fixes)
- ✅ Multi-framework export (Katalon is proprietary)
- ✅ 10x cheaper (self-hosted vs $175-667/month/user)
- ❌ Less mature (Katalon has 8+ years head start)
- ❌ Fewer features (no mobile/desktop testing yet)

**vs TestSprite:**
- ✅ Codebase analysis (TestSprite only does requirement-based generation)
- ✅ Continuous sync (TestSprite is one-time generation)
- ✅ Self-healing (TestSprite has limited self-healing)
- ✅ Open-source option (TestSprite is SaaS only)

**Unique Selling Point:**
> "Upload your code, get comprehensive test suite in 5 minutes"

No competitor offers codebase → test suite transformation. This is our blue ocean.

---

## 2. Current State

### What's Already Built (as of 2026-07-30)

**Platform Features (25+ production features):**
- ✅ AI-powered test generation (requirement-based)
- ✅ Self-healing tests (LLM + vision analysis)
- ✅ Multi-browser support (Chromium, Firefox, WebKit)
- ✅ 8 viewport presets (mobile + desktop)
- ✅ Parallel test execution
- ✅ 15+ browser actions (click, type, navigate, etc)
- ✅ Network interception
- ✅ Test data parameterization
- ✅ Run cancellation
- ✅ Rate limiting (100 req/min/IP)
- ✅ Prometheus metrics
- ✅ Distributed tracing (OpenTelemetry)
- ✅ Lifecycle webhooks (Slack/Teams)
- ✅ JUnit XML export
- ✅ Code export (Playwright, Cypress, Selenium, Puppeteer)
- ✅ Advanced testing (audit, exploratory, performance, accessibility, visual regression)
- ✅ Test tags & filtering
- ✅ LLM resilience (retry + circuit breaker)
- ✅ Docker multi-stage hardened build
- ✅ Graceful shutdown
- ✅ MCP server integration
- ✅ Intelligence engine (risk scoring, confidence, flaky detection)
- ✅ Human-in-the-loop review workflow
- ✅ Schedule-based execution (cron)
- ✅ GitHub webhook integration
- ✅ And more...

**Technical Stack:**
- **Backend:** Go (Chi router, PostgreSQL, Redis, Asynq)
- **Frontend:** Next.js 16 + React 19
- **AI:** Anthropic Claude, OpenAI GPT, Google Gemini, DeepSeek, Mistral, Groq, OpenRouter
- **Browser Automation:** Playwright-Go
- **Infrastructure:** Docker, Prometheus, Jaeger (OpenTelemetry)

**Test Coverage:**
- ✅ 25/25 packages passing
- ✅ E2E smoke test (12/12 actions passed)
- ✅ Distributed tracing verified (Jaeger)

**Deployment:**
- ✅ Docker Compose (backend, frontend, Jaeger, PostgreSQL, Redis, Steel Browser, LangGraph sidecar)
- ✅ Running on http://localhost:3001 (frontend), http://localhost:8080 (API), http://localhost:16686 (Jaeger)

### What's NOT Built Yet (Core Features)

**Missing Features:**
- ❌ **Codebase Analysis** — Parse GitHub repos, extract routes/models/schemas, generate tests
- ❌ **Record & Playback** — Record user actions, AI understand intent, generate reusable tests
- ❌ **Multi-language Support** — Parse JS/TS, Go, PHP (Laravel), Python
- ❌ **GitHub Integration** — OAuth, webhook, continuous sync
- ❌ **Test Library** — Reusable test components, composition
- ❌ **Self-healing (advanced)** — Auto-update selectors when UI changes
- ❌ **Mobile Testing** — iOS/Android (out of scope for now)
- ❌ **Desktop Testing** — Windows/macOS/Linux apps (out of scope)

### Current Limitations

1. **AI only understands requirements text** — Cannot analyze actual codebase
2. **No record & playback** — Users must write requirements manually
3. **No GitHub integration** — Cannot auto-sync with code changes
4. **No test library** — Cannot reuse test components
5. **Limited self-healing** — Only basic selector retry, no AI-driven selector update

---

## 3. Vision & Roadmap

### Vision Statement

> "GoTest Agent becomes the default AI-powered testing platform for modern development teams, replacing manual test writing and traditional automation tools."

### 12-Month Roadmap

#### **Phase 1: Codebase Analysis (Months 1-3)**
**Goal:** Enable "upload code → get tests" workflow

**Deliverables:**
- [ ] GitHub OAuth integration
- [ ] Multi-language parser (JS/TS, Go, PHP, Python)
- [ ] AI codebase analysis (extract routes, models, schemas, business rules)
- [ ] Auto-generate test plan from code
- [ ] Auto-generate executable tests (Playwright)
- [ ] Manual review & edit UI
- [ ] Test execution & reporting

**Success Criteria:**
- 50 beta users
- 500+ tests generated from codebases
- 70% generated test pass rate
- NPS > 40

#### **Phase 2: Record & Playback (Months 4-6)**
**Goal:** Enable "record user actions → AI generate test" workflow

**Deliverables:**
- [ ] Chrome extension for recording
- [ ] AI intent classification
- [ ] Selector optimization (stability scoring)
- [ ] Assertion inference
- [ ] Export to multiple frameworks (Playwright, Cypress, Selenium)
- [ ] Test library (save, browse, reuse)

**Success Criteria:**
- 200 paying customers
- 5,000+ tests recorded
- 80% test pass rate
- $50k MRR

#### **Phase 3: Continuous Sync (Months 7-9)**
**Goal:** Auto-update tests when code changes

**Deliverables:**
- [ ] GitHub webhook integration
- [ ] PR-based test generation (test only changed code)
- [ ] Drift detection (code vs tests out of sync)
- [ ] Auto-regenerate tests for changed routes
- [ ] Notification system (alert when tests need update)

**Success Criteria:**
- 500 paying customers
- 10,000+ tests in library
- 90% test pass rate
- $150k MRR

#### **Phase 4: Enterprise Features (Months 10-12)**
**Goal:** Enterprise-ready platform

**Deliverables:**
- [ ] SSO (SAML, OIDC)
- [ ] RBAC (admin, developer, viewer roles)
- [ ] Audit log
- [ ] On-premise deployment option
- [ ] SOC 2 compliance
- [ ] SLA guarantees
- [ ] Advanced analytics (coverage gaps, flaky tests, trends)

**Success Criteria:**
- 1,000 paying customers
- 50,000+ tests in library
- 95% test pass rate
- $500k MRR
- 10 enterprise customers ($10k+/month)

### 3-Year Vision

**Year 1:** Establish product-market fit, 1,000 customers, $500k MRR  
**Year 2:** Scale to 10,000 customers, $5M ARR, Series A fundraise  
**Year 3:** Market leader in AI testing, 50,000 customers, $50M ARR

---

## 4. Core Features

### Feature 1: Codebase Analysis (Core)

**User Workflow:**
```
1. User connects GitHub account (OAuth)
2. User selects repo (public or private)
3. System clones repo to sandboxed environment
4. Parser analyzes code:
   - Detect language (JS, Go, PHP, Python)
   - Detect framework (Express, Chi, Laravel, Django)
   - Extract routes (endpoints, HTTP methods)
   - Extract models (schemas, relationships, validation rules)
   - Extract business rules (from comments, README)
   - Extract test data (from seeds, fixtures)
5. AI synthesizes all information:
   - Generate test plan (scenarios, steps)
   - Generate executable tests (Playwright code)
   - Generate test data examples
6. User reviews & edits test plan
7. User triggers test execution
8. System runs tests, generates report
```

**Technical Components:**

**Parser (per language):**
- **JavaScript/TypeScript:** Parse Express.js routes (`app.get('/users', handler)`)
- **Go:** Parse Chi/Gin routes (`r.Get("/users", handler)`)
- **PHP (Laravel):** Parse Laravel routes (`Route::get('/users', [Controller::class, 'index'])`)
- **Python:** Parse Django/FastAPI routes (`@app.get("/users")`)

**AI Synthesis:**
- Feed parsed code to LLM
- Generate test plan (scenarios, steps, assertions)
- Generate executable tests (Playwright, Cypress, Selenium)
- Generate test data (based on schemas, validation rules)

**Security:**
- Sandboxed Docker container (isolated, auto-delete after 24h)
- Security scan before processing (malware detection)
- Private repo support (GitHub token)
- On-premise option for enterprise

**Supported Languages (MVP):**
1. JavaScript/TypeScript (Express.js) — largest market
2. Go (Chi + Gin) — our stack, type-safe
3. PHP (Laravel) — massive SME market
4. Python (FastAPI + Flask) — data science community

**Future Languages:**
- Java (Spring Boot)
- Ruby (Rails)
- C# (.NET)

### Feature 2: Record & Playback (Additional)

**User Workflow:**
```
1. User clicks "Record Test"
2. Chrome extension starts recording
3. User interacts with web app (click, type, navigate)
4. Extension captures DOM events + context
5. AI processes recording:
   - Classify intent (login, search, form submission)
   - Optimize selectors (stability scoring)
   - Generate natural language descriptions
   - Infer assertions (what should be verified)
6. System generates structured test (JSON + Markdown)
7. User reviews & edits test
8. User exports to Playwright/Cypress/Selenium
9. User saves to test library (reusable)
```

**Technical Components:**

**Chrome Extension:**
- Capture DOM events (click, input, navigate, scroll)
- Extract context (selectors, text, attributes)
- Send events to backend via WebSocket
- Visual feedback (recording indicator, step counter)

**AI Processing:**
- Intent classification (authentication, search, form, etc)
- Selector optimization (stability scoring 0-100)
- Natural language generation (human-readable descriptions)
- Assertion inference (what should be verified)

**Test Library:**
- Store in database (JSONB for flexibility)
- Search & filter (tags, intent, etc)
- Reuse & composition (reference other tests)
- Version control (Git-like)

### Feature 3: Continuous Sync (Phase 3)

**User Workflow:**
```
1. User connects GitHub repo
2. System generates initial test suite
3. Developer pushes code change
4. GitHub webhook triggers system
5. System analyzes diff:
   - Which routes changed?
   - Which models changed?
   - Which business rules changed?
6. System regenerates affected tests only
7. System runs tests, generates report
8. System notifies user (if tests fail)
```

**Technical Components:**

**GitHub Webhook:**
- Listen to push events
- Extract diff (changed files, lines)
- Map changes to routes/models

**Diff Analysis:**
- Parse changed files
- Determine impact (which tests affected)
- Regenerate only affected tests

**Notification:**
- Slack/Teams webhook
- Email
- In-app notification

---

## 5. Technical Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend Layer                          │
│  Next.js 16 + React 19                                       │
│  - Dashboard, Recorder UI, Test Viewer, Library             │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ REST API
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend Layer                           │
│  Go (Chi router)                                             │
│  - Recorder Service (receive events, store raw data)        │
│  - AI Engine Service (process recordings, generate tests)   │
│  - Parser Service (analyze codebases)                       │
│  - Exporter Service (generate Playwright/Cypress/Selenium)  │
│  - Test Library Service (store, search, reuse)              │
│  - Execution Service (run tests, generate reports)          │
└─────────────────────────────────────────────────────────────┘
                          │
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                              │
│  PostgreSQL (test storage, JSONB)                            │
│  Redis (session management, job queue)                       │
│  S3-compatible storage (screenshots, videos, recordings)     │
└─────────────────────────────────────────────────────────────┘
                          │
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      AI Layer                                │
│  OpenAI GPT-4, Anthropic Claude, Llama 3.1                   │
│  - Intent classification                                     │
│  - Natural language generation                               │
│  - Selector optimization                                     │
│  - Assertion inference                                       │
│  - Codebase analysis                                         │
│  - Test plan generation                                      │
└─────────────────────────────────────────────────────────────┘
                          │
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Execution Layer                         │
│  Playwright (browser automation)                             │
│  - Chromium, Firefox, WebKit                                 │
│  - Self-healing (AI-driven selector update)                  │
│  - Screenshot & video recording                              │
│  - Network interception                                      │
└─────────────────────────────────────────────────────────────┘
                          │
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Observability Layer                     │
│  Prometheus (metrics)                                        │
│  Jaeger (distributed tracing, OpenTelemetry)                 │
│  Grafana (dashboards)                                        │
│  ELK Stack (logs, optional)                                  │
└─────────────────────────────────────────────────────────────┘
```

### Component Breakdown

**Frontend (Next.js):**
- `pages/recorder.tsx` — Recorder UI (start/stop recording)
- `pages/tests/[id].tsx` — Test viewer (view, edit, export)
- `pages/library.tsx` — Test library (browse, search, reuse)
- `pages/codebases/[id].tsx` — Codebase analysis UI
- `components/` — Reusable components (buttons, cards, modals)

**Backend (Go):**
- `cmd/server/main.go` — HTTP server entry point
- `internal/api/` — REST API handlers
- `internal/recorder/` — Recorder service (event streaming)
- `internal/ai/` — AI engine (LLM integration, prompts)
- `internal/parser/` — Codebase parser (multi-language)
- `internal/exporter/` — Test exporter (Playwright, Cypress, Selenium)
- `internal/library/` — Test library (storage, search)
- `internal/execution/` — Test execution (Playwright runner)

**Database (PostgreSQL):**
- `recording_sessions` — Recording metadata
- `recording_events` — Raw events (JSONB)
- `test_recordings` — Processed tests (JSONB)
- `test_library` — Reusable tests
- `codebases` — Codebase metadata
- `test_executions` — Execution history

**AI (LLM):**
- `prompts/intent_classification.txt` — Classify user intent
- `prompts/nlg.txt` — Natural language generation
- `prompts/selector_optimization.txt` — Selector stability scoring
- `prompts/assertion_inference.txt` — Infer assertions
- `prompts/codebase_analysis.txt` — Analyze code structure
- `prompts/test_plan_generation.txt` — Generate test plan

### Technology Stack

| Layer | Technology | Reason |
|-------|-----------|--------|
| Frontend | Next.js 16 + React 19 | Modern, SSR, good DX |
| Backend | Go + Chi | Fast, type-safe, our expertise |
| Database | PostgreSQL 16 | JSONB, mature, reliable |
| Cache/Queue | Redis 7 | Fast, pub/sub, job queue |
| AI | OpenAI, Anthropic, Llama | Best-in-class LLMs |
| Browser Automation | Playwright | Modern, multi-browser, good API |
| Infrastructure | Docker + Kubernetes | Scalable, portable |
| Observability | Prometheus + Jaeger | OpenTelemetry, industry standard |

---

## 6. Development Phases

### Phase 1: Codebase Analysis (Months 1-3)

#### Sprint 1-2 (Weeks 1-4): MVP Parser

**Week 1-2: JavaScript parser**
- [ ] Setup parser infrastructure (tree-sitter or Babel)
- [ ] Parse Express.js routes (`app.get('/users', handler)`)
- [ ] Extract route params, middleware
- [ ] Parse models (Mongoose schemas)
- [ ] Test on 5 real Express.js repos
- [ ] Accuracy target: ≥90%

**Week 3-4: Go parser**
- [ ] Parse Chi routes (`r.Get("/users", handler)`)
- [ ] Parse Gin routes (`router.GET("/users", handler)`)
- [ ] Extract struct tags, validation rules
- [ ] Parse GORM models
- [ ] Test on 5 real Go repos
- [ ] Accuracy target: ≥90%

**Deliverables:**
- ✅ JS parser (Express.js)
- ✅ Go parser (Chi + Gin)
- ✅ Parser accuracy report

#### Sprint 3-4 (Weeks 5-8): PHP + Python parsers

**Week 5-6: PHP (Laravel) parser**
- [ ] Parse Laravel routes (`Route::get('/users', [Controller::class, 'index'])`)
- [ ] Parse migrations
- [ ] Extract validation rules
- [ ] Parse Eloquent models
- [ ] Test on 5 real Laravel repos
- [ ] Accuracy target: ≥90%

**Week 7-8: Python parser**
- [ ] Parse FastAPI routes (`@app.get("/users")`)
- [ ] Parse Flask routes (`@app.route("/users")`)
- [ ] Parse Pydantic models
- [ ] Extract validation rules
- [ ] Test on 5 real Python repos
- [ ] Accuracy target: ≥80%

**Deliverables:**
- ✅ PHP parser (Laravel)
- ✅ Python parser (FastAPI + Flask)
- ✅ Multi-language parser report

#### Sprint 5-6 (Weeks 9-12): AI synthesis + GitHub integration

**Week 9-10: AI synthesis**
- [ ] Feed parsed code to LLM
- [ ] Generate test plan (scenarios, steps, assertions)
- [ ] Generate executable tests (Playwright)
- [ ] Generate test data examples
- [ ] Test on 20 real repos (5 per language)
- [ ] Quality target: ≥70% test plan relevance (human rating ≥4/5)

**Week 11-12: GitHub integration**
- [ ] GitHub OAuth flow
- [ ] Clone private repos
- [ ] Sandboxed environment (Docker)
- [ ] Security scan (malware detection)
- [ ] UI: connect repo, view analysis, edit test plan
- [ ] Beta launch (50 users)

**Deliverables:**
- ✅ AI synthesis pipeline
- ✅ GitHub integration
- ✅ Beta release (50 users)

**Success Criteria:**
- 4 parsers (JS, Go, PHP, Python) with ≥90% accuracy
- 50 beta users
- 500+ tests generated
- 70% generated test pass rate
- NPS > 40

### Phase 2: Record & Playback (Months 4-6)

#### Sprint 7-8 (Weeks 13-16): Chrome extension + backend

**Week 13-14: Chrome extension**
- [ ] Setup Chrome extension (Manifest V3)
- [ ] Capture DOM events (click, input, navigate)
- [ ] Selector generation (data-testid > aria-label > ID > CSS)
- [ ] WebSocket communication with backend
- [ ] Basic UI (start/stop recording)

**Week 15-16: Backend**
- [ ] Recorder service (receive events, store in DB)
- [ ] WebSocket server (real-time streaming)
- [ ] Event storage (PostgreSQL JSONB)
- [ ] Basic frontend (recorder page, test list)

**Deliverables:**
- ✅ Chrome extension
- ✅ Recorder backend
- ✅ End-to-end recording flow

#### Sprint 9-10 (Weeks 17-20): AI processing

**Week 17-18: AI engine**
- [ ] Intent classification (LLM prompt)
- [ ] Natural language generation (LLM prompt)
- [ ] Selector optimization (stability scoring algorithm)
- [ ] Assertion inference (LLM prompt)
- [ ] AI processing pipeline (async job queue)

**Week 19-20: Structured test generation**
- [ ] JSON test structure assembly
- [ ] Markdown documentation generation
- [ ] Test storage (test_recordings table)
- [ ] Test viewer UI (display steps, assertions)

**Deliverables:**
- ✅ AI processing pipeline
- ✅ Structured test format (JSON + Markdown)
- ✅ Test viewer UI

#### Sprint 11-12 (Weeks 21-24): Export + library

**Week 21-22: Export service**
- [ ] Playwright exporter (JavaScript/TypeScript)
- [ ] Cypress exporter (JavaScript)
- [ ] Selenium exporter (Python)
- [ ] Markdown exporter
- [ ] Export API endpoint
- [ ] Download functionality

**Week 23-24: Test library**
- [ ] Test list page (search, filter, paginate)
- [ ] Test detail page (view, edit, export)
- [ ] Tag management
- [ ] Basic team sharing (view-only)
- [ ] Public launch

**Deliverables:**
- ✅ Multi-framework export
- ✅ Test library
- ✅ Public launch

**Success Criteria:**
- 200 paying customers
- 5,000+ tests recorded
- 80% test pass rate
- $50k MRR

### Phase 3: Continuous Sync (Months 7-9)

#### Sprint 13-14 (Weeks 25-28): GitHub webhook + diff analysis

**Week 25-26: GitHub webhook**
- [ ] Setup webhook (listen to push events)
- [ ] Extract diff (changed files, lines)
- [ ] Map changes to routes/models
- [ ] Store webhook events

**Week 27-28: Diff analysis**
- [ ] Parse changed files
- [ ] Determine impact (which tests affected)
- [ ] Regenerate affected tests only
- [ ] Store regenerated tests

**Deliverables:**
- ✅ GitHub webhook integration
- ✅ Diff analysis engine
- ✅ Auto-regeneration pipeline

#### Sprint 15-16 (Weeks 29-32): Notification + drift detection

**Week 29-30: Notification system**
- [ ] Slack webhook
- [ ] Teams webhook
- [ ] Email notification
- [ ] In-app notification
- [ ] Notification preferences

**Week 31-32: Drift detection**
- [ ] Compare code vs tests (out of sync detection)
- [ ] Alert when drift detected
- [ ] Suggest test updates
- [ ] Auto-update tests (optional)

**Deliverables:**
- ✅ Notification system
- ✅ Drift detection
- ✅ Auto-update pipeline

**Success Criteria:**
- 500 paying customers
- 10,000+ tests in library
- 90% test pass rate
- $150k MRR

### Phase 4: Enterprise Features (Months 10-12)

#### Sprint 17-18 (Weeks 33-36): SSO + RBAC

**Week 33-34: SSO**
- [ ] SAML integration
- [ ] OIDC integration
- [ ] OAuth2 (GitHub, Google, Microsoft)
- [ ] SSO settings UI

**Week 35-36: RBAC**
- [ ] Role definitions (admin, developer, viewer)
- [ ] Permission system
- [ ] Role assignment UI
- [ ] Audit log (who did what, when)

**Deliverables:**
- ✅ SSO (SAML, OIDC)
- ✅ RBAC
- ✅ Audit log

#### Sprint 19-20 (Weeks 37-40): On-premise + compliance

**Week 37-38: On-premise deployment**
- [ ] Helm chart (Kubernetes)
- [ ] Docker Compose (simplified)
- [ ] Installation guide
- [ ] Configuration guide
- [ ] Support documentation

**Week 39-40: Compliance**
- [ ] SOC 2 Type I audit
- [ ] Security hardening
- [ ] Data encryption (at rest, in transit)
- [ ] Access controls
- [ ] Compliance documentation

**Deliverables:**
- ✅ On-premise deployment option
- ✅ SOC 2 Type I certification
- ✅ Compliance documentation

#### Sprint 21-22 (Weeks 41-44): Advanced analytics + SLA

**Week 41-42: Advanced analytics**
- [ ] Coverage gaps (what's not tested)
- [ ] Flaky tests (inconsistent results)
- [ ] Trends (pass rate over time)
- [ ] Insights (AI-powered recommendations)
- [ ] Dashboard UI

**Week 43-44: SLA + support**
- [ ] SLA guarantees (99.9% uptime)
- [ ] Priority support (24h response)
- [ ] Dedicated account manager
- [ ] Custom integrations
- [ ] Enterprise pricing ($10k+/month)

**Deliverables:**
- ✅ Advanced analytics
- ✅ SLA guarantees
- ✅ Enterprise support

**Success Criteria:**
- 1,000 paying customers
- 50,000+ tests in library
- 95% test pass rate
- $500k MRR
- 10 enterprise customers ($10k+/month)

---

## 7. Decision Log

### Decision 1: Codebase Analysis as Core Feature

**Date:** 2026-07-30  
**Status:** Approved  
**Context:** We initially planned Record & Playback as core feature, but realized Codebase Analysis is our unique selling point (no competitor offers "upload code → get tests").  
**Decision:** Prioritize Codebase Analysis (Phase 1), Record & Playback becomes Phase 2.  
**Rationale:** Codebase Analysis differentiates us from Katalon, TestSprite, and all competitors. Record & Playback is table stakes (Katalon already has it).  
**Alternatives considered:**
- Record & Playback first (rejected: not unique)
- Both in parallel (rejected: too much scope)

### Decision 2: Multi-language Support (4 languages for MVP)

**Date:** 2026-07-30  
**Status:** Approved  
**Context:** Need to support multiple programming languages to reach broader market.  
**Decision:** Support JS/TS, Go, PHP (Laravel), Python for MVP.  
**Rationale:**
- JS/TS: Largest market (Express.js, Next.js)
- Go: Our stack, type-safe, high-value market (startups, fintech)
- PHP (Laravel): Massive SME market, clean conventions
- Python: Data science community, FastAPI growing

**Alternatives considered:**
- JS only (rejected: too narrow)
- All languages (rejected: too much scope)
- Java + C# (rejected: enterprise focus, not our target market)

### Decision 3: OpenTelemetry for Observability

**Date:** 2026-07-29  
**Status:** Approved  
**Context:** Need distributed tracing for debugging production issues.  
**Decision:** Use OpenTelemetry (Jaeger for trace visualization).  
**Rationale:** OpenTelemetry is industry standard, vendor-neutral, well-supported. Jaeger is open-source, easy to deploy.  
**Alternatives considered:**
- Datadog (rejected: expensive, vendor lock-in)
- New Relic (rejected: expensive)
- Custom solution (rejected: too much work)

### Decision 4: LLM Resilience with Circuit Breaker

**Date:** 2026-07-28  
**Status:** Approved  
**Context:** LLM APIs have transient failures (rate limits, timeouts). Need retry logic.  
**Decision:** Implement exponential backoff (1s → 2s → 4s, max 8s) + circuit breaker (5 failures → open, 30s recovery).  
**Rationale:** Exponential backoff handles rate limits, circuit breaker prevents cascading failures.  
**Alternatives considered:**
- Fixed backoff (rejected: not adaptive)
- No circuit breaker (rejected: risk of cascading failures)

### Decision 5: Docker Multi-stage Hardened Build

**Date:** 2026-07-27  
**Status:** Approved  
**Context:** Production deployment needs secure, minimal Docker image.  
**Decision:** Multi-stage build (golang:1.24-alpine for build, debian:bookworm-slim for runtime), non-root user.  
**Rationale:** Multi-stage reduces image size (2GB → 400MB), non-root user improves security.  
**Alternatives considered:**
- Single-stage build (rejected: too large)
- Root user (rejected: security risk)

### Decision 6: PostgreSQL JSONB for Test Storage

**Date:** 2026-07-26  
**Status:** Approved  
**Context:** Need flexible storage for test steps (diverse structures).  
**Decision:** Use PostgreSQL JSONB (JSON with indexing).  
**Rationale:** JSONB is flexible (store any structure), queryable (filter by tags, intent), performant (indexing).  
**Alternatives considered:**
- Separate tables (rejected: too rigid)
- MongoDB (rejected: less mature, harder to operate)

### Decision 7: Playwright as Primary Browser Automation

**Date:** 2026-07-25  
**Status:** Approved  
**Context:** Need browser automation for test execution.  
**Decision:** Use Playwright (Chromium, Firefox, WebKit).  
**Rationale:** Playwright is modern, multi-browser, good API, active development.  
**Alternatives considered:**
- Selenium (rejected: older, slower)
- Puppeteer (rejected: Chromium only)
- Cypress (rejected: not headless-friendly)

### Decision 8: Rate Limiting (100 req/min/IP)

**Date:** 2026-07-24  
**Status:** Approved  
**Context:** Need to protect API from abuse.  
**Decision:** Token bucket rate limiter, 100 requests per minute per IP.  
**Rationale:** 100 req/min is generous for legitimate users, blocks abuse.  
**Alternatives considered:**
- No rate limiting (rejected: abuse risk)
- 50 req/min (rejected: too restrictive)
- 200 req/min (rejected: too permissive)

### Decision 9: JUnit XML Export for CI/CD

**Date:** 2026-07-23  
**Status:** Approved  
**Context:** Need to integrate with CI/CD systems (Jenkins, GitLab, GitHub Actions).  
**Decision:** Export test results as JUnit XML.  
**Rationale:** JUnit XML is universal format, supported by all major CI systems.  
**Alternatives considered:**
- Custom format (rejected: not portable)
- JSON only (rejected: not supported by CI systems)

### Decision 10: Freemium Pricing Model

**Date:** 2026-07-22  
**Status:** Tentative (to be validated)  
**Context:** Need pricing model for SaaS offering.  
**Decision:** Freemium (free tier + paid tiers).  
**Rationale:** Free tier drives adoption, paid tiers monetize.  
**Alternatives considered:**
- Free trial only (rejected: lower conversion)
- Paid only (rejected: slower adoption)
- Fully open-source (rejected: harder to monetize)

**Pricing tiers (tentative):**
- Free: 10 recordings/month, public repos only
- Pro ($49/month): Unlimited recordings, private repos, team sharing (5 members)
- Enterprise ($199/month): Everything in Pro, unlimited members, SSO, on-premise option

---

## 8. Open Questions

### Question 1: Record & Playback — Chrome Extension vs Playwright Inspector

**Status:** Open  
**Context:** For Record & Playback feature (Phase 2), need to decide implementation approach.  
**Options:**
- **Option A:** Build Chrome extension (more control, better UX, more work)
- **Option B:** Wrap Playwright Inspector (faster to ship, battle-tested, less control)

**Decision criteria:**
- Development effort (hours)
- Control over UX
- Selector generation quality
- Integration with our backend

**Recommendation:** Try both during spike, pick winner.

### Question 2: LLM Choice for AI Processing

**Status:** Open  
**Context:** Need to choose LLM for intent classification, NLG, assertion inference.  
**Options:**
- **Option A:** OpenAI GPT-4 (best quality, expensive: $0.03/1K tokens)
- **Option B:** Anthropic Claude 3.5 Sonnet (good quality, competitive: $0.003/1K tokens)
- **Option C:** Llama 3.1 405B (open-source, cheap: $0.001/1K tokens, lower quality)
- **Option D:** Hybrid (GPT-4 for complex tasks, smaller model for simple)

**Decision criteria:**
- Accuracy (scoring rubric)
- Cost ($ per user per month)
- Latency (user experience)
- Consistency (reliability)

**Recommendation:** Test all 3 during spike, compare metrics, pick best accuracy/cost ratio.

### Question 3: Self-healing Strategy

**Status:** Open  
**Context:** When UI changes and selectors break, how should we self-heal?  
**Options:**
- **Option A:** Rule-based (if selector fails, try alternatives from predefined list)
- **Option B:** AI-driven (LLM analyzes DOM, suggests new selector)
- **Option C:** Hybrid (try alternatives first, fall back to AI if all fail)

**Decision criteria:**
- Healing success rate
- Latency (how long to heal)
- Cost (AI calls are expensive)

**Recommendation:** Option C (hybrid) — fast for common cases, AI for edge cases.

### Question 4: Test Composition Model

**Status:** Open  
**Context:** How should users reuse recorded tests?  
**Options:**
- **Option A:** Copy-paste (duplicate test, edit manually)
- **Option B:** Reference (test A includes test B as step)
- **Option C:** Inheritance (test A extends test B, override specific steps)

**Decision criteria:**
- Ease of use
- Flexibility
- Maintainability

**Recommendation:** Option B (reference) — simple, powerful, no duplication.

### Question 5: Multi-tenancy Model

**Status:** Open  
**Context:** How to isolate customer data in SaaS offering?  
**Options:**
- **Option A:** Database per tenant (strongest isolation, expensive)
- **Option B:** Schema per tenant (good isolation, moderate cost)
- **Option C:** Row-level security (weakest isolation, cheapest)

**Decision criteria:**
- Security
- Cost
- Complexity

**Recommendation:** Option B (schema per tenant) — good balance of security and cost.

### Question 6: On-premise vs SaaS-only

**Status:** Open  
**Context:** Should we offer on-premise deployment for enterprise customers?  
**Options:**
- **Option A:** SaaS only (simpler, higher margins)
- **Option B:** SaaS + on-premise (more complex, addresses enterprise needs)

**Decision criteria:**
- Market demand
- Engineering effort
- Support burden

**Recommendation:** Option B (SaaS + on-premise) — enterprise customers require on-premise for compliance.

### Question 7: Mobile Testing Scope

**Status:** Open  
**Context:** Should we support mobile app testing (iOS/Android)?  
**Options:**
- **Option A:** Web only (focus, faster to market)
- **Option B:** Web + mobile (broader market, more work)

**Decision criteria:**
- Market demand
- Engineering effort
- Competitive landscape

**Recommendation:** Option A (web only) for Year 1, re-evaluate in Year 2 based on customer demand.

### Question 8: Open-source Strategy

**Status:** Open  
**Context:** Should we open-source the platform?  
**Options:**
- **Option A:** Fully open-source (build community, monetize via cloud hosting)
- **Option B:** Open-core (free basic, paid advanced features)
- **Option C:** Closed-source (proprietary, paid only)

**Decision criteria:**
- Community building
- Monetization
- Competitive advantage

**Recommendation:** Option B (open-core) — build community, monetize advanced features.

### Question 9: Funding Strategy

**Status:** Open  
**Context:** How to fund development?  
**Options:**
- **Option A:** Bootstrap (self-funded, slower growth)
- **Option B:** Seed round ($1-2M, faster growth)
- **Option C:** Series A ($5-10M, aggressive growth)

**Decision criteria:**
- Capital needs
- Growth targets
- Founder control

**Recommendation:** Option A (bootstrap) for MVP, evaluate B/C based on traction after 6 months.

### Question 10: Go-to-market Strategy

**Status:** Open  
**Context:** How to acquire customers?  
**Options:**
- **Option A:** Product Hunt launch (fast, broad reach)
- **Option B:** Content marketing (slow, high-quality leads)
- **Option C:** Direct sales (targeted, high-touch)
- **Option D:** Hybrid (Product Hunt + content + partnerships)

**Decision criteria:**
- Customer acquisition cost (CAC)
- Time to first customers
- Scalability

**Recommendation:** Option D (hybrid) — Product Hunt for initial traction, content for long-term growth, partnerships for enterprise.

---

## 9. Success Metrics

### Phase 1 Metrics (Codebase Analysis)

**Product Metrics:**
- 50 beta users
- 500+ tests generated from codebases
- 70% generated test pass rate
- NPS > 40

**Technical Metrics:**
- Parser accuracy ≥90% (per language)
- AI synthesis quality ≥70% (human rating ≥4/5)
- Test generation latency <60s per codebase
- GitHub integration uptime ≥99%

**Business Metrics:**
- Beta user retention ≥60% (Month 1)
- Beta user satisfaction ≥80% (survey)
- Beta user referrals ≥20% (word-of-mouth)

### Phase 2 Metrics (Record & Playback)

**Product Metrics:**
- 200 paying customers
- 5,000+ tests recorded
- 80% test pass rate
- $50k MRR

**Technical Metrics:**
- Intent classification accuracy ≥85%
- Selector stability match ≥80% (AI vs human judgment)
- Assertion usefulness ≥70% (human rating ≥3/5)
- Self-healing success rate ≥70%

**Business Metrics:**
- Customer acquisition cost (CAC) <$300
- Lifetime value (LTV) >$1,200
- LTV/CAC ratio >3x
- Churn rate <5%/month

### Phase 3 Metrics (Continuous Sync)

**Product Metrics:**
- 500 paying customers
- 10,000+ tests in library
- 90% test pass rate
- $150k MRR

**Technical Metrics:**
- Drift detection accuracy ≥90%
- Auto-regeneration success rate ≥80%
- Notification delivery rate ≥99%
- Webhook processing latency <5s

**Business Metrics:**
- Expansion revenue ≥30% (upsell to existing customers)
- Net revenue retention ≥110% (existing customers spend more)
- Enterprise pipeline ≥20 opportunities

### Phase 4 Metrics (Enterprise Features)

**Product Metrics:**
- 1,000 paying customers
- 50,000+ tests in library
- 95% test pass rate
- $500k MRR
- 10 enterprise customers ($10k+/month)

**Technical Metrics:**
- Uptime ≥99.9%
- On-premise deployment success rate ≥95%
- SOC 2 audit pass (no critical findings)
- SLA compliance ≥99%

**Business Metrics:**
- Enterprise customer acquisition ≥2/quarter
- Enterprise retention ≥95% (annual)
- Enterprise expansion ≥50% (upsell/cross-sell)
- ARR $6M (run rate)

### 3-Year Targets

**Year 1:**
- 1,000 customers
- $500k MRR ($6M ARR)
- 10 enterprise customers
- Series A fundraise ($5-10M)

**Year 2:**
- 10,000 customers
- $5M ARR
- 100 enterprise customers
- Series B fundraise ($20-30M)

**Year 3:**
- 50,000 customers
- $50M ARR
- 500 enterprise customers
- Market leader in AI testing

---

## 10. Resources

### Documentation

**Technical Design:**
- [AI Record & Playback Design](./DESIGN-ai-record-playback.md) — 2,400+ lines, comprehensive technical design for Record & Playback feature
- [Spike Plan](./SPIKE-ai-record-playback.md) — 2-week validation plan for AI quality

**Architecture:**
- [ADR-001: Executor Pattern](./ADR-001.md) — Canonical executor pattern for test runs
- [ADR-002: Steel Browser Integration](./ADR-002.md) — Cloud browser integration strategy
- [ADR-005: LangGraph Sidecar](./ADR-005.md) — Python sidecar for advanced AI features
- [ADR-006: LLM Unification](./ADR-006.md) — Unified LLM transport layer

**Changelog:**
- [CHANGELOG_AI.md](./CHANGELOG_AI.md) — Append-only log of all AI-generated changes
- [TECHNICAL_DEBT.md](./TECHNICAL_DEBT.md) — Known technical debt and resolution status

**Setup:**
- [TRACING_SETUP.md](../TRACING_SETUP.md) — OpenTelemetry distributed tracing setup guide
- [docker-compose.yml](../docker-compose.yml) — Docker Compose configuration (backend, frontend, Jaeger, PostgreSQL, Redis)

### Code

**Backend:**
- `cmd/server/main.go` — HTTP server entry point
- `internal/api/` — REST API handlers (25+ endpoints)
- `internal/agent/` — Test execution engine (Playwright runner, self-healing)
- `internal/ai/` — AI engine (LLM integration, resilience with circuit breaker)
- `internal/appmetrics/` — Prometheus metrics
- `internal/tracing/` — OpenTelemetry distributed tracing
- `internal/junit/` — JUnit XML export

**Frontend:**
- `frontend/src/app/` — Next.js pages (dashboard, runs, tests, monitoring)
- `frontend/src/components/` — Reusable components
- `frontend/src/lib/api.ts` — API client functions

**Infrastructure:**
- `Dockerfile` — Multi-stage hardened build (debian:bookworm-slim, non-root user)
- `docker-compose.yml` — Full stack (backend, frontend, Jaeger, PostgreSQL, Redis, Steel Browser, LangGraph sidecar)
- `.env.example` — Environment variables template

### Tools

**Development:**
- Go 1.24
- Next.js 16 + React 19
- PostgreSQL 16
- Redis 7
- Docker + Docker Compose

**AI:**
- OpenAI GPT-4 (API)
- Anthropic Claude 3.5 Sonnet (API)
- Google Gemini (API)
- DeepSeek (API)
- Mistral (API)
- Groq (API)
- OpenRouter (API)

**Browser Automation:**
- Playwright-Go (Chromium, Firefox, WebKit)

**Observability:**
- Prometheus (metrics)
- Jaeger (distributed tracing)
- Grafana (dashboards, optional)

**CI/CD:**
- GitHub Actions (recommended)
- GitLab CI (supported)
- Jenkins (supported via JUnit XML)

### References

**Competitors:**
- [Katalon](https://www.katalon.com/) — Traditional test automation, $175-667/month/user
- [TestSprite](https://www.testsprite.com/) — AI test generation, SaaS only
- [Testim](https://www.testim.io/) — AI-powered test automation, enterprise focus
- [Mabl](https://www.mabl.com/) — Intelligent test automation, enterprise focus

**Research:**
- [Playwright Documentation](https://playwright.dev/docs/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [PostgreSQL JSONB Documentation](https://www.postgresql.org/docs/current/datatype-json.html)

**Market Analysis:**
- Test automation market size: $20B (2024), growing 15% annually
- AI in testing market size: $2B (2024), growing 30% annually
- Target segments: Startups (10-100 engineers), SMEs (100-1000 employees), Dev-first companies

### Contact

**Engineering Team:** engineering@gotest.ai  
**Product Team:** product@gotest.ai  
**Support:** support@gotest.ai

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-07-30 | Engineering Team | Initial master plan |

---

**End of Master Development Plan**
