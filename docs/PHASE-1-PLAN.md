# Phase 1: Codebase Analysis & Test Generation

**Timeline:** 16 minggu (8 sprint)  
**Goal:** AI agent bisa analyze codebase dari GitHub repo dan generate executable test plan  
**Target:** 50 beta users, 4 bahasa (JS, Go, PHP, Python)

---

## Overview

### Apa yang Kita Build

System yang bisa:
1. Clone GitHub repo (public/private)
2. Detect bahasa & framework
3. Parse code → extract routes, models, handlers
4. Feed ke LLM → generate test plan
5. Convert test plan → executable Playwright code
6. User review & execute tests

### Tech Stack

- **Backend:** Go + Chi router
- **Database:** PostgreSQL (JSONB untuk flexible storage)
- **Parser:** tree-sitter (multi-language AST parsing)
- **AI:** OpenAI GPT-4 / Claude (untuk synthesis)
- **Frontend:** Next.js (existing)
- **Integration:** GitHub API (OAuth + webhooks)

### Constraints

- **Performance:** Analysis harus selesai < 5 menit untuk repo < 100MB
- **Accuracy:** Parser accuracy > 90% (validated dengan test cases)
- **Security:** Isolated execution (Docker sandbox untuk cloning)
- **Cost:** LLM cost < $0.50 per analysis (optimize prompt & caching)

---

## Sprint 1-2: Foundation (Minggu 1-4)

**Goal:** Setup parser infrastructure & basic API

### Features

1. **Parser Core**
   - Common types (Route, Model, Schema, Handler)
   - Parser interface (support multi-language)
   - Parser registry (detect language → select parser)

2. **Tree-sitter Integration**
   - Setup grammar untuk JS, Go, PHP, Python
   - AST traversal utilities
   - Node extraction helpers

3. **Codebase Storage**
   - Store codebase metadata (repo URL, language, analysis status)
   - PostgreSQL schema
   - Basic CRUD API

### Tasks

#### Task 1.1: Define Parser Interface

**File:** `internal/parser/interface.go`

```go
package parser

type Parser interface {
  // Parse codebase dari root directory
  Parse(ctx context.Context, rootDir string) (*Codebase, error)
  
  // SupportedLanguages return list bahasa yang di-support
  SupportedLanguages() []string
}

type Codebase struct {
  Language  string
  Framework string
  Routes    []Route
  Models    []Model
  Handlers  []Handler
}
```

**Success Criteria:**
- Interface defined
- No compilation errors

---

#### Task 1.2: Create Common Types

**File:** `internal/parser/types/route.go`

```go
package types

type Route struct {
  Method     string            // GET, POST, PUT, DELETE
  Path       string            // /users/:id
  Handler    string            // UserController.GetUser
  Middleware []string          // [authenticate, authorize:admin]
  Params     map[string]string // {id: "uuid", name: "string"}
  File       string            // src/routes/users.js
  Line       int               // Line number
}
```

**File:** `internal/parser/types/model.go`

```go
package types

type Model struct {
  Name       string            // User
  Table      string            // users
  Fields     []Field           // [ID, Name, Email]
  Relations  []Relation        // [HasMany:Post, BelongsTo:Company]
  Validation []ValidationRule  // [email: required|email]
  File       string
}

type Field struct {
  Name     string
  Type     string // string, int, bool, uuid
  Required bool
  Unique   bool
  Default  string
}

type Relation struct {
  Type      string // HasOne, HasMany, BelongsTo, ManyToMany
  Model     string // Post, Comment
  ForeignKey string // user_id
}

type ValidationRule struct {
  Field string
  Rule  string // required, email, min:3, max:100
}
```

**Success Criteria:**
- Types defined
- No compilation errors

---

#### Task 1.3-1.6: Setup Tree-sitter Grammars

**Files:**
- `internal/parser/javascript/parser.go`
- `internal/parser/go/parser.go`
- `internal/parser/php/parser.go`
- `internal/parser/python/parser.go`

**Implementation:**

```go
package javascript

import (
  "context"
  "github.com/smacker/go-tree-sitter/javascript"
  "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

type Parser struct {
  parser *sitter.Parser
}

func NewParser() *Parser {
  p := sitter.NewParser()
  p.SetLanguage(javascript.GetLanguage())
  return &Parser{parser: p}
}

func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
  // Walk directory, find .js/.ts files
  // Parse each file ke AST
  // Extract routes, models, handlers
  return &types.Codebase{
    Language:  "javascript",
    Framework: "express", // detect dari package.json
    Routes:    routes,
    Models:    models,
    Handlers:  handlers,
  }, nil
}
```

**Dependencies:**
```bash
go get github.com/smacker/go-tree-sitter/javascript
go get github.com/smacker/go-tree-sitter/golang
go get github.com/smacker/go-tree-sitter/php
go get github.com/smacker/go-tree-sitter/python
```

**Success Criteria:**
- All 4 parsers compile
- Can parse simple example files
- Unit tests pass

---

#### Task 1.7: Implement Parser Registry

**File:** `internal/parser/registry.go`

```go
package parser

import (
  "github.com/go-go-golems/gotest-agent/internal/parser/javascript"
  "github.com/go-go-golems/gotest-agent/internal/parser/go"
  "github.com/go-go-golems/gotest-agent/internal/parser/php"
  "github.com/go-go-golems/gotest-agent/internal/parser/python"
)

type Registry struct {
  parsers map[string]Parser
}

func NewRegistry() *Registry {
  return &Registry{
    parsers: map[string]Parser{
      "javascript": javascript.NewParser(),
      "go":         golang.NewParser(),
      "php":        php.NewParser(),
      "python":     python.NewParser(),
    },
  }
}

func (r *Registry) GetParser(language string) (Parser, error) {
  parser, ok := r.parsers[language]
  if !ok {
    return nil, fmt.Errorf("unsupported language: %s", language)
  }
  return parser, nil
}

func (r *Registry) DetectLanguage(rootDir string) (string, error) {
  // Check for package.json → javascript
  // Check for go.mod → go
  // Check for composer.json → php
  // Check for requirements.txt / pyproject.toml → python
  // Return error jika tidak detected
}
```

**Success Criteria:**
- Registry can detect language dari project structure
- Registry can return correct parser
- Unit tests pass

---

#### Task 1.8-1.9: Database Schema & Store

**File:** `migrations/001_create_codebases.sql`

```sql
CREATE TABLE codebases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  
  -- Repository info
  github_url TEXT NOT NULL,
  github_token TEXT, -- encrypted
  language TEXT NOT NULL,
  framework TEXT,
  
  -- Analysis status
  status TEXT NOT NULL DEFAULT 'pending', -- pending, analyzing, completed, failed
  error_message TEXT,
  
  -- Results (JSONB)
  routes JSONB,
  models JSONB,
  handlers JSONB,
  
  -- Timestamps
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  analyzed_at TIMESTAMP
);

CREATE INDEX idx_codebases_user_id ON codebases(user_id);
CREATE INDEX idx_codebases_status ON codebases(status);
```

**File:** `internal/codebase/store.go`

```go
package codebase

import (
  "context"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
  db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
  return &Store{db: db}
}

type Codebase struct {
  ID           string
  UserID       string
  GithubURL    string
  Language     string
  Framework    string
  Status       string
  Routes       []Route
  Models       []Model
  Handlers     []Handler
  CreatedAt    time.Time
  AnalyzedAt   *time.Time
}

func (s *Store) Create(ctx context.Context, cb *Codebase) error {
  // INSERT INTO codebases
}

func (s *Store) Get(ctx context.Context, id string) (*Codebase, error) {
  // SELECT * FROM codebases WHERE id = $1
}

func (s *Store) Update(ctx context.Context, cb *Codebase) error {
  // UPDATE codebases SET ...
}

func (s *Store) List(ctx context.Context, userID string) ([]*Codebase, error) {
  // SELECT * FROM codebases WHERE user_id = $1
}

func (s *Store) Delete(ctx context.Context, id string) error {
  // DELETE FROM codebases WHERE id = $1
}
```

**Success Criteria:**
- Migration runs successfully
- CRUD operations work
- Unit tests pass

---

#### Task 1.10: API Endpoints

**File:** `internal/api/codebase.go`

```go
package api

import (
  "net/http"
  "github.com/go-chi/chi/v5"
)

func (s *Server) registerCodebaseRoutes(r chi.Router) {
  r.Route("/codebases", func(r chi.Router) {
    r.Post("/", s.handleCreateCodebase)
    r.Get("/", s.handleListCodebases)
    r.Get("/{id}", s.handleGetCodebase)
    r.Delete("/{id}", s.handleDeleteCodebase)
    r.Post("/{id}/analyze", s.handleAnalyzeCodebase)
  })
}

func (s *Server) handleCreateCodebase(w http.ResponseWriter, r *http.Request) {
  var req struct {
    GithubURL   string `json:"github_url"`
    GithubToken string `json:"github_token"` // optional, for private repos
  }
  
  // Validate URL
  // Create codebase record
  // Queue analysis job (async)
  // Return 202 Accepted
}

func (s *Server) handleAnalyzeCodebase(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  
  // Get codebase
  // Clone repo ke temp directory
  // Detect language
  // Parse codebase
  // Update codebase with results
  // Trigger test generation (next sprint)
}
```

**Success Criteria:**
- Endpoints work (curl test)
- Can create & list codebases
- Analysis runs (basic parsing)

---

#### Task 1.11-1.20: Unit Tests

**Files:**
- `internal/parser/registry_test.go`
- `internal/parser/javascript/parser_test.go`
- `internal/parser/go/parser_test.go`
- `internal/parser/php/parser_test.go`
- `internal/parser/python/parser_test.go`
- `internal/codebase/store_test.go`
- `internal/api/codebase_test.go`

**Test Cases:**
```go
func TestParserRegistry_DetectLanguage(t *testing.T) {
  tests := []struct {
    name     string
    files    map[string]string // filename → content
    expected string
  }{
    {
      name: "detect javascript from package.json",
      files: map[string]string{
        "package.json": `{"name": "test"}`,
      },
      expected: "javascript",
    },
    {
      name: "detect go from go.mod",
      files: map[string]string{
        "go.mod": "module test",
      },
      expected: "go",
    },
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      // Create temp dir with files
      // Call DetectLanguage
      // Assert result
    })
  }
}
```

**Success Criteria:**
- All tests pass
- Code coverage > 80%

---

## Sprint 3-4: JavaScript Parser (Minggu 5-8)

**Goal:** Parse Express.js routes, models, handlers dengan accuracy > 90%

### Features

1. **Express Route Extraction**
   - Parse `app.get()`, `router.post()`, dll
   - Extract middleware chains
   - Handle route parameters (`:id`, `:slug`)
   - Support route grouping

2. **Mongoose Model Extraction**
   - Parse schema definitions
   - Extract field types & validation
   - Detect relationships (`ref`)

3. **Controller Function Analysis**
   - Identify handler functions
   - Extract request/response types

### Tasks

#### Task 2.1: Implement Express Route Parser

**File:** `internal/parser/javascript/express.go`

```go
package javascript

import (
  "github.com/smacker/go-tree-sitter"
  "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func (p *Parser) extractExpressRoutes(node *sitter.Node, content []byte) []types.Route {
  var routes []types.Route
  
  // Walk AST, find call expressions
  // Match patterns:
  //   app.get('/path', handler)
  //   router.post('/path', middleware1, middleware2, handler)
  //   app.use('/prefix', router)
  
  return routes
}

func extractRouteFromCall(node *sitter.Node, content []byte) types.Route {
  // Extract method (get, post, put, delete)
  // Extract path (/users/:id)
  // Extract handler name
  // Extract middleware
  return types.Route{
    Method:     method,
    Path:       path,
    Handler:    handler,
    Middleware: middleware,
  }
}
```

**Test Cases:**

```go
// testdata/express/basic.js
const express = require('express');
const app = express();

app.get('/users', listUsers);
app.post('/users', authenticate, createUser);
app.get('/users/:id', getUser);
app.put('/users/:id', authenticate, updateUser);
app.delete('/users/:id', authenticate, authorize('admin'), deleteUser);
```

**Expected Output:**
```go
[]types.Route{
  {Method: "GET", Path: "/users", Handler: "listUsers", Middleware: []},
  {Method: "POST", Path: "/users", Handler: "createUser", Middleware: ["authenticate"]},
  {Method: "GET", Path: "/users/:id", Handler: "getUser", Middleware: []},
  {Method: "PUT", Path: "/users/:id", Handler: "updateUser", Middleware: ["authenticate"]},
  {Method: "DELETE", Path: "/users/:id", Handler: "deleteUser", Middleware: ["authenticate", "authorize"]},
}
```

**Success Criteria:**
- Parse 50+ Express.js examples dengan accuracy > 90%
- Handle edge cases (nested routes, dynamic middleware)
- Unit tests pass

---

#### Task 2.2: Parse Route Parameters

**Implementation:**

```go
func extractRouteParams(path string) map[string]string {
  params := make(map[string]string)
  
  // Match :param patterns
  // /users/:id → {id: "string"}
  // /posts/:postId/comments/:commentId → {postId: "string", commentId: "string"}
  
  return params
}
```

**Test Cases:**
- `/users/:id` → `{id: "string"}`
- `/users/:userId/posts/:postId` → `{userId: "string", postId: "string"}`
- `/files/:path*` → `{path: "string"}` (wildcard)

**Success Criteria:**
- Extract all parameters correctly
- Handle wildcards & optional params

---

#### Task 2.3-2.4: Middleware & Route Grouping

**Implementation:**

```go
func extractMiddleware(node *sitter.Node, content []byte) []string {
  // Parse middleware function names
  // Handle inline middleware (anonymous functions)
  // Return list of middleware names
}

func extractRouteGrouping(node *sitter.Node, content []byte) []types.Route {
  // Parse express.Router() usage
  // Handle app.use('/prefix', router)
  // Prefix all routes dengan /prefix
}
```

**Test Cases:**

```javascript
// Middleware chain
app.post('/users', authenticate, validate, createUser);
// Expected: [authenticate, validate]

// Route grouping
const userRouter = express.Router();
userRouter.get('/', listUsers);
userRouter.post('/', createUser);
app.use('/api/users', userRouter);
// Expected: /api/users, /api/users (POST)
```

**Success Criteria:**
- Extract middleware correctly
- Handle route grouping dengan prefix

---

#### Task 2.5-2.8: Mongoose Model Extraction

**File:** `internal/parser/javascript/mongoose.go`

```go
package javascript

import (
  "github.com/smacker/go-tree-sitter"
  "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func (p *Parser) extractMongooseModels(node *sitter.Node, content []byte) []types.Model {
  var models []types.Model
  
  // Find mongoose.Schema() calls
  // Parse schema definition
  // Extract fields, types, validation
  // Detect relationships (ref)
  
  return models
}

func extractModelFromSchema(node *sitter.Node, content []byte) types.Model {
  // Parse schema object
  // Extract field definitions
  // Map mongoose types → our types
  // Extract validation rules
  
  return types.Model{
    Name:       modelName,
    Fields:     fields,
    Relations:  relations,
    Validation: validation,
  }
}
```

**Test Cases:**

```javascript
// testdata/mongoose/user.js
const userSchema = new mongoose.Schema({
  name: { type: String, required: true },
  email: { type: String, required: true, unique: true },
  age: { type: Number, min: 18, max: 120 },
  company: { type: mongoose.Schema.Types.ObjectId, ref: 'Company' },
  posts: [{ type: mongoose.Schema.Types.ObjectId, ref: 'Post' }],
}, { timestamps: true });

module.exports = mongoose.model('User', userSchema);
```

**Expected Output:**
```go
types.Model{
  Name:  "User",
  Table: "users",
  Fields: []types.Field{
    {Name: "name", Type: "string", Required: true},
    {Name: "email", Type: "string", Required: true, Unique: true},
    {Name: "age", Type: "number"},
    {Name: "company", Type: "objectId"},
    {Name: "posts", Type: "array"},
  },
  Relations: []types.Relation{
    {Type: "BelongsTo", Model: "Company", ForeignKey: "company"},
    {Type: "HasMany", Model: "Post", ForeignKey: "posts"},
  },
  Validation: []types.ValidationRule{
    {Field: "name", Rule: "required"},
    {Field: "email", Rule: "required|email"},
    {Field: "age", Rule: "min:18|max:120"},
  },
}
```

**Success Criteria:**
- Parse 30+ Mongoose examples dengan accuracy > 90%
- Handle nested schemas, arrays, relationships
- Extract validation rules correctly

---

#### Task 2.9-2.13: Controller & Validation

**Implementation:**

```go
func (p *Parser) extractControllers(node *sitter.Node, content []byte) []types.Handler {
  // Find exported functions
  // Identify request/response patterns (req.body, res.json)
  // Extract function signature
  
  return handlers
}

func (p *Parser) extractJoiValidation(node *sitter.Node, content []byte) []types.ValidationRule {
  // Parse Joi schema definitions
  // Extract validation rules
  // Map Joi rules → our validation format
}
```

**Success Criteria:**
- Identify controller functions correctly
- Extract Joi/Yup validation rules
- Unit tests pass

---

#### Task 2.14-2.30: Comprehensive Tests

**Test Dataset:**
- 50 Express.js route examples (simple → complex)
- 30 Mongoose model examples
- 20 Joi validation examples
- 10 full Express apps (routes + models + controllers)

**Success Criteria:**
- Overall accuracy > 90%
- No critical bugs (routes/models missed)
- All tests pass

---

## Sprint 5-7: Go, PHP, Python Parsers (Minggu 9-14)

**Structure mirip Sprint 3-4, tapi untuk:**

### Sprint 5: Go Parser

**Features:**
- Chi/Gin/Echo route extraction
- Struct tag parsing (json, db, validate)
- Handler function analysis

**Test Dataset:**
- 30 Chi route examples
- 20 Gin route examples
- 20 struct tag examples

### Sprint 6: PHP Parser (Laravel)

**Features:**
- Laravel route extraction (Route::get, Route::resource)
- Eloquent model parsing
- Migration parsing

**Test Dataset:**
- 30 Laravel route examples
- 20 Eloquent model examples
- 20 migration examples

### Sprint 7: Python Parser

**Features:**
- FastAPI/Flask route extraction
- Pydantic model parsing
- SQLAlchemy model parsing

**Test Dataset:**
- 30 FastAPI route examples
- 20 Flask route examples
- 20 Pydantic model examples

**Success Criteria (semua parser):**
- Accuracy > 90%
- Unit tests pass
- Integration tests pass

---

## Sprint 8: AI Synthesis (Minggu 15-16)

**Goal:** Convert parsed code → test plan → Playwright code

### Features

1. **Test Plan Generation**
   - Feed routes/models/handlers ke LLM
   - Generate test scenarios (happy path, edge cases, errors)
   - Generate test steps

2. **Playwright Code Generation**
   - Convert test plan → executable code
   - Generate assertions
   - Generate test data

3. **Confidence Scoring**
   - Score setiap generated test
   - Flag low-confidence tests

### Tasks

#### Task 6.1-6.5: Test Plan Generation

**File:** `internal/ai/test_plan.go`

```go
package ai

import (
  "context"
  "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

type TestPlanGenerator struct {
  client *Client
}

func (g *TestPlanGenerator) Generate(ctx context.Context, codebase *types.Codebase) (*TestPlan, error) {
  prompt := g.buildPrompt(codebase)
  
  response, err := g.client.Chat(ctx, prompt)
  if err != nil {
    return nil, err
  }
  
  // Parse JSON response → TestPlan
  plan, err := parseTestPlan(response)
  if err != nil {
    return nil, err
  }
  
  // Score confidence
  g.scoreConfidence(plan)
  
  return plan, nil
}

func (g *TestPlanGenerator) buildPrompt(codebase *types.Codebase) string {
  return fmt.Sprintf(`
You are a test automation expert. Generate a comprehensive test plan for this codebase.

CODEBASE:
Language: %s
Framework: %s

ROUTES:
%s

MODELS:
%s

Generate test scenarios covering:
1. Happy path (valid inputs)
2. Edge cases (boundary values, empty inputs)
3. Error cases (invalid inputs, unauthorized access)

For each scenario, provide:
- Scenario name
- Test steps (navigate, fill, click, assert)
- Expected outcomes

Output as JSON:
{
  "scenarios": [
    {
      "name": "Create user with valid data",
      "type": "happy_path",
      "steps": [...],
      "assertions": [...]
    }
  ]
}
`, codebase.Language, codebase.Framework, routesJSON, modelsJSON)
}
```

**Success Criteria:**
- Generate test plans untuk 20 codebases
- Plans cover happy path, edge cases, errors
- JSON parsing works
- Confidence scoring works

---

#### Task 6.6-6.9: Playwright Code Generation

**File:** `internal/ai/playwright.go`

```go
package ai

import (
  "context"
)

type PlaywrightGenerator struct {
  client *Client
}

func (g *PlaywrightGenerator) Generate(ctx context.Context, plan *TestPlan) (string, error) {
  prompt := g.buildPrompt(plan)
  
  response, err := g.client.Chat(ctx, prompt)
  if err != nil {
    return "", err
  }
  
  // Response is Playwright code (TypeScript)
  return response, nil
}

func (g *PlaywrightGenerator) buildPrompt(plan *TestPlan) string {
  return fmt.Sprintf(`
Convert this test plan into Playwright code (TypeScript).

TEST PLAN:
%s

Generate:
1. Test file structure
2. Test cases dengan proper assertions
3. Test data (valid & invalid inputs)
4. Before/after hooks jika perlu

Output executable Playwright code.
`, planJSON)
}
```

**Success Criteria:**
- Generate valid Playwright code
- Code runs without syntax errors
- Assertions work correctly

---

## Sprint 9-10: GitHub Integration (Minggu 17-20)

**Goal:** User bisa connect GitHub repo & auto-analyze

### Features

1. **GitHub OAuth**
   - Login with GitHub
   - Store access token (encrypted)

2. **Repo Cloning**
   - Clone public/private repos
   - Shallow clone (performance)

3. **Webhook Integration**
   - Listen to push events
   - Auto-reanalyze on changes

### Tasks

#### Task 7.1-7.3: GitHub OAuth

**File:** `internal/github/oauth.go`

```go
package github

import (
  "golang.org/x/oauth2"
)

type OAuth struct {
  config *oauth2.Config
}

func NewOAuth(clientID, clientSecret string) *OAuth {
  return &OAuth{
    config: &oauth2.Config{
      ClientID:     clientID,
      ClientSecret: clientSecret,
      Endpoint:     github.Endpoint,
      Scopes:       []string{"repo"},
    },
  }
}

func (o *OAuth) GetAuthURL(state string) string {
  return o.config.AuthCodeURL(state)
}

func (o *OAuth) Exchange(code string) (*oauth2.Token, error) {
  return o.config.Exchange(context.Background(), code)
}
```

**File:** `internal/api/github.go`

```go
func (s *Server) handleGithubLogin(w http.ResponseWriter, r *http.Request) {
  state := generateRandomState()
  url := s.githubOAuth.GetAuthURL(state)
  http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
  code := r.URL.Query().Get("code")
  token, err := s.githubOAuth.Exchange(code)
  
  // Store token (encrypted) di database
  // Redirect ke dashboard
}
```

**Success Criteria:**
- OAuth flow works
- Token stored securely
- Can list user repos

---

#### Task 7.4-7.6: Repo Cloning

**File:** `internal/github/clone.go`

```go
package github

import (
  "os/exec"
)

type Cloner struct {
  tempDir string
}

func (c *Cloner) Clone(url, token string) (string, error) {
  // Create temp directory
  dir, err := os.MkdirTemp("", "gotest-*")
  if err != nil {
    return "", err
  }
  
  // Build clone URL dengan token
  cloneURL := c.buildCloneURL(url, token)
  
  // Shallow clone (depth 1) untuk speed
  cmd := exec.Command("git", "clone", "--depth", "1", cloneURL, dir)
  if err := cmd.Run(); err != nil {
    return "", err
  }
  
  return dir, nil
}

func (c *Cloner) Cleanup(dir string) error {
  return os.RemoveAll(dir)
}
```

**Success Criteria:**
- Clone public repos
- Clone private repos dengan token
- Cleanup temp directories
- Shallow clone works (performance)

---

#### Task 7.7-7.10: Webhook Integration

**File:** `internal/github/webhook.go`

```go
package github

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
)

type Webhook struct {
  secret string
}

func (w *Webhook) VerifySignature(payload []byte, signature string) bool {
  mac := hmac.New(sha256.New, []byte(w.secret))
  mac.Write(payload)
  expectedMAC := hex.EncodeToString(mac.Sum(nil))
  return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

func (w *Webhook) ParsePushEvent(payload []byte) (*PushEvent, error) {
  var event PushEvent
  if err := json.Unmarshal(payload, &event); err != nil {
    return nil, err
  }
  return &event, nil
}
```

**File:** `internal/api/webhook.go`

```go
func (s *Server) handleGithubWebhook(w http.ResponseWriter, r *http.Request) {
  // Read payload
  // Verify signature
  // Parse event
  // Jika push event:
  //   - Find codebase dengan repo URL
  //   - Queue reanalysis job
}
```

**Success Criteria:**
- Webhook signature verification works
- Parse push events correctly
- Queue reanalysis on push

---

## Sprint 11-12: UI/UX (Minggu 21-24)

**Goal:** User-friendly interface untuk manage codebases & view results

### Features

1. **Codebase Dashboard**
   - List semua codebases
   - Show analysis status
   - Quick actions (analyze, delete)

2. **Analysis Results**
   - Show extracted routes, models, handlers
   - Show generated test plan
   - Edit test plan (human review)

3. **Test Execution**
   - Run generated tests
   - Show results (pass/fail)
   - Video recording & screenshots

### Tasks

#### Task 8.1-8.4: Codebase Dashboard

**File:** `frontend/src/app/codebases/page.tsx`

```tsx
export default function CodebasesPage() {
  const [codebases, setCodebases] = useState([]);
  
  useEffect(() => {
    fetch('/api/v1/codebases')
      .then(res => res.json())
      .then(data => setCodebases(data));
  }, []);
  
  return (
    <div>
      <h1>My Codebases</h1>
      <button>Add Codebase</button>
      
      <table>
        <thead>
          <tr>
            <th>Repository</th>
            <th>Language</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {codebases.map(cb => (
            <tr key={cb.id}>
              <td>{cb.github_url}</td>
              <td>{cb.language}</td>
              <td><StatusBadge status={cb.status} /></td>
              <td>
                <button onClick={() => analyzeCodebase(cb.id)}>Analyze</button>
                <button onClick={() => deleteCodebase(cb.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

**Success Criteria:**
- List codebases works
- Status badge shows correctly
- Actions work (analyze, delete)

---

#### Task 8.5-8.9: Analysis Results View

**File:** `frontend/src/app/codebases/[id]/page.tsx`

```tsx
export default function CodebaseDetailPage() {
  const [codebase, setCodebase] = useState(null);
  
  return (
    <div>
      <h1>{codebase.github_url}</h1>
      
      <Tabs>
        <Tab label="Routes">
          <RoutesTable routes={codebase.routes} />
        </Tab>
        <Tab label="Models">
          <ModelsTable models={codebase.models} />
        </Tab>
        <Tab label="Test Plan">
          <TestPlanEditor plan={codebase.test_plan} />
        </Tab>
        <Tab label="Generated Code">
          <CodeViewer code={codebase.playwright_code} />
        </Tab>
      </Tabs>
      
      <button onClick={runTests}>Run Tests</button>
    </div>
  );
}
```

**Success Criteria:**
- Show routes, models, handlers
- Edit test plan works
- View generated code works

---

#### Task 8.10-8.14: Test Execution

**File:** `frontend/src/app/codebases/[id]/tests/page.tsx`

```tsx
export default function TestExecutionPage() {
  const [results, setResults] = useState(null);
  
  return (
    <div>
      <h1>Test Results</h1>
      
      <div className="summary">
        <span>Passed: {results.passed}</span>
        <span>Failed: {results.failed}</span>
        <span>Total: {results.total}</span>
      </div>
      
      <div className="video">
        <video src={results.video_url} controls />
      </div>
      
      <div className="screenshots">
        {results.screenshots.map(ss => (
          <img key={ss.id} src={ss.url} alt={ss.name} />
        ))}
      </div>
      
      <table>
        <thead>
          <tr>
            <th>Test</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Error</th>
          </tr>
        </thead>
        <tbody>
          {results.tests.map(test => (
            <tr key={test.id}>
              <td>{test.name}</td>
              <td><StatusBadge status={test.status} /></td>
              <td>{test.duration}ms</td>
              <td>{test.error}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

**Success Criteria:**
- Show test results (pass/fail)
- Video playback works
- Screenshots display correctly

---

## Sprint 13-14: Testing & Beta Prep (Minggu 25-28)

**Goal:** Ensure quality & prepare untuk beta launch

### Features

1. **End-to-End Tests**
   - Full workflow tests
   - Test dengan 10 real GitHub repos

2. **Performance Optimization**
   - Optimize parser speed
   - Optimize AI generation
   - Caching layer

3. **Documentation**
   - User guide
   - API documentation
   - Onboarding flow

### Tasks

#### Task 9.1-9.2: End-to-End Tests

**File:** `tests/e2e/codebase_analysis_test.go`

```go
func TestCodebaseAnalysis_FullWorkflow(t *testing.T) {
  // 1. Create codebase (GitHub URL)
  // 2. Wait for analysis to complete
  // 3. Verify routes, models, handlers extracted
  // 4. Verify test plan generated
  // 5. Verify Playwright code generated
  // 6. Run generated tests
  // 7. Verify tests execute successfully
}

func TestCodebaseAnalysis_RealRepos(t *testing.T) {
  repos := []string{
    "https://github.com/expressjs/express",
    "https://github.com/gin-gonic/gin",
    "https://github.com/laravel/laravel",
    "https://github.com/tiangolo/fastapi",
    // ... 6 more repos
  }
  
  for _, repo := range repos {
    t.Run(repo, func(t *testing.T) {
      // Analyze repo
      // Verify accuracy > 90%
    })
  }
}
```

**Success Criteria:**
- E2E tests pass
- 10 real repos analyzed successfully
- Accuracy > 90% untuk semua repos

---

#### Task 9.3-9.5: Performance Optimization

**Implementation:**

```go
// Caching layer
type Cache struct {
  redis *redis.Client
}

func (c *Cache) GetCodebaseAnalysis(url string) (*Codebase, error) {
  // Check cache first
  // Jika tidak ada, analyze & cache result
}

// Parallel parsing
func (p *Parser) ParseParallel(ctx context.Context, files []string) (*Codebase, error) {
  var wg sync.WaitGroup
  results := make(chan parseResult, len(files))
  
  for _, file := range files {
    wg.Add(1)
    go func(f string) {
      defer wg.Done()
      result := p.parseFile(f)
      results <- result
    }(file)
  }
  
  wg.Wait()
  close(results)
  
  // Merge results
}
```

**Success Criteria:**
- Analysis < 5 menit untuk repo < 100MB
- Cache hit rate > 50%
- Parallel parsing works

---

#### Task 9.6-9.10: Documentation

**Files:**
- `docs/USER_GUIDE.md`
- `docs/API_REFERENCE.md`
- `docs/ONBOARDING.md`

**Success Criteria:**
- User guide covers all features
- API docs complete
- Onboarding flow clear

---

## Sprint 15-16: Beta Launch (Minggu 29-32)

**Goal:** Deploy ke production & onboard 50 beta users

### Features

1. **Production Deployment**
   - Deploy ke staging → production
   - Monitoring & alerting
   - Backup & disaster recovery

2. **Beta User Onboarding**
   - Invite 50 users
   - Collect feedback
   - Iterate based on feedback

3. **Quick Fixes**
   - Fix critical bugs
   - Improve parser accuracy
   - Improve AI generation quality

### Tasks

#### Task 10.1-10.3: Production Deployment

**Checklist:**
- [ ] Database backup setup
- [ ] Monitoring (Prometheus + Grafana)
- [ ] Alerting (PagerDuty)
- [ ] SSL certificates
- [ ] Domain setup
- [ ] DNS configuration
- [ ] Load balancer
- [ ] Auto-scaling
- [ ] Disaster recovery plan

**Success Criteria:**
- Production deployment successful
- Monitoring & alerting active
- Zero downtime

---

#### Task 10.4-10.10: Beta Launch

**Checklist:**
- [ ] Send beta invitations (50 users)
- [ ] Setup feedback collection (form + email)
- [ ] Monitor usage metrics
- [ ] Collect feedback daily
- [ ] Prioritize improvements
- [ ] Fix critical bugs
- [ ] Improve parser accuracy
- [ ] Improve AI generation quality

**Success Criteria:**
- 50 beta users onboarded
- Feedback collected
- Critical bugs fixed
- Accuracy improved based on feedback

---

## Definition of Done

Phase 1 selesai ketika:

- [x] 4 parser (JS, Go, PHP, Python) accuracy > 90%
- [x] GitHub OAuth works
- [x] Repo cloning works
- [x] AI synthesis generates valid test plans
- [x] Playwright code generation works
- [x] UI/UX complete & user-friendly
- [x] E2E tests pass
- [x] Performance < 5 menit per analysis
- [x] Documentation complete
- [x] Production deployment successful
- [x] 50 beta users onboarded
- [x] Feedback collected & critical bugs fixed

---

## Risks & Mitigations

### Risk 1: Parser Accuracy < 90%

**Mitigation:**
- Extensive test dataset (50+ examples per language)
- Manual review of failed cases
- Iterate on parser logic

### Risk 2: LLM Cost Too High

**Mitigation:**
- Prompt optimization (reduce token count)
- Caching layer (reuse similar analyses)
- Batch processing (multiple codebases per LLM call)

### Risk 3: Performance Too Slow

**Mitigation:**
- Parallel parsing
- Shallow cloning
- Caching layer
- Optimize AST traversal

### Risk 4: GitHub API Rate Limits

**Mitigation:**
- Use GitHub App (higher limits)
- Cache API responses
- Retry dengan exponential backoff

---

## Success Metrics

**Technical Metrics:**
- Parser accuracy > 90% (per language)
- Analysis time < 5 menit
- LLM cost < $0.50 per analysis
- Zero critical bugs

**Product Metrics:**
- 50 beta users onboarded
- User satisfaction > 4/5
- Test generation success rate > 80%

**Business Metrics:**
- 20% of beta users convert to paid (Phase 2)
- Positive feedback dari beta users
- Case studies dari 5 users

---

## Conclusion

Phase 1 adalah foundation untuk GoTest Agent. Dengan menyelesaikan Phase 1, kita punya:

1. **Working product** yang bisa analyze codebase & generate tests
2. **Validated technology** (parser accuracy, AI synthesis, GitHub integration)
3. **Beta users** yang memberikan feedback untuk improvement
4. **Foundation** untuk Phase 2 (Record & Playback)

**Next:** Setelah Phase 1 selesai, kita lanjut ke Phase 2 (Record & Playback) atau iterate based on beta feedback.
