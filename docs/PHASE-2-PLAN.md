# Phase 2: Record & Playback

**Timeline:** 12 minggu (6 sprint)  
**Goal:** AI agent bisa record user actions dan generate executable test  
**Target:** 200 paying customers, 5,000+ tests recorded

---

## Overview

### What We're Building

System yang bisa:
1. Record user interactions (Chrome extension)
2. Capture DOM events (click, input, navigate, scroll)
3. AI processes recording:
   - Intent classification (authentication, search, form submission)
   - Selector optimization (stability scoring 0-100)
   - Natural language generation (human-readable descriptions)
   - Assertion inference (what should be verified)
4. Generate structured test (JSON + Markdown)
5. Export to multiple frameworks (Playwright, Cypress, Selenium)
6. Save to test library (reusable, searchable)

### User Workflow

```
1. User clicks "Record Test"
2. Chrome extension starts recording
3. User interacts with web app (click, type, navigate)
4. Extension captures DOM events + context
5. User clicks "Stop Recording"
6. AI processes recording:
   - Classify intent (login, search, form submission)
   - Optimize selectors (stability scoring)
   - Generate natural language descriptions
   - Infer assertions (what should be verified)
7. System generates structured test (JSON + Markdown)
8. User reviews & edits test
9. User exports to Playwright/Cypress/Selenium
10. User saves to test library (reusable)
```

### Success Criteria

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

---

## Sprint 7-8 (Weeks 13-16): Chrome Extension + Backend

### Week 13-14: Chrome Extension

**Goal:** Build Chrome extension (Manifest V3) that captures user interactions

**Tasks:**

#### Task 7.1: Setup Chrome Extension (Manifest V3)

**Description:** Create Chrome extension with Manifest V3 configuration

**Deliverables:**
- `chrome-extension/manifest.json` — Manifest V3 configuration
- `chrome-extension/background.js` — Service worker
- `chrome-extension/popup.html` — Popup UI (start/stop recording)
- `chrome-extension/popup.js` — Popup logic
- `chrome-extension/content.js` — Content script (DOM event capture)
- `chrome-extension/icon.png` — Extension icon

**Technical Specifications:**

```json
// manifest.json
{
  "manifest_version": 3,
  "name": "GoTest Agent Recorder",
  "version": "1.0.0",
  "description": "Record user interactions for test generation",
  "permissions": [
    "activeTab",
    "storage",
    "scripting"
  ],
  "action": {
    "default_popup": "popup.html",
    "default_icon": "icon.png"
  },
  "background": {
    "service_worker": "background.js"
  },
  "content_scripts": [
    {
      "matches": ["<all_urls>"],
      "js": ["content.js"],
      "run_at": "document_idle"
    }
  ]
}
```

**Acceptance Criteria:**
- Extension loads in Chrome (chrome://extensions)
- Popup displays when extension icon clicked
- Content script injects into web pages
- No console errors

**Testing:**
```bash
# Load extension in Chrome
# 1. Go to chrome://extensions
# 2. Enable "Developer mode"
# 3. Click "Load unpacked"
# 4. Select chrome-extension/ directory
# 5. Verify extension appears in toolbar
```

**Success Criteria:**
- Extension loads without errors
- Popup displays correctly
- Content script injects into pages

---

#### Task 7.2: Capture DOM Events (click, input, navigate)

**Description:** Implement DOM event capture in content script

**Deliverables:**
- `chrome-extension/content.js` — Event listeners for click, input, navigate, scroll

**Technical Specifications:**

```javascript
// content.js
class EventRecorder {
  constructor() {
    this.events = [];
    this.isRecording = false;
  }

  start() {
    this.isRecording = true;
    this.events = [];
    
    // Capture click events
    document.addEventListener('click', this.handleEvent.bind(this), true);
    document.addEventListener('input', this.handleEvent.bind(this), true);
    document.addEventListener('scroll', this.handleEvent.bind(this), true);
    
    // Capture navigation (SPA)
    const originalPushState = history.pushState;
    history.pushState = (...args) => {
      this.handleEvent({ type: 'navigate', url: args[2] });
      return originalPushState.apply(history, args);
    };
    
    // Capture page load
    window.addEventListener('load', () => {
      this.handleEvent({ type: 'navigate', url: window.location.href });
    });
  }

  stop() {
    this.isRecording = false;
    // Remove event listeners
    document.removeEventListener('click', this.handleEvent, true);
    document.removeEventListener('input', this.handleEvent, true);
    document.removeEventListener('scroll', this.handleEvent, true);
    
    return this.events;
  }

  handleEvent(event) {
    if (!this.isRecording) return;

    const eventObj = {
      timestamp: Date.now(),
      type: event.type,
      target: this.serializeTarget(event.target),
      value: event.target.value || null,
      url: window.location.href
    };

    this.events.push(eventObj);
    
    // Send to background script
    chrome.runtime.sendMessage({
      action: 'recordEvent',
      event: eventObj
    });
  }

  serializeTarget(target) {
    return {
      tagName: target.tagName.toLowerCase(),
      id: target.id || null,
      className: target.className || null,
      type: target.type || null,
      placeholder: target.placeholder || null,
      ariaLabel: target.getAttribute('aria-label') || null,
      dataTestId: target.getAttribute('data-testid') || null,
      textContent: target.textContent?.trim() || null,
      // Generate selector candidates
      selectors: this.generateSelectors(target)
    };
  }

  generateSelectors(target) {
    const selectors = [];
    
    // Priority 1: data-testid
    if (target.getAttribute('data-testid')) {
      selectors.push(`[data-testid="${target.getAttribute('data-testid')}"]`);
    }
    
    // Priority 2: aria-label
    if (target.getAttribute('aria-label')) {
      selectors.push(`[aria-label="${target.getAttribute('aria-label')}"]`);
    }
    
    // Priority 3: ID
    if (target.id) {
      selectors.push(`#${target.id}`);
    }
    
    // Priority 4: role + text
    if (target.getAttribute('role') && target.textContent) {
      selectors.push(`[role="${target.getAttribute('role')}"][textContent="${target.textContent.trim()}"]`);
    }
    
    // Priority 5: CSS selector (fallback)
    selectors.push(this.generateCSSSelector(target));
    
    return selectors;
  }

  generateCSSSelector(target) {
    // Simple CSS selector generation
    // Priority: id > class > tag
    if (target.id) {
      return `#${target.id}`;
    }
    
    let selector = target.tagName.toLowerCase();
    if (target.className) {
      const classes = target.className.split(' ').filter(c => c.trim()).slice(0, 2);
      selector += '.' + classes.join('.');
    }
    
    return selector;
  }
}

// Initialize recorder
const recorder = new EventRecorder();

// Listen for messages from popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'startRecording') {
    recorder.start();
    sendResponse({ status: 'recording' });
  } else if (message.action === 'stopRecording') {
    const events = recorder.stop();
    sendResponse({ status: 'stopped', events });
  }
});
```

**Acceptance Criteria:**
- Click events captured with target information
- Input events captured with value
- Navigation events captured (SPA + page load)
- Scroll events captured
- Events sent to background script

**Testing:**
```javascript
// Manual test
// 1. Start recording via popup
// 2. Navigate to https://example.com
// 3. Click links, fill forms, scroll
// 4. Stop recording
// 5. Verify events captured (check console)
```

**Success Criteria:**
- All event types captured
- Target serialization includes selectors
- Events sent to background script

---

#### Task 7.3: Selector Generation (data-testid > aria-label > ID > CSS)

**Description:** Implement robust selector generation with priority ordering

**Deliverables:**
- `chrome-extension/selector-generator.js` — Selector generation logic

**Technical Specifications:**

```javascript
// selector-generator.js
class SelectorGenerator {
  generate(target) {
    const candidates = [];
    
    // Priority 1: data-testid (most stable)
    const dataTestId = target.getAttribute('data-testid');
    if (dataTestId) {
      candidates.push({
        selector: `[data-testid="${dataTestId}"]`,
        type: 'data-testid',
        stability: 100
      });
    }
    
    // Priority 2: aria-label (accessibility)
    const ariaLabel = target.getAttribute('aria-label');
    if (ariaLabel) {
      candidates.push({
        selector: `[aria-label="${ariaLabel}"]`,
        type: 'aria-label',
        stability: 90
      });
    }
    
    // Priority 3: role + text content
    const role = target.getAttribute('role');
    const textContent = target.textContent?.trim();
    if (role && textContent) {
      candidates.push({
        selector: `${target.tagName.toLowerCase()}[role="${role}"]`,
        type: 'role',
        stability: 80,
        textContent: textContent
      });
    }
    
    // Priority 4: ID (if unique and not auto-generated)
    const id = target.id;
    if (id && !this.isAutoGeneratedId(id)) {
      candidates.push({
        selector: `#${id}`,
        type: 'id',
        stability: 70
      });
    }
    
    // Priority 5: placeholder (for inputs)
    const placeholder = target.placeholder;
    if (placeholder && target.tagName === 'INPUT') {
      candidates.push({
        selector: `input[placeholder="${placeholder}"]`,
        type: 'placeholder',
        stability: 75
      });
    }
    
    // Priority 6: CSS selector (fallback)
    const cssSelector = this.generateCSSSelector(target);
    if (cssSelector) {
      candidates.push({
        selector: cssSelector,
        type: 'css',
        stability: 50
      });
    }
    
    // Sort by stability (descending)
    candidates.sort((a, b) => b.stability - a.stability);
    
    return candidates;
  }

  isAutoGeneratedId(id) {
    // Detect auto-generated IDs (e.g., "react-123", "ember-456")
    const patterns = [
      /^react-\d+$/,
      /^ember-\d+$/,
      /^vue-\d+$/,
      /^:r[0-9a-z]+:$/,
      /^rc-[a-z]+-\d+$/
    ];
    
    return patterns.some(pattern => pattern.test(id));
  }

  generateCSSSelector(target) {
    // Generate CSS selector with fallback strategy
    // Strategy: tag > class > nth-child
    
    let selector = target.tagName.toLowerCase();
    
    // Add classes (max 2)
    if (target.className && typeof target.className === 'string') {
      const classes = target.className
        .split(' ')
        .filter(c => c.trim() && !this.isDynamicClass(c))
        .slice(0, 2);
      
      if (classes.length > 0) {
        selector += '.' + classes.join('.');
      }
    }
    
    // Add nth-child if needed (for disambiguation)
    const parent = target.parentElement;
    if (parent) {
      const siblings = Array.from(parent.children).filter(
        child => child.tagName === target.tagName
      );
      
      if (siblings.length > 1) {
        const index = siblings.indexOf(target) + 1;
        selector += `:nth-child(${index})`;
      }
    }
    
    return selector;
  }

  isDynamicClass(className) {
    // Detect dynamic classes (e.g., "css-123abc", "sc-abc123")
    const patterns = [
      /^css-[a-z0-9]+$/,
      /^sc-[a-z0-9]+$/,
      /^[a-z]+-[a-f0-9]{6,}$/,
      /^_[a-z0-9]+$/
    ];
    
    return patterns.some(pattern => pattern.test(className));
  }
}
```

**Acceptance Criteria:**
- Selectors generated with priority ordering
- Auto-generated IDs detected and excluded
- Dynamic classes detected and excluded
- CSS selector fallback works

**Testing:**
```javascript
// Test cases
const testCases = [
  {
    target: { tagName: 'BUTTON', getAttribute: (attr) => attr === 'data-testid' ? 'submit-btn' : null },
    expected: '[data-testid="submit-btn"]'
  },
  {
    target: { tagName: 'INPUT', placeholder: 'Enter email' },
    expected: 'input[placeholder="Enter email"]'
  }
];
```

**Success Criteria:**
- Selectors follow priority ordering
- Stability scores accurate
- Auto-generated IDs excluded

---

#### Task 7.4: WebSocket Communication with Backend

**Description:** Implement WebSocket communication between Chrome extension and backend

**Deliverables:**
- `chrome-extension/background.js` — WebSocket client
- `internal/api/websocket.go` — WebSocket server endpoint

**Technical Specifications:**

```javascript
// background.js (WebSocket client)
class WebSocketClient {
  constructor() {
    this.ws = null;
    this.isConnected = false;
  }

  connect(url) {
    this.ws = new WebSocket(url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.isConnected = true;
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log('Received:', message);
      
      // Handle messages from backend
      if (message.action === 'processComplete') {
        // Notify popup
        chrome.runtime.sendMessage({
          action: 'processingComplete',
          result: message.result
        });
      }
    };
    
    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.isConnected = false;
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  send(message) {
    if (this.isConnected) {
      this.ws.send(JSON.stringify(message));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// Initialize WebSocket client
const wsClient = new WebSocketClient();

// Listen for messages from content script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'connect') {
    wsClient.connect(message.url);
    sendResponse({ status: 'connecting' });
  } else if (message.action === 'disconnect') {
    wsClient.disconnect();
    sendResponse({ status: 'disconnected' });
  } else if (message.action === 'sendEvents') {
    wsClient.send({
      action: 'processEvents',
      events: message.events
    });
    sendResponse({ status: 'sent' });
  }
});
```

```go
// internal/api/websocket.go
package api

import (
  "encoding/json"
  "log"
  "net/http"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  CheckOrigin: func(r *http.Request) bool {
    return true // Allow all origins for development
  },
}

// WebSocketEvent represents a WebSocket message
type WebSocketEvent struct {
  Action string      `json:"action"`
  Data   interface{} `json:"data"`
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Printf("WebSocket upgrade failed: %v", err)
    return
  }
  defer conn.Close()

  log.Println("WebSocket client connected")

  for {
    var event WebSocketEvent
    err := conn.ReadJSON(&event)
    if err != nil {
      log.Printf("WebSocket read error: %v", err)
      break
    }

    log.Printf("Received event: %s", event.Action)

    // Handle different actions
    switch event.Action {
    case "processEvents":
      // Process recorded events
      events, ok := event.Data.([]interface{})
      if !ok {
        log.Printf("Invalid events format")
        continue
      }

      // Send to AI processing pipeline
      result, err := s.processRecording(events)
      if err != nil {
        log.Printf("Processing failed: %v", err)
        conn.WriteJSON(WebSocketEvent{
          Action: "error",
          Data:   err.Error(),
        })
        continue
      }

      // Send result back to extension
      conn.WriteJSON(WebSocketEvent{
        Action: "processComplete",
        Data:   result,
      })
    }
  }
}
```

**Acceptance Criteria:**
- WebSocket connection established
- Events sent from extension to backend
- Processing results sent back to extension
- Error handling works

**Testing:**
```bash
# Manual test
# 1. Start backend server
# 2. Connect WebSocket from extension
# 3. Send test events
# 4. Verify processing result received
```

**Success Criteria:**
- WebSocket connection stable
- Bidirectional communication works
- Error handling robust

---

#### Task 7.5: Basic UI (start/stop recording)

**Description:** Implement popup UI for start/stop recording

**Deliverables:**
- `chrome-extension/popup.html` — Popup HTML
- `chrome-extension/popup.js` — Popup logic
- `chrome-extension/popup.css` — Popup styles

**Technical Specifications:**

```html
<!-- popup.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GoTest Agent Recorder</title>
  <link rel="stylesheet" href="popup.css">
</head>
<body>
  <div class="container">
    <h1>GoTest Agent</h1>
    
    <div id="status" class="status">Ready</div>
    
    <button id="startBtn" class="btn btn-primary">Start Recording</button>
    <button id="stopBtn" class="btn btn-danger" disabled>Stop Recording</button>
    
    <div id="events" class="events" style="display: none;">
      <h2>Recorded Events</h2>
      <div id="eventList"></div>
    </div>
    
    <button id="processBtn" class="btn btn-success" style="display: none;">Process Recording</button>
  </div>
  
  <script src="popup.js"></script>
</body>
</html>
```

```javascript
// popup.js
let isRecording = false;
let events = [];

const startBtn = document.getElementById('startBtn');
const stopBtn = document.getElementById('stopBtn');
const processBtn = document.getElementById('processBtn');
const statusDiv = document.getElementById('status');
const eventsDiv = document.getElementById('events');
const eventList = document.getElementById('eventList');

startBtn.addEventListener('click', async () => {
  isRecording = true;
  events = [];
  
  startBtn.disabled = true;
  stopBtn.disabled = false;
  processBtn.style.display = 'none';
  eventsDiv.style.display = 'none';
  
  statusDiv.textContent = 'Recording...';
  statusDiv.className = 'status recording';
  
  // Send message to content script
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  chrome.tabs.sendMessage(tab.id, { action: 'startRecording' });
});

stopBtn.addEventListener('click', async () => {
  isRecording = false;
  
  startBtn.disabled = false;
  stopBtn.disabled = true;
  processBtn.style.display = 'block';
  eventsDiv.style.display = 'block';
  
  statusDiv.textContent = 'Stopped';
  statusDiv.className = 'status stopped';
  
  // Send message to content script
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  chrome.tabs.sendMessage(tab.id, { action: 'stopRecording' }, (response) => {
    events = response.events;
    displayEvents();
  });
});

processBtn.addEventListener('click', async () => {
  statusDiv.textContent = 'Processing...';
  statusDiv.className = 'status processing';
  
  // Send events to backend via WebSocket
  chrome.runtime.sendMessage({
    action: 'sendEvents',
    events: events
  });
});

// Listen for messages from background script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'processingComplete') {
    statusDiv.textContent = 'Processing complete!';
    statusDiv.className = 'status complete';
    
    // Display result
    displayResult(message.result);
  }
});

function displayEvents() {
  eventList.innerHTML = '';
  
  events.forEach((event, index) => {
    const eventDiv = document.createElement('div');
    eventDiv.className = 'event';
    eventDiv.textContent = `${index + 1}. ${event.type} on ${event.target.tagName}`;
    eventList.appendChild(eventDiv);
  });
}

function displayResult(result) {
  // Display processed test result
  console.log('Result:', result);
}
```

```css
/* popup.css */
body {
  width: 300px;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

.container {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

h1 {
  font-size: 20px;
  margin: 0;
}

.status {
  padding: 10px;
  border-radius: 5px;
  font-weight: bold;
  text-align: center;
}

.status.recording {
  background-color: #fee;
  color: #c33;
}

.status.stopped {
  background-color: #efe;
  color: #3c3;
}

.status.processing {
  background-color: #ffe;
  color: #cc3;
}

.status.complete {
  background-color: #efe;
  color: #3c3;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  font-weight: bold;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-danger {
  background-color: #dc3545;
  color: white;
}

.btn-success {
  background-color: #28a745;
  color: white;
}

.events {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #ddd;
  border-radius: 5px;
  padding: 10px;
}

.event {
  padding: 5px;
  border-bottom: 1px solid #eee;
  font-size: 12px;
}

.event:last-child {
  border-bottom: none;
}
```

**Acceptance Criteria:**
- Popup displays correctly
- Start/stop buttons work
- Events displayed after recording
- Process button triggers processing

**Testing:**
```bash
# Manual test
# 1. Open extension popup
# 2. Click "Start Recording"
# 3. Interact with page
# 4. Click "Stop Recording"
# 5. Verify events displayed
# 6. Click "Process Recording"
# 7. Verify processing complete
```

**Success Criteria:**
- UI responsive
- Recording workflow complete
- Events displayed correctly

---

**Week 15-16: Backend**

**Goal:** Implement backend services for recording

**Tasks:**

#### Task 7.6: Recorder Service (receive events, store in DB)

**Description:** Implement service to receive and store recording events

**Deliverables:**
- `internal/recorder/service.go` — Recorder service
- `internal/recorder/types.go` — Recording types

**Technical Specifications:**

```go
// internal/recorder/types.go
package recorder

import "time"

// RecordingSession represents a recording session
type RecordingSession struct {
  ID        string    `json:"id" db:"id"`
  UserID    string    `json:"user_id" db:"user_id"`
  URL       string    `json:"url" db:"url"`
  Events    []Event   `json:"events" db:"events"`
  Status    string    `json:"status" db:"status"` // recording, stopped, processing, complete
  CreatedAt time.Time `json:"created_at" db:"created_at"`
  UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Event represents a recorded DOM event
type Event struct {
  Timestamp int64         `json:"timestamp"`
  Type      string        `json:"type"` // click, input, navigate, scroll
  Target    EventTarget   `json:"target"`
  Value     string        `json:"value,omitempty"`
  URL       string        `json:"url"`
}

// EventTarget represents the target of a DOM event
type EventTarget struct {
  TagName     string            `json:"tagName"`
  ID          string            `json:"id,omitempty"`
  ClassName   string            `json:"className,omitempty"`
  Type        string            `json:"type,omitempty"`
  Placeholder string            `json:"placeholder,omitempty"`
  AriaLabel   string            `json:"ariaLabel,omitempty"`
  DataTestId  string            `json:"dataTestId,omitempty"`
  TextContent string            `json:"textContent,omitempty"`
  Selectors   []SelectorCandidate `json:"selectors"`
}

// SelectorCandidate represents a selector candidate with stability score
type SelectorCandidate struct {
  Selector  string `json:"selector"`
  Type      string `json:"type"` // data-testid, aria-label, id, css
  Stability int    `json:"stability"` // 0-100
}
```

```go
// internal/recorder/service.go
package recorder

import (
  "context"
  "database/sql"
  "encoding/json"
  "time"

  "github.com/google/uuid"
)

// Service handles recording operations
type Service struct {
  db *sql.DB
}

// NewService creates a new recorder service
func NewService(db *sql.DB) *Service {
  return &Service{db: db}
}

// CreateSession creates a new recording session
func (s *Service) CreateSession(ctx context.Context, userID, url string) (*RecordingSession, error) {
  session := &RecordingSession{
    ID:        uuid.New().String(),
    UserID:    userID,
    URL:       url,
    Events:    []Event{},
    Status:    "recording",
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
  }

  // Store in database
  eventsJSON, _ := json.Marshal(session.Events)
  _, err := s.db.ExecContext(ctx,
    `INSERT INTO recording_sessions (id, user_id, url, events, status, created_at, updated_at)
     VALUES ($1, $2, $3, $4, $5, $6, $7)`,
    session.ID, session.UserID, session.URL, eventsJSON, session.Status,
    session.CreatedAt, session.UpdatedAt)

  if err != nil {
    return nil, err
  }

  return session, nil
}

// AddEvent adds an event to a recording session
func (s *Service) AddEvent(ctx context.Context, sessionID string, event Event) error {
  // Get current session
  var eventsJSON []byte
  err := s.db.QueryRowContext(ctx,
    `SELECT events FROM recording_sessions WHERE id = $1`,
    sessionID).Scan(&eventsJSON)

  if err != nil {
    return err
  }

  // Parse events
  var events []Event
  if err := json.Unmarshal(eventsJSON, &events); err != nil {
    return err
  }

  // Add new event
  events = append(events, event)

  // Update database
  eventsJSON, _ = json.Marshal(events)
  _, err = s.db.ExecContext(ctx,
    `UPDATE recording_sessions SET events = $1, updated_at = $2 WHERE id = $3`,
    eventsJSON, time.Now(), sessionID)

  return err
}

// StopSession stops a recording session
func (s *Service) StopSession(ctx context.Context, sessionID string) error {
  _, err := s.db.ExecContext(ctx,
    `UPDATE recording_sessions SET status = 'stopped', updated_at = $1 WHERE id = $2`,
    time.Now(), sessionID)

  return err
}

// GetSession retrieves a recording session
func (s *Service) GetSession(ctx context.Context, sessionID string) (*RecordingSession, error) {
  var session RecordingSession
  var eventsJSON []byte

  err := s.db.QueryRowContext(ctx,
    `SELECT id, user_id, url, events, status, created_at, updated_at
     FROM recording_sessions WHERE id = $1`,
    sessionID).Scan(
    &session.ID, &session.UserID, &session.URL, &eventsJSON,
    &session.Status, &session.CreatedAt, &session.UpdatedAt)

  if err != nil {
    return nil, err
  }

  // Parse events
  if err := json.Unmarshal(eventsJSON, &session.Events); err != nil {
    return nil, err
  }

  return &session, nil
}

// ListSessions lists all recording sessions for a user
func (s *Service) ListSessions(ctx context.Context, userID string) ([]*RecordingSession, error) {
  rows, err := s.db.QueryContext(ctx,
    `SELECT id, user_id, url, events, status, created_at, updated_at
     FROM recording_sessions WHERE user_id = $1 ORDER BY created_at DESC`,
    userID)

  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var sessions []*RecordingSession
  for rows.Next() {
    var session RecordingSession
    var eventsJSON []byte

    err := rows.Scan(
      &session.ID, &session.UserID, &session.URL, &eventsJSON,
      &session.Status, &session.CreatedAt, &session.UpdatedAt)

    if err != nil {
      return nil, err
    }

    // Parse events
    if err := json.Unmarshal(eventsJSON, &session.Events); err != nil {
      return nil, err
    }

    sessions = append(sessions, &session)
  }

  return sessions, nil
}
```

**Acceptance Criteria:**
- Sessions created successfully
- Events added to sessions
- Sessions retrieved correctly
- Session list works

**Testing:**
```go
func TestRecorderService(t *testing.T) {
  // Create test database
  db := setupTestDB(t)
  defer db.Close()

  service := NewService(db)

  // Test create session
  session, err := service.CreateSession(context.Background(), "user-1", "https://example.com")
  if err != nil {
    t.Fatalf("CreateSession failed: %v", err)
  }

  // Test add event
  event := Event{
    Timestamp: time.Now().Unix(),
    Type:      "click",
    Target: EventTarget{
      TagName: "button",
      ID:      "submit-btn",
    },
  }

  err = service.AddEvent(context.Background(), session.ID, event)
  if err != nil {
    t.Fatalf("AddEvent failed: %v", err)
  }

  // Test get session
  retrieved, err := service.GetSession(context.Background(), session.ID)
  if err != nil {
    t.Fatalf("GetSession failed: %v", err)
  }

  if len(retrieved.Events) != 1 {
    t.Fatalf("Expected 1 event, got %d", len(retrieved.Events))
  }
}
```

**Success Criteria:**
- All CRUD operations work
- JSON serialization/deserialization correct
- Performance acceptable (<100ms per operation)

---

*(Continuing with remaining tasks in Sprint 7-8 and Sprint 9-12...)*

**Note:** Due to document length, I'll create separate files for:
- `docs/PHASE-2-PLAN.md` (this file - Chrome Extension + Backend + AI Processing)
- `docs/PHASE-2-EXPORT.md` (Export + Library)
- `docs/PHASE-2-TASK-SPECIFICATIONS.md` (Detailed task specs)

Would you like me to continue with the detailed task specifications for the remaining tasks in Phase 2, or move on to creating the Phase 3 plan document?
