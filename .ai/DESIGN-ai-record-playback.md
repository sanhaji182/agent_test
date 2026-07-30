# Technical Design: AI-Powered Record & Playback with Text Understanding

**Status:** Draft  
**Created:** 2026-07-30  
**Authors:** Engineering Team  
**Reviewers:** TBD  
**Target Ship Date:** Q4 2026 (MVP)

---

## 1. Executive Summary

### 1.1 Problem Statement

Current test automation tools (Katalon, Selenium IDE, Playwright Codegen) suffer from:

1. **Black-box recordings** — Generated code is unreadable and hard to maintain
2. **Brittle selectors** — Tests break when UI changes slightly
3. **No semantic understanding** — Tools don't understand user intent
4. **Vendor lock-in** — Proprietary formats prevent portability
5. **Manual maintenance** — Engineers spend 40% of time fixing broken tests

### 1.2 Proposed Solution

Build an **AI-powered Record & Playback system** that:

1. **Records** user interactions in the browser
2. **Understands** user intent using LLM
3. **Generates** human-readable test documentation (Markdown)
4. **Optimizes** selectors using AI (stable, semantic, maintainable)
5. **Infers** assertions automatically (what should be verified)
6. **Exports** to multiple frameworks (Playwright, Cypress, Selenium)
7. **Enables** test reuse and composition (library system)
8. **Self-heals** when UI changes (auto-update selectors)

### 1.3 Value Proposition

**For QA Teams:**
- Record tests 10x faster than writing manually
- No coding required (natural language descriptions)
- Tests are understandable by non-technical stakeholders

**For Developers:**
- AI-generated code follows best practices
- Stable selectors reduce maintenance burden
- Multi-framework export prevents vendor lock-in

**For Organizations:**
- 80% reduction in test maintenance time
- 50% increase in test coverage
- Faster time-to-market (less time writing tests)

### 1.4 Success Metrics

**MVP (Month 3):**
- 50 beta users
- 500+ tests recorded
- 80% test pass rate (self-healing)
- NPS score > 40

**Post-Launch (Month 6):**
- 500+ paying customers
- 10,000+ tests recorded
- 90% test pass rate
- $50k MRR

---

## 2. Goals & Non-Goals

### 2.1 Goals (MVP)

- [ ] Record user interactions via Chrome extension
- [ ] Generate structured test format (JSON + Markdown)
- [ ] AI-powered selector optimization (stability scoring)
- [ ] AI-powered natural language descriptions
- [ ] AI-powered assertion inference
- [ ] Export to Playwright (JavaScript/TypeScript)
- [ ] Basic test library (save, browse, rerun)
- [ ] Team sharing (view-only for now)

### 2.2 Non-Goals (MVP)

- [ ] Self-healing (Phase 2)
- [ ] Multi-framework export (Phase 2: Cypress, Selenium)
- [ ] Advanced test composition (Phase 3)
- [ ] Mobile app recording (out of scope)
- [ ] Desktop app recording (out of scope)
- [ ] API testing (separate feature)
- [ ] Performance testing (separate feature)

### 2.3 Future Phases

**Phase 2 (Month 6-9):**
- Self-healing selectors
- Cypress & Selenium export
- Parameterization (data-driven tests)
- Test versioning (Git-like)

**Phase 3 (Month 9-12):**
- Test composition (reuse recorded tests)
- Multi-recording test generation (learn from many recordings)
- Advanced analytics (flaky tests, coverage gaps)
- CI/CD integration (GitHub Actions, GitLab CI)

---

## 3. Architecture Overview

### 3.1 High-Level Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Browser Layer                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Chrome Extension (Recorder)                          │  │
│  │  - Capture DOM events (click, input, navigate)        │  │
│  │  - Capture screenshots (optional)                     │  │
│  │  - Capture network requests (optional)                │  │
│  │  - Send events to backend via WebSocket               │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ WebSocket
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Recorder    │  │  AI Engine   │  │  Exporter    │     │
│  │  Service     │  │  Service     │  │  Service     │     │
│  │              │  │              │  │              │     │
│  │  - Receive   │  │  - Intent    │  │  - JSON      │     │
│  │    events    │  │    classify  │  │  - Markdown  │     │
│  │  - Store     │  │  - NLG       │  │  - Playwright│     │
│  │    raw data  │  │  - Selector  │  │  - Cypress   │     │
│  │  - Manage    │  │    optimize  │  │  - Selenium  │     │
│  │    sessions  │  │  - Assert    │  │              │     │
│  │              │  │    infer     │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│           │                │                │               │
│           └────────────────┼────────────────┘               │
│                            │                                │
│                            ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PostgreSQL (Test Storage)                            │  │
│  │  - test_recordings (JSONB steps)                      │  │
│  │  - test_library (reusable tests)                      │  │
│  │  - test_parameters (data-driven)                      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ REST API
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Frontend Layer                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Next.js App                                          │  │
│  │  - Test recorder UI (start/stop recording)            │  │
│  │  - Test viewer (view recorded tests)                  │  │
│  │  - Test editor (edit steps, add assertions)           │  │
│  │  - Test library (browse, search, reuse)               │  │
│  │  - Export dialog (choose format, download)            │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Technology Stack

**Browser Extension:**
- Chrome Extension Manifest V3
- JavaScript/TypeScript
- Chrome DevTools Protocol (for advanced capture)

**Backend:**
- Go (existing stack)
- PostgreSQL (JSONB for flexible test storage)
- Redis (session management, real-time events)
- WebSocket (real-time event streaming)

**AI Engine:**
- OpenAI GPT-4 (intent classification, NLG, assertion inference)
- Custom ML models (selector optimization, stability scoring)
- Tree-sitter (code parsing for multi-language support)

**Frontend:**
- Next.js 16 (existing stack)
- React 19
- Tailwind CSS
- Shadcn/ui (components)

**Infrastructure:**
- Docker (containerization)
- Kubernetes (orchestration, future)
- Prometheus + Grafana (monitoring)
- Jaeger (distributed tracing, already implemented)

---

## 4. Component Breakdown

### 4.1 Chrome Extension (Recorder)

#### 4.1.1 Responsibilities

- Capture user interactions (click, input, navigate, scroll, hover)
- Extract DOM context (selectors, text, attributes)
- Send events to backend in real-time
- Provide visual feedback (recording indicator, step counter)

#### 4.1.2 Event Capture Strategy

**Approach 1: Event Listeners (MVP)**

```javascript
// content-script.js
document.addEventListener('click', async (event) => {
  const element = event.target;
  const selector = await generateSelector(element);
  
  chrome.runtime.sendMessage({
    type: 'RECORD_EVENT',
    event: {
      action: 'click',
      selector: selector,
      timestamp: Date.now(),
      context: {
        url: window.location.href,
        text: element.textContent.trim().substring(0, 100),
        tag: element.tagName.toLowerCase(),
        attributes: extractAttributes(element)
      }
    }
  });
}, true);

document.addEventListener('input', async (event) => {
  const element = event.target;
  
  // Debounce to avoid too many events
  clearTimeout(window.inputDebounce);
  window.inputDebounce = setTimeout(async () => {
    const selector = await generateSelector(element);
    
    chrome.runtime.sendMessage({
      type: 'RECORD_EVENT',
      event: {
        action: 'input',
        selector: selector,
        value: element.value,
        timestamp: Date.now(),
        context: {
          url: window.location.href,
          field_type: element.type,
          placeholder: element.placeholder
        }
      }
    });
  }, 500);
}, true);
```

**Approach 2: Playwright Inspector Wrapper (Alternative)**

Instead of building Chrome extension from scratch, wrap Playwright's built-in recorder:

```bash
# Launch browser with Playwright Inspector
npx playwright open --save-storage=auth.json https://app.example.com
```

Pros:
- Already battle-tested
- Better selector generation
- Less development effort

Cons:
- Less control over UI/UX
- Harder to integrate with our backend

**Decision:** Start with Chrome Extension (Approach 1) for MVP, evaluate Playwright Inspector for Phase 2.

#### 4.1.3 Selector Generation

**Algorithm:**

```javascript
async function generateSelector(element) {
  // Priority 1: data-testid (most stable)
  if (element.dataset.testid) {
    return `[data-testid="${element.dataset.testid}"]`;
  }
  
  // Priority 2: aria-label (accessibility-first)
  if (element.getAttribute('aria-label')) {
    return `[aria-label="${element.getAttribute('aria-label')}"]`;
  }
  
  // Priority 3: role + text (semantic)
  if (element.getAttribute('role')) {
    const role = element.getAttribute('role');
    const text = element.textContent.trim().substring(0, 50);
    return `${role}[name="${text}"]`;
  }
  
  // Priority 4: ID (if stable)
  if (element.id && !isDynamicID(element.id)) {
    return `#${element.id}`;
  }
  
  // Priority 5: CSS selector (fallback)
  return generateCSSSelector(element);
}

function isDynamicID(id) {
  // Detect React/Vue generated IDs (e.g., "react-123", "vue-456")
  return /^(react|vue|angular|ember)-\d+/.test(id);
}

function generateCSSSelector(element) {
  const path = [];
  let current = element;
  
  while (current && current !== document.body) {
    let selector = current.tagName.toLowerCase();
    
    if (current.id) {
      selector += `#${current.id}`;
      path.unshift(selector);
      break; // ID is unique, stop here
    }
    
    if (current.className) {
      const classes = Array.from(current.classList)
        .filter(c => !isDynamicClass(c))
        .slice(0, 2); // Max 2 classes
      if (classes.length > 0) {
        selector += '.' + classes.join('.');
      }
    }
    
    // Add nth-child if needed
    const parent = current.parentElement;
    if (parent) {
      const siblings = Array.from(parent.children)
        .filter(c => c.tagName === current.tagName);
      if (siblings.length > 1) {
        const index = siblings.indexOf(current) + 1;
        selector += `:nth-child(${index})`;
      }
    }
    
    path.unshift(selector);
    current = current.parentElement;
  }
  
  return path.join(' > ');
}
```

### 4.2 Recorder Service (Backend)

#### 4.2.1 Responsibilities

- Receive events from Chrome extension via WebSocket
- Store raw events in database
- Manage recording sessions (start, stop, pause)
- Trigger AI processing pipeline

#### 4.2.2 API Endpoints

```go
// POST /api/v1/recordings/start
// Start a new recording session
{
  "name": "Login flow test",
  "description": "Verify admin can login",
  "tags": ["auth", "login"]
}

Response:
{
  "recording_id": "rec_abc123",
  "websocket_url": "wss://api.gotest.ai/recordings/rec_abc123/stream"
}
```

```go
// POST /api/v1/recordings/:id/stop
// Stop recording and trigger AI processing
{
  "recording_id": "rec_abc123"
}

Response:
{
  "status": "processing",
  "estimated_completion": "2026-07-30T10:35:00Z"
}
```

```go
// WebSocket /api/v1/recordings/:id/stream
// Real-time event streaming
Client → Server:
{
  "type": "event",
  "event": {
    "action": "click",
    "selector": "button#login",
    "timestamp": 1690716900000,
    "context": { ... }
  }
}

Server → Client:
{
  "type": "ack",
  "event_id": "evt_xyz789",
  "timestamp": 1690716900100
}
```

#### 4.2.3 Database Schema

```sql
-- Recording sessions
CREATE TABLE recording_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  tags TEXT[],
  
  -- Status
  status VARCHAR(50) NOT NULL DEFAULT 'recording',
  -- Values: recording, processing, completed, failed
  
  -- Ownership
  user_id UUID NOT NULL REFERENCES users(id),
  team_id UUID REFERENCES teams(id),
  
  -- Timestamps
  started_at TIMESTAMP NOT NULL DEFAULT NOW(),
  stopped_at TIMESTAMP,
  completed_at TIMESTAMP,
  
  -- Metadata
  duration_seconds INTEGER,
  event_count INTEGER DEFAULT 0,
  
  -- Soft delete
  deleted_at TIMESTAMP
);

-- Raw events (before AI processing)
CREATE TABLE recording_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES recording_sessions(id),
  
  -- Event data
  action VARCHAR(50) NOT NULL,
  selector TEXT,
  value TEXT,
  timestamp BIGINT NOT NULL,
  
  -- Context (JSONB for flexibility)
  context JSONB NOT NULL,
  
  -- Ordering
  sequence_number INTEGER NOT NULL,
  
  -- Timestamps
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_recording_events_session ON recording_events(session_id);
CREATE INDEX idx_recording_events_sequence ON recording_events(session_id, sequence_number);
CREATE INDEX idx_recording_sessions_user ON recording_sessions(user_id);
CREATE INDEX idx_recording_sessions_status ON recording_sessions(status);
```

### 4.3 AI Engine Service

#### 4.3.1 Responsibilities

- Process raw events into structured test steps
- Classify user intent (login, search, form submission, etc)
- Generate natural language descriptions
- Optimize selectors (stability scoring)
- Infer assertions (what should be verified)

#### 4.3.2 Processing Pipeline

```
Raw Events
    │
    ▼
┌──────────────────────────────────────┐
│  Step 1: Event Grouping              │
│  - Group related events              │
│  - Example: input + click = form     │
└──────────────────────────────────────┘
    │
    ▼
┌──────────────────────────────────────┐
│  Step 2: Intent Classification       │
│  - LLM: What is user trying to do?  │
│  - Output: intent, confidence, tags │
└──────────────────────────────────────┘
    │
    ▼
┌──────────────────────────────────────┐
│  Step 3: Selector Optimization       │
│  - Score selectors by stability      │
│  - Suggest alternatives              │
│  - Output: best selector + alts      │
└──────────────────────────────────────┘
    │
    ▼
┌──────────────────────────────────────┐
│  Step 4: Natural Language Gen        │
│  - LLM: Describe each action        │
│  - Output: human-readable text       │
└──────────────────────────────────────┘
    │
    ▼
┌──────────────────────────────────────┐
│  Step 5: Assertion Inference         │
│  - LLM: What should be verified?    │
│  - Output: assertions array          │
└──────────────────────────────────────┘
    │
    ▼
┌──────────────────────────────────────┐
│  Step 6: Test Assembly               │
│  - Combine all outputs               │
│  - Generate final test structure     │
│  - Output: JSON + Markdown           │
└──────────────────────────────────────┘
    │
    ▼
Structured Test (JSON + Markdown)
```

#### 4.3.3 AI Prompts

**Prompt 1: Intent Classification**

```python
INTENT_CLASSIFICATION_PROMPT = """
You are an AI assistant analyzing user interactions with a web application.

Given a sequence of user actions, classify the user's intent and suggest a test name.

User Actions:
{actions_json}

Context:
- Page URL: {url}
- Page Title: {title}

Analyze the actions and return JSON with:
- intent: Primary intent (authentication, search, form_submission, navigation, crud_operation, payment, other)
- confidence: Confidence score 0.0-1.0
- test_name: Clear, concise test name (e.g., "Admin login with valid credentials")
- tags: Array of relevant tags (e.g., ["auth", "login", "admin"])
- description: One-sentence description of what the test verifies

Example output:
{{
  "intent": "authentication",
  "confidence": 0.95,
  "test_name": "Admin login with valid credentials",
  "tags": ["auth", "login", "admin", "critical"],
  "description": "Verify that admin users can successfully login and access the dashboard"
}}
"""
```

**Prompt 2: Natural Language Generation**

```python
NLG_PROMPT = """
You are an AI assistant converting technical test actions into clear, human-readable descriptions.

Given a test action and its context, generate a concise description (1 sentence).

Action:
{action_json}

Context:
- Page URL: {url}
- Page Title: {title}
- Visible Elements: {visible_elements}

Generate a clear, concise description that a non-technical stakeholder would understand.

Examples:
- Action: click on button[type='submit'] → "Click the 'Submit' button"
- Action: type 'user@example.com' in input[name='email'] → "Enter 'user@example.com' in the email field"
- Action: navigate to /dashboard → "Navigate to the dashboard page"

Your description:
"""
```

**Prompt 3: Assertion Inference**

```python
ASSERTION_INFERENCE_PROMPT = """
You are an AI assistant inferring what should be verified after a sequence of user actions.

Given user actions and context, suggest 3-5 assertions that should be verified.

User Actions:
{actions_json}

Context:
- Page URL: {url}
- Page Title: {title}
- Expected Outcome: {expected_outcome}

Return JSON array of assertions:
[
  {{
    "type": "text_contains",
    "selector": ".success-message",
    "expected": "Operation completed successfully",
    "description": "Verify success message appears"
  }},
  {{
    "type": "url_contains",
    "expected": "/dashboard",
    "description": "Verify redirect to dashboard"
  }},
  {{
    "type": "visible",
    "selector": ".user-profile",
    "description": "Verify user profile is visible"
  }}
]

Supported assertion types:
- text_contains: Element contains specific text
- visible: Element is visible on page
- hidden: Element is hidden
- url_contains: URL contains specific string
- element_count: Specific number of elements exist
- attribute: Element has specific attribute value

Your assertions:
"""
```

**Prompt 4: Selector Optimization**

```python
SELECTOR_OPTIMIZATION_PROMPT = """
You are an AI assistant optimizing test selectors for stability and maintainability.

Given an element and multiple selector options, rank them by stability and suggest the best one.

Element Context:
{element_context}

Available Selectors:
{selectors_json}

Rank each selector with a stability score (0-100) and explanation:
- 100-90: Very stable (data-testid, aria-label, semantic selectors)
- 89-70: Stable (ID, role-based, text-based)
- 69-50: Moderate (specific classes, short CSS paths)
- 49-30: Unstable (generic classes, long CSS paths)
- 29-0: Very unstable (positional selectors, dynamic IDs)

Return JSON:
{{
  "rankings": [
    {{
      "selector": "[data-testid='login-button']",
      "score": 95,
      "reason": "Test-specific attribute, highly stable"
    }},
    {{
      "selector": "button[type='submit']",
      "score": 85,
      "reason": "Semantic selector based on button type"
    }},
    ...
  ],
  "best_selector": "[data-testid='login-button']",
  "recommendation": "Use data-testid for maximum stability"
}}
"""
```

#### 4.3.4 Selector Stability Scoring Algorithm

```python
def score_selector(selector: str, element_context: dict) -> int:
    """
    Score selector stability from 0-100.
    
    Scoring criteria:
    - data-testid: +100
    - aria-label: +90
    - role + text: +85
    - ID (non-dynamic): +80
    - type attribute: +75
    - name attribute: +70
    - placeholder: +65
    - text content: +60
    - specific class: +40
    - generic class: +20
    - XPath: +10
    - positional (nth-child): -30
    - long path (>3 levels): -20
    - dynamic ID pattern: -50
    """
    score = 0
    
    # High stability indicators
    if '[data-testid=' in selector:
        score = 100
    elif '[aria-label=' in selector:
        score = 90
    elif any(x in selector for x in ['role=', 'type=', 'name=']):
        score = 85
    elif selector.startswith('#') and not is_dynamic_id(selector[1:]):
        score = 80
    elif ':has-text(' in selector or ':text(' in selector:
        score = 75
    
    # Medium stability
    elif '.' in selector:
        # Count classes
        class_count = selector.count('.')
        if class_count == 1:
            score = 50
        elif class_count == 2:
            score = 40
        else:
            score = 30
    
    # Low stability
    elif '//' in selector:  # XPath
        score = 10
    
    # Penalties
    if 'nth-child' in selector or 'nth-of-type' in selector:
        score -= 30
    
    if selector.count('>') > 3:  # Long path
        score -= 20
    
    if is_dynamic_pattern(selector):
        score -= 50
    
    # Clamp to 0-100
    return max(0, min(100, score))

def is_dynamic_id(id_str: str) -> bool:
    """Detect dynamically generated IDs"""
    patterns = [
        r'^(react|vue|angular|ember)-\d+',  # Framework-generated
        r'^[a-f0-9]{8}-[a-f0-9]{4}-',       # UUID-like
        r'^\d+$',                             # Pure numbers
        r'^[a-z]+-\d+-\d+',                  # Pattern like "item-123-456"
    ]
    return any(re.match(p, id_str) for p in patterns)
```

### 4.4 Exporter Service

#### 4.4.1 Responsibilities

- Convert structured test (JSON) to executable code
- Support multiple frameworks (Playwright, Cypress, Selenium)
- Generate human-readable documentation (Markdown)

#### 4.4.2 Playwright Exporter

```go
type PlaywrightExporter struct{}

func (e *PlaywrightExporter) Export(test *TestRecording) (string, error) {
    var sb strings.Builder
    
    // Header
    sb.WriteString("import { test, expect } from '@playwright/test';\n\n")
    sb.WriteString(fmt.Sprintf("test('%s', async ({ page }) => {\n", test.Name))
    
    // Steps
    for _, step := range test.Steps {
        code := e.generateStepCode(step)
        sb.WriteString(fmt.Sprintf("  %s\n", code))
    }
    
    // Footer
    sb.WriteString("});\n")
    
    return sb.String(), nil
}

func (e *PlaywrightExporter) generateStepCode(step *TestStep) string {
    switch step.Action {
    case "navigate":
        return fmt.Sprintf("await page.goto('%s');", step.URL)
    
    case "click":
        return fmt.Sprintf("await page.click('%s');", step.Selector)
    
    case "type":
        return fmt.Sprintf("await page.fill('%s', '%s');", step.Selector, step.Value)
    
    case "select":
        return fmt.Sprintf("await page.selectOption('%s', '%s');", step.Selector, step.Value)
    
    case "wait":
        return fmt.Sprintf("await page.waitForTimeout(%d);", step.Timeout)
    
    case "assert":
        return e.generateAssertionCode(step.Assertion)
    
    default:
        return fmt.Sprintf("// Unsupported action: %s", step.Action)
    }
}

func (e *PlaywrightExporter) generateAssertionCode(assertion *Assertion) string {
    switch assertion.Type {
    case "text_contains":
        return fmt.Sprintf("await expect(page.locator('%s')).toContainText('%s');",
            assertion.Selector, assertion.Expected)
    
    case "visible":
        return fmt.Sprintf("await expect(page.locator('%s')).toBeVisible();",
            assertion.Selector)
    
    case "url_contains":
        return fmt.Sprintf("await expect(page).toHaveURL(/%s/);",
            assertion.Expected)
    
    default:
        return fmt.Sprintf("// Unsupported assertion: %s", assertion.Type)
    }
}
```

#### 4.4.3 Markdown Exporter

```go
type MarkdownExporter struct{}

func (e *MarkdownExporter) Export(test *TestRecording) (string, error) {
    var sb strings.Builder
    
    // Title
    sb.WriteString(fmt.Sprintf("# Test: %s\n\n", test.Name))
    
    // Description
    sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", test.Description))
    
    // Metadata
    sb.WriteString("## Metadata\n\n")
    sb.WriteString(fmt.Sprintf("- **Duration:** %d seconds\n", test.Duration))
    sb.WriteString(fmt.Sprintf("- **Tags:** %s\n", strings.Join(test.Tags, ", ")))
    sb.WriteString(fmt.Sprintf("- **Confidence:** %.0f%%\n\n", test.Confidence*100))
    
    // Steps
    sb.WriteString("## Steps\n\n")
    for i, step := range test.Steps {
        sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, step.Description))
        if step.Selector != "" {
            sb.WriteString(fmt.Sprintf("   - Selector: `%s`\n", step.Selector))
        }
        if step.Value != "" {
            sb.WriteString(fmt.Sprintf("   - Value: `%s`\n", step.Value))
        }
        sb.WriteString("\n")
    }
    
    // Assertions
    if len(test.Assertions) > 0 {
        sb.WriteString("## Assertions\n\n")
        for _, assertion := range test.Assertions {
            sb.WriteString(fmt.Sprintf("- ✅ %s\n", assertion.Description))
        }
    }
    
    return sb.String(), nil
}
```

### 4.5 Frontend (Next.js)

#### 4.5.1 Pages

**1. Recorder Page (`/recorder`)**

```tsx
// pages/recorder.tsx
export default function RecorderPage() {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingId, setRecordingId] = useState<string | null>(null);
  const [events, setEvents] = useState<Event[]>([]);
  
  const startRecording = async () => {
    const response = await fetch('/api/v1/recordings/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: 'New Recording',
        description: '',
        tags: []
      })
    });
    
    const { recording_id, websocket_url } = await response.json();
    setRecordingId(recording_id);
    
    // Connect to WebSocket
    const ws = new WebSocket(websocket_url);
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'event') {
        setEvents(prev => [...prev, data.event]);
      }
    };
    
    // Send message to Chrome extension
    chrome.runtime.sendMessage({
      type: 'START_RECORDING',
      recordingId: recording_id
    });
    
    setIsRecording(true);
  };
  
  const stopRecording = async () => {
    await fetch(`/api/v1/recordings/${recordingId}/stop`, {
      method: 'POST'
    });
    
    chrome.runtime.sendMessage({ type: 'STOP_RECORDING' });
    setIsRecording(false);
    
    // Redirect to test viewer
    router.push(`/tests/${recordingId}`);
  };
  
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Test Recorder</h1>
      
      {!isRecording ? (
        <Button onClick={startRecording} size="lg">
          <RecordIcon className="mr-2" />
          Start Recording
        </Button>
      ) : (
        <div>
          <div className="bg-red-100 border border-red-400 p-4 rounded mb-4">
            <p className="text-red-700 font-semibold">
              🔴 Recording in progress... ({events.length} events captured)
            </p>
          </div>
          
          <Button onClick={stopRecording} variant="destructive" size="lg">
            <StopIcon className="mr-2" />
            Stop Recording
          </Button>
          
          <div className="mt-6">
            <h2 className="text-xl font-semibold mb-3">Captured Events</h2>
            <EventList events={events} />
          </div>
        </div>
      )}
    </div>
  );
}
```

**2. Test Viewer Page (`/tests/[id]`)**

```tsx
// pages/tests/[id].tsx
export default function TestViewerPage() {
  const { id } = useRouter().query;
  const [test, setTest] = useState<TestRecording | null>(null);
  const [exportFormat, setExportFormat] = useState('playwright');
  
  useEffect(() => {
    fetch(`/api/v1/tests/${id}`)
      .then(res => res.json())
      .then(data => setTest(data));
  }, [id]);
  
  const handleExport = async () => {
    const response = await fetch(`/api/v1/tests/${id}/export`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ format: exportFormat })
    });
    
    const { code, filename } = await response.json();
    
    // Download file
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
  };
  
  if (!test) return <LoadingSpinner />;
  
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-2">{test.name}</h1>
      <p className="text-gray-600 mb-6">{test.description}</p>
      
      <div className="grid grid-cols-3 gap-6">
        {/* Left: Markdown documentation */}
        <div className="col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Test Documentation</CardTitle>
            </CardHeader>
            <CardContent>
              <MarkdownRenderer content={test.markdown} />
            </CardContent>
          </Card>
        </div>
        
        {/* Right: Export options */}
        <div className="col-span-1">
          <Card>
            <CardHeader>
              <CardTitle>Export Test</CardTitle>
            </CardHeader>
            <CardContent>
              <Select value={exportFormat} onChange={setExportFormat}>
                <SelectItem value="playwright">Playwright (JS/TS)</SelectItem>
                <SelectItem value="cypress">Cypress (JS)</SelectItem>
                <SelectItem value="selenium">Selenium (Python)</SelectItem>
                <SelectItem value="markdown">Markdown</SelectItem>
                <SelectItem value="json">JSON</SelectItem>
              </Select>
              
              <Button onClick={handleExport} className="w-full mt-4">
                <DownloadIcon className="mr-2" />
                Export
              </Button>
            </CardContent>
          </Card>
          
          <Card className="mt-4">
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full mb-2">
                <PlayIcon className="mr-2" />
                Run Test
              </Button>
              <Button variant="outline" className="w-full mb-2">
                <EditIcon className="mr-2" />
                Edit Test
              </Button>
              <Button variant="outline" className="w-full">
                <ShareIcon className="mr-2" />
                Share
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
```

**3. Test Library Page (`/library`)**

```tsx
// pages/library.tsx
export default function TestLibraryPage() {
  const [tests, setTests] = useState<TestRecording[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    if (selectedTags.length > 0) params.set('tags', selectedTags.join(','));
    
    fetch(`/api/v1/tests?${params.toString()}`)
      .then(res => res.json())
      .then(data => setTests(data.tests));
  }, [searchQuery, selectedTags]);
  
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Test Library</h1>
      
      {/* Search & Filter */}
      <div className="flex gap-4 mb-6">
        <Input
          placeholder="Search tests..."
          value={searchQuery}
          onChange={setSearchQuery}
          className="flex-1"
        />
        <TagFilter
          selectedTags={selectedTags}
          onChange={setSelectedTags}
        />
      </div>
      
      {/* Test Grid */}
      <div className="grid grid-cols-3 gap-4">
        {tests.map(test => (
          <TestCard key={test.id} test={test} />
        ))}
      </div>
    </div>
  );
}

function TestCard({ test }: { test: TestRecording }) {
  return (
    <Link href={`/tests/${test.id}`}>
      <Card className="hover:shadow-lg transition-shadow cursor-pointer">
        <CardHeader>
          <CardTitle className="text-lg">{test.name}</CardTitle>
          <CardDescription>{test.description}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 mb-3">
            {test.tags.map(tag => (
              <Badge key={tag} variant="secondary">{tag}</Badge>
            ))}
          </div>
          <div className="flex justify-between text-sm text-gray-500">
            <span>{test.steps.length} steps</span>
            <span>{test.duration}s</span>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
```

---

## 5. Data Flow

### 5.1 Recording Flow

```
User clicks "Start Recording"
    │
    ▼
Frontend: POST /api/v1/recordings/start
    │
    ▼
Backend: Create recording_sessions record
    │
    ▼
Backend: Return recording_id + WebSocket URL
    │
    ▼
Frontend: Connect to WebSocket
    │
    ▼
Frontend: Send START_RECORDING to Chrome extension
    │
    ▼
Chrome Extension: Begin capturing DOM events
    │
    ▼
User interacts with web app (click, type, navigate)
    │
    ▼
Chrome Extension: Capture event + generate selector
    │
    ▼
Chrome Extension: Send event via WebSocket
    │
    ▼
Backend: Receive event
    │
    ▼
Backend: Store in recording_events table
    │
    ▼
Backend: Send ACK to Chrome extension
    │
    ▼
Frontend: Display event in real-time UI
    │
    ▼
[Repeat for each user action]
    │
    ▼
User clicks "Stop Recording"
    │
    ▼
Frontend: POST /api/v1/recordings/:id/stop
    │
    ▼
Backend: Update status to "processing"
    │
    ▼
Backend: Trigger AI processing pipeline (async)
    │
    ▼
Frontend: Redirect to /tests/:id
```

### 5.2 AI Processing Flow

```
Backend: recording_sessions.status = "processing"
    │
    ▼
AI Service: Fetch all recording_events for session
    │
    ▼
AI Service: Event grouping (group related events)
    │
    ▼
AI Service: Intent classification (LLM call)
    │
    ▼
AI Service: Selector optimization (scoring algorithm)
    │
    ▼
AI Service: Natural language generation (LLM call)
    │
    ▼
AI Service: Assertion inference (LLM call)
    │
    ▼
AI Service: Assemble structured test (JSON)
    │
    ▼
AI Service: Generate Markdown documentation
    │
    ▼
Backend: Store in test_recordings table
    │
    ▼
Backend: Update status to "completed"
    │
    ▼
Frontend: Poll status, display test when ready
```

### 5.3 Export Flow

```
User clicks "Export" button
    │
    ▼
Frontend: POST /api/v1/tests/:id/export
    │
    ▼
Backend: Fetch test_recordings record
    │
    ▼
Backend: Select exporter (Playwright/Cypress/Selenium/Markdown)
    │
    ▼
Backend: Generate code/documentation
    │
    ▼
Backend: Return code + filename
    │
    ▼
Frontend: Trigger file download
```

---

## 6. API Design

### 6.1 Recording APIs

```yaml
# Start recording
POST /api/v1/recordings/start
Request:
  {
    "name": "string",
    "description": "string",
    "tags": ["string"]
  }
Response:
  {
    "recording_id": "uuid",
    "websocket_url": "wss://..."
  }

# Stop recording
POST /api/v1/recordings/:id/stop
Response:
  {
    "status": "processing",
    "estimated_completion": "timestamp"
  }

# Get recording status
GET /api/v1/recordings/:id
Response:
  {
    "id": "uuid",
    "status": "recording|processing|completed|failed",
    "event_count": 10,
    "duration_seconds": 30,
    "test_id": "uuid"  # Only when completed
  }

# WebSocket stream
WS /api/v1/recordings/:id/stream
Client → Server:
  {
    "type": "event",
    "event": {
      "action": "click|input|navigate|...",
      "selector": "string",
      "value": "string",
      "timestamp": 1234567890,
      "context": { ... }
    }
  }
Server → Client:
  {
    "type": "ack",
    "event_id": "uuid",
    "timestamp": 1234567890
  }
```

### 6.2 Test APIs

```yaml
# Get test
GET /api/v1/tests/:id
Response:
  {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "tags": ["string"],
    "steps": [
      {
        "id": 1,
        "action": "click",
        "selector": "button#login",
        "selector_alternatives": ["[data-testid='login']"],
        "value": null,
        "description": "Click the login button",
        "context": { ... }
      }
    ],
    "assertions": [
      {
        "type": "text_contains",
        "selector": ".success-message",
        "expected": "Welcome!",
        "description": "Verify success message appears"
      }
    ],
    "metadata": {
      "duration_seconds": 30,
      "confidence_score": 0.95,
      "recorded_at": "timestamp"
    },
    "markdown": "# Test: Login flow\n\n..."
  }

# List tests
GET /api/v1/tests?q=search&tags=auth,login&page=1&limit=20
Response:
  {
    "tests": [ ... ],
    "total": 100,
    "page": 1,
    "limit": 20
  }

# Export test
POST /api/v1/tests/:id/export
Request:
  {
    "format": "playwright|cypress|selenium|markdown|json"
  }
Response:
  {
    "code": "import { test } from '@playwright/test';\n...",
    "filename": "login-flow.spec.ts"
  }

# Delete test
DELETE /api/v1/tests/:id
Response: 204 No Content

# Update test
PATCH /api/v1/tests/:id
Request:
  {
    "name": "Updated name",
    "description": "Updated description",
    "tags": ["auth", "login", "critical"]
  }
Response:
  {
    "id": "uuid",
    "name": "Updated name",
    ...
  }
```

### 6.3 Test Library APIs

```yaml
# Search tests
GET /api/v1/tests/search?q=login&tags=auth&intent=authentication
Response:
  {
    "results": [ ... ],
    "total": 50,
    "facets": {
      "tags": { "auth": 30, "login": 25, ... },
      "intents": { "authentication": 20, ... }
    }
  }

# Get popular tests
GET /api/v1/tests/popular?limit=10
Response:
  {
    "tests": [ ... ]
  }

# Get recent tests
GET /api/v1/tests/recent?limit=10
Response:
  {
    "tests": [ ... ]
  }
```

---

## 7. Database Schema

### 7.1 Core Tables

```sql
-- Users (existing)
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  created_at TIMESTAMP DEFAULT NOW()
);

-- Teams (existing)
CREATE TABLE teams (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Recording sessions
CREATE TABLE recording_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  tags TEXT[],
  
  status VARCHAR(50) NOT NULL DEFAULT 'recording',
  -- Values: recording, processing, completed, failed
  
  user_id UUID NOT NULL REFERENCES users(id),
  team_id UUID REFERENCES teams(id),
  
  started_at TIMESTAMP NOT NULL DEFAULT NOW(),
  stopped_at TIMESTAMP,
  completed_at TIMESTAMP,
  
  duration_seconds INTEGER,
  event_count INTEGER DEFAULT 0,
  
  deleted_at TIMESTAMP
);

-- Raw events
CREATE TABLE recording_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES recording_sessions(id) ON DELETE CASCADE,
  
  action VARCHAR(50) NOT NULL,
  selector TEXT,
  value TEXT,
  timestamp BIGINT NOT NULL,
  
  context JSONB NOT NULL,
  
  sequence_number INTEGER NOT NULL,
  
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Processed tests
CREATE TABLE test_recordings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recording_session_id UUID REFERENCES recording_sessions(id),
  
  name VARCHAR(255) NOT NULL,
  description TEXT,
  tags TEXT[],
  
  intent VARCHAR(100),
  confidence_score FLOAT,
  
  steps JSONB NOT NULL,
  assertions JSONB NOT NULL DEFAULT '[]',
  
  markdown TEXT NOT NULL,
  
  duration_seconds INTEGER,
  step_count INTEGER,
  
  user_id UUID NOT NULL REFERENCES users(id),
  team_id UUID REFERENCES teams(id),
  
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP
);

-- Test parameters (for data-driven tests)
CREATE TABLE test_parameters (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  test_id UUID NOT NULL REFERENCES test_recordings(id) ON DELETE CASCADE,
  
  name VARCHAR(255) NOT NULL,
  type VARCHAR(50) NOT NULL,
  default_value TEXT,
  description TEXT,
  
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Test execution history
CREATE TABLE test_executions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  test_id UUID NOT NULL REFERENCES test_recordings(id),
  
  status VARCHAR(50) NOT NULL,
  -- Values: passed, failed, error, skipped
  
  duration_ms INTEGER,
  error_message TEXT,
  
  executed_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 7.2 Indexes

```sql
-- Recording sessions
CREATE INDEX idx_recording_sessions_user ON recording_sessions(user_id);
CREATE INDEX idx_recording_sessions_team ON recording_sessions(team_id);
CREATE INDEX idx_recording_sessions_status ON recording_sessions(status);
CREATE INDEX idx_recording_sessions_started ON recording_sessions(started_at DESC);

-- Recording events
CREATE INDEX idx_recording_events_session ON recording_events(session_id);
CREATE INDEX idx_recording_events_sequence ON recording_events(session_id, sequence_number);

-- Test recordings
CREATE INDEX idx_test_recordings_user ON test_recordings(user_id);
CREATE INDEX idx_test_recordings_team ON test_recordings(team_id);
CREATE INDEX idx_test_recordings_tags ON test_recordings USING GIN(tags);
CREATE INDEX idx_test_recordings_intent ON test_recordings(intent);
CREATE INDEX idx_test_recordings_created ON test_recordings(created_at DESC);

-- Full-text search
CREATE INDEX idx_test_recordings_search ON test_recordings USING GIN(
  to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
);

-- Test executions
CREATE INDEX idx_test_executions_test ON test_executions(test_id);
CREATE INDEX idx_test_executions_status ON test_executions(status);
CREATE INDEX idx_test_executions_executed ON test_executions(executed_at DESC);
```

---

## 8. Security Considerations

### 8.1 Data Privacy

**Sensitive Data in Recordings:**
- Passwords
- Credit card numbers
- Personal information (PII)
- API keys
- Session tokens

**Mitigation Strategies:**

1. **Automatic Sensitive Data Detection**

```python
SENSITIVE_PATTERNS = [
    # Passwords
    r'password',
    r'passwd',
    r'pwd',
    
    # Credit cards
    r'\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b',
    
    # Email
    r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b',
    
    # Phone numbers
    r'\b\d{3}[-.]?\d{3}[-.]?\d{4}\b',
    
    # SSN
    r'\b\d{3}-\d{2}-\d{4}\b',
    
    # API keys (generic pattern)
    r'\b[a-zA-Z0-9]{32,}\b',
]

def mask_sensitive_data(value: str, field_name: str) -> str:
    """Mask sensitive data in recorded values"""
    
    # Check field name
    if any(pattern in field_name.lower() for pattern in ['password', 'secret', 'key', 'token']):
        return '***MASKED***'
    
    # Check value patterns
    for pattern in SENSITIVE_PATTERNS:
        if re.search(pattern, value):
            return '***MASKED***'
    
    return value
```

2. **User Consent & Warnings**

```tsx
// Before recording starts
<Dialog open={showConsentDialog}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>⚠️ Sensitive Data Warning</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      This recorder will capture all user interactions, including:
      <ul className="list-disc ml-6 mt-2">
        <li>Text entered in input fields</li>
        <li>URLs visited</li>
        <li>Button clicks and form submissions</li>
      </ul>
      
      <div className="bg-yellow-100 border border-yellow-400 p-4 mt-4 rounded">
        <p className="font-semibold text-yellow-900">
          ⚠️ Do NOT record sessions containing:
        </p>
        <ul className="list-disc ml-6 mt-2 text-yellow-800">
          <li>Real passwords or credentials</li>
          <li>Credit card numbers</li>
          <li>Personal information (SSN, phone, email)</li>
          <li>API keys or secrets</li>
        </ul>
      </div>
      
      <p className="mt-4">
        Use test accounts and dummy data only.
      </p>
    </DialogDescription>
    <DialogFooter>
      <Button variant="outline" onClick={cancelRecording}>
        Cancel
      </Button>
      <Button onClick={confirmRecording}>
        I Understand, Start Recording
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

3. **Data Retention Policy**

```sql
-- Auto-delete old recordings
CREATE POLICY auto_delete_old_recordings ON recording_sessions
  FOR DELETE USING (
    created_at < NOW() - INTERVAL '90 days'
    AND deleted_at IS NULL
  );

-- Soft delete instead of hard delete
UPDATE recording_sessions
SET deleted_at = NOW()
WHERE id = :recording_id;
```

### 8.2 Authentication & Authorization

```go
// Middleware: Require authentication
func requireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        user, err := validateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Middleware: Require team membership
func requireTeamAccess(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := r.Context().Value("user").(*User)
        teamID := r.URL.Query().Get("team_id")
        
        if !user.HasTeamAccess(teamID) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 8.3 Rate Limiting

```go
// Rate limit recording API
func rateLimitRecording(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 10) // 10 recordings per minute
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## 9. Performance Requirements

### 9.1 Latency Targets

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Event capture (Chrome extension) | < 50ms | Time from user action to event sent |
| Event storage (backend) | < 100ms | Time from event received to DB commit |
| AI processing (per recording) | < 30s | Time from recording stop to test ready |
| Test export | < 2s | Time from export request to code generated |
| Test search | < 500ms | Time from search query to results displayed |

### 9.2 Scalability Targets

| Metric | MVP (Month 3) | Post-Launch (Month 6) |
|--------|---------------|----------------------|
| Concurrent recordings | 100 | 1,000 |
| Tests recorded per day | 1,000 | 10,000 |
| Total tests in database | 50,000 | 500,000 |
| API requests per second | 100 | 1,000 |
| WebSocket connections | 500 | 5,000 |

### 9.3 Optimization Strategies

**1. Event Batching**

```javascript
// Chrome extension: Batch events to reduce WebSocket messages
let eventBuffer = [];
const BATCH_SIZE = 10;
const BATCH_INTERVAL = 1000; // 1 second

setInterval(() => {
  if (eventBuffer.length > 0) {
    websocket.send(JSON.stringify({
      type: 'event_batch',
      events: eventBuffer
    }));
    eventBuffer = [];
  }
}, BATCH_INTERVAL);

function recordEvent(event) {
  eventBuffer.push(event);
  
  if (eventBuffer.length >= BATCH_SIZE) {
    websocket.send(JSON.stringify({
      type: 'event_batch',
      events: eventBuffer
    }));
    eventBuffer = [];
  }
}
```

**2. AI Processing Queue**

```go
// Use job queue for AI processing (avoid blocking)
type AIProcessingJob struct {
    RecordingID uuid.UUID
    CreatedAt   time.Time
}

func enqueueAIProcessing(recordingID uuid.UUID) error {
    job := AIProcessingJob{
        RecordingID: recordingID,
        CreatedAt:   time.Now(),
    }
    
    return redisQueue.Enqueue("ai_processing", job)
}

// Worker processes jobs asynchronously
func aiProcessingWorker(ctx context.Context, job AIProcessingJob) error {
    // Fetch events
    events, err := getRecordingEvents(job.RecordingID)
    if err != nil {
        return err
    }
    
    // Process with AI
    test, err := processWithAI(events)
    if err != nil {
        return err
    }
    
    // Store result
    err = storeTestRecording(test)
    if err != nil {
        return err
    }
    
    // Update status
    err = updateRecordingStatus(job.RecordingID, "completed")
    return err
}
```

**3. Database Optimization**

```sql
-- Partition large tables by date
CREATE TABLE recording_events (
  id UUID,
  session_id UUID,
  ...
  created_at TIMESTAMP
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE recording_events_2026_07 PARTITION OF recording_events
  FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE recording_events_2026_08 PARTITION OF recording_events
  FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

-- Archive old data
CREATE TABLE recording_events_archive (LIKE recording_events);

-- Move old data to archive monthly
INSERT INTO recording_events_archive
SELECT * FROM recording_events
WHERE created_at < NOW() - INTERVAL '6 months';

DELETE FROM recording_events
WHERE created_at < NOW() - INTERVAL '6 months';
```

---

## 10. Implementation Phases

### Phase 1: Basic Recording (Weeks 1-4)

**Week 1-2: Chrome Extension**
- [ ] Setup Chrome extension project (Manifest V3)
- [ ] Implement event listeners (click, input, navigate)
- [ ] Selector generation algorithm
- [ ] WebSocket communication with backend
- [ ] Basic UI (start/stop recording button)

**Week 3-4: Backend**
- [ ] Database schema (recording_sessions, recording_events)
- [ ] REST API endpoints (start, stop, get status)
- [ ] WebSocket server (event streaming)
- [ ] Event storage logic
- [ ] Basic frontend (recorder page, test list)

**Deliverables:**
- ✅ Chrome extension that records user actions
- ✅ Backend that stores raw events
- ✅ Frontend to start/stop recordings
- ✅ Real-time event display

### Phase 2: AI Processing (Weeks 5-8)

**Week 5-6: AI Engine**
- [ ] Intent classification prompt + LLM integration
- [ ] Natural language generation prompt
- [ ] Selector optimization algorithm (scoring)
- [ ] Assertion inference prompt
- [ ] AI processing pipeline (async job queue)

**Week 7-8: Structured Test Generation**
- [ ] JSON test structure assembly
- [ ] Markdown documentation generation
- [ ] Test storage (test_recordings table)
- [ ] Test viewer UI (display steps, assertions)

**Deliverables:**
- ✅ AI processes raw events into structured tests
- ✅ Human-readable Markdown documentation
- ✅ Optimized selectors with stability scores
- ✅ Inferred assertions

### Phase 3: Export & Library (Weeks 9-12)

**Week 9-10: Export Service**
- [ ] Playwright exporter (JavaScript/TypeScript)
- [ ] Markdown exporter
- [ ] JSON exporter
- [ ] Export API endpoint
- [ ] Download functionality in UI

**Week 11-12: Test Library**
- [ ] Test list page (search, filter, paginate)
- [ ] Test detail page (view, edit, export)
- [ ] Tag management
- [ ] Basic team sharing (view-only)

**Deliverables:**
- ✅ Export to Playwright code
- ✅ Export to Markdown
- ✅ Test library with search & filter
- ✅ Team sharing (view-only)

### Phase 4: Polish & Beta (Weeks 13-16)

**Week 13-14: Polish**
- [ ] Error handling & edge cases
- [ ] Loading states & error messages
- [ ] Performance optimization
- [ ] Security review
- [ ] Documentation (user guide, API docs)

**Week 15-16: Beta Launch**
- [ ] Deploy to production
- [ ] Onboard 50 beta users
- [ ] Collect feedback
- [ ] Bug fixes & improvements
- [ ] Prepare for public launch

**Deliverables:**
- ✅ Production-ready MVP
- ✅ 50+ beta users
- ✅ User documentation
- ✅ Feedback collection system

---

## 11. Testing Strategy

### 11.1 Unit Tests

**Chrome Extension:**
- [ ] Selector generation (10+ test cases)
- [ ] Event capture (click, input, navigate)
- [ ] Sensitive data masking
- [ ] WebSocket message formatting

**Backend:**
- [ ] Event storage (CRUD operations)
- [ ] Selector scoring algorithm
- [ ] AI prompt formatting
- [ ] Export code generation (Playwright, Markdown)

**Frontend:**
- [ ] Recorder UI (start/stop, event display)
- [ ] Test viewer (step rendering, export dialog)
- [ ] Test library (search, filter, pagination)

### 11.2 Integration Tests

**End-to-End Recording Flow:**
1. Start recording via API
2. Simulate user actions (Chrome extension mock)
3. Send events via WebSocket
4. Stop recording
5. Wait for AI processing
6. Verify structured test generated
7. Export to Playwright
8. Verify exported code is valid

**Test Coverage Target:** 80%

### 11.3 AI Quality Tests

**Intent Classification Accuracy:**
- [ ] Prepare 100 labeled recordings (ground truth)
- [ ] Run AI classification
- [ ] Measure accuracy (target: >90%)
- [ ] Analyze misclassifications
- [ ] Iterate on prompts

**Selector Optimization Quality:**
- [ ] Prepare 50 elements with known stable selectors
- [ ] Run selector scoring algorithm
- [ ] Verify top-ranked selector matches ground truth
- [ ] Target: >85% accuracy

**Assertion Inference Relevance:**
- [ ] Prepare 50 recordings with expected assertions
- [ ] Run AI assertion inference
- [ ] Manual review: Are assertions relevant?
- [ ] Target: >80% relevance

### 11.4 Performance Tests

**Load Testing (k6):**
```javascript
// k6 load test script
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 50 },  // Ramp up to 50 users
    { duration: '5m', target: 50 },  // Stay at 50 users
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
};

export default function () {
  // Simulate recording start
  const startRes = http.post('http://localhost:8080/api/v1/recordings/start', JSON.stringify({
    name: 'Load test recording',
    description: 'Automated load test',
    tags: ['load-test']
  }));
  
  check(startRes, {
    'start recording status is 200': (r) => r.status === 200,
  });
  
  const { recording_id } = JSON.parse(startRes.body);
  
  // Simulate event streaming (10 events)
  for (let i = 0; i < 10; i++) {
    // In real test, would use WebSocket
    sleep(0.5);
  }
  
  // Stop recording
  const stopRes = http.post(`http://localhost:8080/api/v1/recordings/${recording_id}/stop`);
  
  check(stopRes, {
    'stop recording status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}
```

**Metrics to Monitor:**
- Response time (p50, p95, p99)
- Throughput (requests/second)
- Error rate (<1%)
- WebSocket connection stability
- AI processing queue depth

---

## 12. Risks & Mitigations

### 12.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Chrome extension rejected by Chrome Web Store | Medium | High | Distribute via direct download initially; apply for store later |
| LLM costs too high | Medium | High | Batch AI calls; cache similar requests; use smaller models for simple tasks |
| Selector generation inaccurate | High | Medium | Iterative improvement; user feedback loop; manual override option |
| AI hallucinations (wrong intent/assertions) | High | Medium | Confidence scoring; user review step; feedback collection |
| WebSocket scalability issues | Medium | Medium | Use Redis Pub/Sub; horizontal scaling; connection pooling |
| Database performance degradation | Medium | High | Partitioning; archiving; read replicas; query optimization |

### 12.2 Product Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Users don't understand value proposition | Medium | High | Clear onboarding; video tutorials; case studies |
| Competitors copy feature | High | Medium | Move fast; build network effects; patent key algorithms |
| Low adoption (nobody uses it) | Medium | High | Beta program; user interviews; iterate based on feedback |
| Users record sensitive data accidentally | High | High | Clear warnings; auto-detection; consent dialog |
| Test maintenance still too high | Medium | High | Self-healing (Phase 2); better selectors; AI suggestions |

### 12.3 Business Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Pricing too high/low | Medium | Medium | A/B testing; competitor analysis; customer interviews |
| CAC (Customer Acquisition Cost) too high | Medium | High | Content marketing; open-source strategy; referral program |
| Churn rate too high | Medium | High | Improve product; customer success; long-term contracts |
| Funding runs out before product-market fit | Medium | High | Lean development; focus on revenue; extend runway |

---

## 13. Success Metrics & KPIs

### 13.1 Product Metrics

**Acquisition:**
- Website visitors (target: 10k/month by Month 6)
- Signups (target: 500/month by Month 6)
- Activation rate (target: 30% of signups record first test)

**Engagement:**
- Tests recorded per user per week (target: 5+)
- Tests exported per user per week (target: 3+)
- Test library usage (target: 50% of users browse library)

**Retention:**
- Week 1 retention (target: 60%)
- Month 1 retention (target: 40%)
- Month 3 retention (target: 25%)

**Quality:**
- Test pass rate (target: 80%+)
- Self-healing success rate (target: 70%+ by Phase 2)
- AI accuracy (intent classification: 90%+, selector optimization: 85%+)

### 13.2 Business Metrics

**Revenue:**
- MRR (Monthly Recurring Revenue) (target: $50k by Month 12)
- ARPU (Average Revenue Per User) (target: $50/month)
- LTV (Lifetime Value) (target: $1,200)

**Costs:**
- CAC (Customer Acquisition Cost) (target: <$300)
- LLM API costs per user (target: <$5/month)
- Infrastructure costs per user (target: <$2/month)

**Health:**
- LTV/CAC ratio (target: >3x)
- Churn rate (target: <5%/month)
- NPS (Net Promoter Score) (target: >40)

### 13.3 Technical Metrics

**Performance:**
- API latency (p95 < 500ms)
- AI processing time (<30s per recording)
- WebSocket connection uptime (>99.9%)

**Reliability:**
- Uptime (>99.9%)
- Error rate (<1%)
- Data durability (99.999%)

**Scalability:**
- Concurrent recordings (target: 1,000 by Month 6)
- Tests per day (target: 10,000 by Month 6)
- Database size (monitor growth, plan for sharding)

---

## 14. Open Questions & Decisions Needed

### 14.1 Product Decisions

**Q1: Should we build Chrome extension or wrap Playwright Inspector?**
- **Option A:** Build Chrome extension (more control, better UX)
- **Option B:** Wrap Playwright Inspector (faster to ship, battle-tested)
- **Recommendation:** Option A for MVP, evaluate Option B for Phase 2
- **Decision needed by:** Week 1

**Q2: Should we support multiple frameworks from day 1?**
- **Option A:** Playwright only (faster to ship, focus)
- **Option B:** Playwright + Cypress + Selenium (more users, more work)
- **Recommendation:** Option A for MVP, add others in Phase 2
- **Decision needed by:** Week 1

**Q3: How to handle sensitive data in recordings?**
- **Option A:** Auto-detect and mask (complex, might miss some)
- **Option B:** User responsibility + warnings (simple, legal risk)
- **Option C:** Hybrid (auto-detect + warnings + user review)
- **Recommendation:** Option C
- **Decision needed by:** Week 2

**Q4: Pricing model?**
- **Option A:** Freemium (free tier + paid)
- **Option B:** Free trial (14 days) + paid only
- **Option C:** Fully open-source + paid cloud hosting
- **Recommendation:** Option A (freemium)
- **Decision needed by:** Week 8 (before beta launch)

### 14.2 Technical Decisions

**Q5: Which LLM to use for AI processing?**
- **Option A:** OpenAI GPT-4 (best quality, expensive)
- **Option B:** Anthropic Claude (good quality, competitive pricing)
- **Option C:** Open-source (Llama, Mistral) (cheap, lower quality)
- **Option D:** Hybrid (GPT-4 for complex tasks, smaller models for simple)
- **Recommendation:** Option D (hybrid approach)
- **Decision needed by:** Week 5

**Q6: How to store test steps?**
- **Option A:** JSONB in PostgreSQL (flexible, queryable)
- **Option B:** Separate table per step type (normalized, complex)
- **Option C:** Document database (MongoDB) (flexible, less SQL)
- **Recommendation:** Option A (JSONB in PostgreSQL)
- **Decision needed by:** Week 3

**Q7: Real-time event streaming approach?**
- **Option A:** WebSocket (bidirectional, low latency)
- **Option B:** Server-Sent Events (SSE, simpler, unidirectional)
- **Option C:** HTTP polling (simplest, high latency)
- **Recommendation:** Option A (WebSocket)
- **Decision needed by:** Week 3

**Q8: AI processing: synchronous or asynchronous?**
- **Option A:** Synchronous (user waits, simpler)
- **Option B:** Asynchronous (background job, better UX)
- **Recommendation:** Option B (async with job queue)
- **Decision needed by:** Week 5

### 14.3 Business Decisions

**Q9: Target market segment?**
- **Option A:** Startups (10-100 engineers)
- **Option B:** SMEs (100-1000 employees)
- **Option C:** Enterprise (1000+ employees)
- **Recommendation:** Option A (startups) for MVP, expand later
- **Decision needed by:** Week 1

**Q10: Go-to-market strategy?**
- **Option A:** Product Hunt launch (fast, broad reach)
- **Option B:** Content marketing (slow, high-quality leads)
- **Option C:** Direct sales (targeted, high-touch)
- **Option D:** Hybrid (Product Hunt + content + partnerships)
- **Recommendation:** Option D (hybrid)
- **Decision needed by:** Week 12

**Q11: Open-source strategy?**
- **Option A:** Fully open-source (build community, monetize cloud)
- **Option B:** Open-core (free basic, paid advanced features)
- **Option C:** Closed-source (proprietary, paid only)
- **Recommendation:** Option B (open-core)
- **Decision needed by:** Week 12

**Q12: Funding strategy?**
- **Option A:** Bootstrap (self-funded, slower growth)
- **Option B:** Seed round ($1-2M, faster growth)
- **Option C:** Series A ($5-10M, aggressive growth)
- **Recommendation:** Option A for MVP, evaluate B/C based on traction
- **Decision needed by:** Month 6

---

## 15. Appendix

### 15.1 Glossary

- **Recording Session:** A single recording instance from start to stop
- **Raw Event:** Unprocessed user action (click, input, navigate)
- **Structured Test:** AI-processed test with steps, descriptions, assertions
- **Selector:** CSS/XPath selector to locate element on page
- **Stability Score:** 0-100 score indicating selector reliability
- **Intent Classification:** AI-determined purpose of user actions (login, search, etc)
- **Assertion Inference:** AI-suggested verifications (what should be true after actions)
- **Self-Healing:** Automatic selector update when UI changes (Phase 2)
- **Test Composition:** Reusing recorded tests as building blocks (Phase 3)

### 15.2 References

- [Playwright Documentation](https://playwright.dev/docs/)
- [Chrome Extension Manifest V3](https://developer.chrome.com/docs/extensions/mv3/)
- [OpenAI API Documentation](https://platform.openai.com/docs/)
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
- [WebSocket Protocol](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

### 15.3 Competitive Analysis

**Katalon Studio:**
- Strengths: Mature, enterprise-ready, comprehensive
- Weaknesses: Expensive, vendor lock-in, no AI understanding
- Our advantage: AI-powered, human-readable, multi-framework export

**Selenium IDE:**
- Strengths: Free, open-source, widely used
- Weaknesses: Limited features, no AI, brittle selectors
- Our advantage: AI optimization, self-healing, modern UX

**Playwright Codegen:**
- Strengths: Good selector generation, modern
- Weaknesses: No AI understanding, single framework
- Our advantage: Intent classification, assertion inference, test library

**Testim:**
- Strengths: AI-powered, self-healing
- Weaknesses: Expensive, proprietary format
- Our advantage: Open-source, multi-framework, human-readable

### 15.4 Wireframes

**Recorder UI:**

```
┌─────────────────────────────────────────────────────────────┐
│  Test Recorder                                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  🔴 Recording in progress... (15 events captured)           │
│                                                              │
│  [Stop Recording]                                            │
│                                                              │
│  ────────────────────────────────────────────────────────   │
│                                                              │
│  Captured Events:                                            │
│                                                              │
│  1. Navigate to https://app.example.com/login               │
│  2. Type "admin@example.com" in email field                 │
│  3. Type "password123" in password field                    │
│  4. Click "Sign In" button                                  │
│  5. Wait for dashboard page                                 │
│  ...                                                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Test Viewer UI:**

```
┌─────────────────────────────────────────────────────────────┐
│  Test: Admin login with valid credentials                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [Export ▼]  [Run Test]  [Edit]  [Share]                   │
│                                                              │
│  ────────────────────────────────────────────────────────   │
│                                                              │
│  ## Description                                              │
│  Verify that admin users can successfully login and         │
│  access the dashboard                                       │
│                                                              │
│  ## Steps                                                    │
│  1. **Open login page**                                     │
│     - Navigate to `https://app.example.com/login`           │
│                                                              │
│  2. **Enter admin email**                                   │
│     - Type `admin@example.com` in email field               │
│     - Selector: `input[name='email']`                       │
│                                                              │
│  3. **Enter password**                                      │
│     - Type `password123` in password field                  │
│     - Selector: `input[name='password']`                    │
│                                                              │
│  4. **Click login button**                                  │
│     - Click "Sign In" button                                │
│     - Selector: `button[type='submit']`                     │
│                                                              │
│  ## Assertions                                               │
│  ✅ Verify URL contains `/dashboard`                        │
│  ✅ Verify welcome message appears                          │
│  ✅ Verify user profile is visible                          │
│                                                              │
│  ────────────────────────────────────────────────────────   │
│                                                              │
│  Export Format: [Playwright ▼]                              │
│  [Download Code]                                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 16. Sign-Off

**Technical Lead:** _________________________ Date: _________

**Product Manager:** _________________________ Date: _________

**Engineering Manager:** _________________________ Date: _________

**CTO:** _________________________ Date: _________

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-07-30 | Engineering Team | Initial draft |

---

**End of Technical Design Document**
