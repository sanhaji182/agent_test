# Phase 2 Implementation Plan: Record & Playback

**Status**: In Progress  
**Date**: July 31, 2026  
**Goal**: Implement Chrome extension for recording user interactions and converting them to Playwright tests

---

## Overview

Phase 2 focuses on building a Chrome extension that allows users to record their interactions with web applications and automatically convert those recordings into executable Playwright tests.

### Key Features
- Chrome extension for recording user interactions
- Backend event capture and storage
- AI-powered test generation from recordings
- Test library management
- Export to multiple frameworks (Playwright, Cypress, Selenium)

---

## Architecture

### Components

```
┌─────────────────────┐
│  Chrome Extension   │
│  (Manifest V3)      │
│  - Event capture    │
│  - DOM recording    │
│  - User interface   │
└──────────┬──────────┘
           │ WebSocket
           ▼
┌─────────────────────┐
│  Backend API        │
│  - Event storage    │
│  - AI processing    │
│  - Test generation  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Test Library       │
│  - Storage          │
│  - Search           │
│  - Export           │
└─────────────────────┘
```

### Database Schema

```sql
-- Recording sessions
CREATE TABLE recording_sessions (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'recording', -- recording, stopped, processing, completed
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Recorded events
CREATE TABLE recorded_events (
    id UUID PRIMARY KEY,
    session_id UUID REFERENCES recording_sessions(id),
    event_type VARCHAR(50) NOT NULL, -- click, input, navigate, scroll, etc.
    selector TEXT,
    value TEXT,
    timestamp BIGINT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Generated tests
CREATE TABLE generated_tests (
    id UUID PRIMARY KEY,
    recording_session_id UUID REFERENCES recording_sessions(id),
    name VARCHAR(255) NOT NULL,
    framework VARCHAR(50) DEFAULT 'playwright',
    code TEXT NOT NULL,
    confidence_score INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Chrome Extension Implementation

### Manifest V3 Configuration

```json
{
  "manifest_version": 3,
  "name": "GoTest Agent Recorder",
  "version": "1.0.0",
  "description": "Record user interactions for automated testing",
  "permissions": [
    "activeTab",
    "storage",
    "scripting"
  ],
  "host_permissions": [
    "<all_urls>"
  ],
  "action": {
    "default_popup": "popup.html",
    "default_icon": {
      "16": "icons/icon16.png",
      "48": "icons/icon48.png",
      "128": "icons/icon128.png"
    }
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

### Content Script (content.js)

```javascript
// Event capture and DOM recording
class EventRecorder {
    constructor() {
        this.events = [];
        this.isRecording = false;
        this.sessionId = null;
    }

    startRecording(sessionId) {
        this.sessionId = sessionId;
        this.isRecording = true;
        this.events = [];
        
        // Add event listeners
        document.addEventListener('click', this.handleClick.bind(this), true);
        document.addEventListener('input', this.handleInput.bind(this), true);
        document.addEventListener('scroll', this.handleScroll.bind(this), true);
        document.addEventListener('keydown', this.handleKeyDown.bind(this), true);
        
        console.log('Recording started');
    }

    stopRecording() {
        this.isRecording = false;
        console.log('Recording stopped');
        return this.events;
    }

    handleClick(event) {
        if (!this.isRecording) return;
        
        const selector = this.generateSelector(event.target);
        const event_data = {
            type: 'click',
            selector: selector,
            timestamp: Date.now(),
            metadata: {
                tagName: event.target.tagName,
                text: event.target.textContent?.trim(),
                href: event.target.href || null
            }
        };
        
        this.events.push(event_data);
        this.sendEvent(event_data);
    }

    handleInput(event) {
        if (!this.isRecording) return;
        
        const selector = this.generateSelector(event.target);
        const event_data = {
            type: 'input',
            selector: selector,
            value: event.target.value,
            timestamp: Date.now(),
            metadata: {
                tagName: event.target.tagName,
                inputType: event.target.type,
                placeholder: event.target.placeholder
            }
        };
        
        this.events.push(event_data);
        this.sendEvent(event_data);
    }

    handleScroll(event) {
        if (!this.isRecording) return;
        
        const event_data = {
            type: 'scroll',
            timestamp: Date.now(),
            metadata: {
                scrollX: window.scrollX,
                scrollY: window.scrollY
            }
        };
        
        this.events.push(event_data);
        this.sendEvent(event_data);
    }

    handleKeyDown(event) {
        if (!this.isRecording) return;
        
        // Only record special keys (Enter, Tab, etc.)
        const specialKeys = ['Enter', 'Tab', 'Escape', 'ArrowUp', 'ArrowDown'];
        if (!specialKeys.includes(event.key)) return;
        
        const selector = this.generateSelector(event.target);
        const event_data = {
            type: 'keydown',
            selector: selector,
            key: event.key,
            timestamp: Date.now(),
            metadata: {
                tagName: event.target.tagName
            }
        };
        
        this.events.push(event_data);
        this.sendEvent(event_data);
    }

    generateSelector(element) {
        // Priority 1: data-testid
        if (element.hasAttribute('data-testid')) {
            return `[data-testid="${element.getAttribute('data-testid')}"]`;
        }
        
        // Priority 2: ID
        if (element.id) {
            return `#${element.id}`;
        }
        
        // Priority 3: aria-label
        if (element.hasAttribute('aria-label')) {
            return `[aria-label="${element.getAttribute('aria-label')}"]`;
        }
        
        // Priority 4: role
        if (element.hasAttribute('role')) {
            return `[role="${element.getAttribute('role')}"]`;
        }
        
        // Fallback: CSS selector
        return this.generateCSSSelector(element);
    }

    generateCSSSelector(element) {
        // Simple CSS selector generation
        const tagName = element.tagName.toLowerCase();
        const classes = Array.from(element.classList).slice(0, 2).join('.');
        
        if (classes) {
            return `${tagName}.${classes}`;
        }
        
        return tagName;
    }

    sendEvent(event_data) {
        // Send to background script
        chrome.runtime.sendMessage({
            type: 'EVENT_RECORDED',
            sessionId: this.sessionId,
            event: event_data
        });
    }
}

// Initialize recorder
const recorder = new EventRecorder();

// Listen for messages from popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'START_RECORDING') {
        recorder.startRecording(message.sessionId);
        sendResponse({ success: true });
    } else if (message.type === 'STOP_RECORDING') {
        const events = recorder.stopRecording();
        sendResponse({ success: true, events: events });
    }
    
    return true; // Keep message channel open
});
```

### Background Script (background.js)

```javascript
// Background service worker for Chrome extension
class BackgroundService {
    constructor() {
        this.sessions = new Map();
        this.websocket = null;
    }

    async startSession(sessionId, projectId) {
        this.sessions.set(sessionId, {
            projectId: projectId,
            startTime: Date.now(),
            events: []
        });
        
        // Connect to backend WebSocket
        this.connectWebSocket(sessionId);
    }

    async stopSession(sessionId) {
        const session = this.sessions.get(sessionId);
        if (!session) return;
        
        // Send all events to backend
        await this.sendEventsToBackend(sessionId, session.events);
        
        // Close WebSocket
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }
        
        this.sessions.delete(sessionId);
    }

    connectWebSocket(sessionId) {
        const wsUrl = `ws://localhost:8080/ws/recording/${sessionId}`;
        this.websocket = new WebSocket(wsUrl);
        
        this.websocket.onmessage = (event) => {
            console.log('WebSocket message:', event.data);
        };
        
        this.websocket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    async sendEventsToBackend(sessionId, events) {
        const response = await fetch(`http://localhost:8080/api/v1/recordings/${sessionId}/events`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ events: events })
        });
        
        if (!response.ok) {
            console.error('Failed to send events:', response.statusText);
        }
    }

    async createRecordingSession(projectId, name) {
        const response = await fetch('http://localhost:8080/api/v1/recordings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                project_id: projectId,
                name: name
            })
        });
        
        if (!response.ok) {
            throw new Error('Failed to create recording session');
        }
        
        const data = await response.json();
        return data.id;
    }
}

const backgroundService = new BackgroundService();

// Listen for messages from content script and popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'START_SESSION') {
        backgroundService.createRecordingSession(message.projectId, message.name)
            .then(sessionId => {
                backgroundService.startSession(sessionId, message.projectId);
                sendResponse({ success: true, sessionId: sessionId });
            })
            .catch(error => {
                sendResponse({ success: false, error: error.message });
            });
    } else if (message.type === 'STOP_SESSION') {
        backgroundService.stopSession(message.sessionId);
        sendResponse({ success: true });
    } else if (message.type === 'EVENT_RECORDED') {
        const session = backgroundService.sessions.get(message.sessionId);
        if (session) {
            session.events.push(message.event);
        }
    }
    
    return true;
});
```

---

## Backend API Endpoints

### Recording Endpoints

```go
// POST /api/v1/recordings
// Create a new recording session
func (s *Server) handleCreateRecording(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ProjectID string `json:"project_id"`
        Name      string `json:"name"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSONError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    // Create recording session in database
    session := &RecordingSession{
        ID:        uuid.New().String(),
        ProjectID: req.ProjectID,
        Name:      req.Name,
        Status:    "recording",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := s.store.CreateRecordingSession(r.Context(), session); err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to create recording")
        return
    }
    
    writeJSON(w, http.StatusCreated, session)
}

// POST /api/v1/recordings/:id/events
// Store recorded events
func (s *Server) handleStoreEvents(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "id")
    
    var req struct {
        Events []RecordedEvent `json:"events"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSONError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    // Store events in database
    for _, event := range req.Events {
        event.ID = uuid.New().String()
        event.SessionID = sessionID
        event.CreatedAt = time.Now()
        
        if err := s.store.CreateRecordedEvent(r.Context(), &event); err != nil {
            writeJSONError(w, http.StatusInternalServerError, "failed to store events")
            return
        }
    }
    
    // Update session status
    if err := s.store.UpdateRecordingSessionStatus(r.Context(), sessionID, "stopped"); err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to update session")
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// POST /api/v1/recordings/:id/generate
// Generate tests from recorded events
func (s *Server) handleGenerateTests(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "id")
    
    // Get recorded events
    events, err := s.store.GetRecordedEvents(r.Context(), sessionID)
    if err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to get events")
        return
    }
    
    // Generate tests using AI
    testCode, err := s.aiService.GenerateTestsFromRecording(r.Context(), events)
    if err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to generate tests")
        return
    }
    
    // Store generated test
    test := &GeneratedTest{
        ID:                 uuid.New().String(),
        RecordingSessionID: sessionID,
        Name:              "Generated Test",
        Framework:         "playwright",
        Code:              testCode,
        ConfidenceScore:   85,
        CreatedAt:         time.Now(),
        UpdatedAt:         time.Now(),
    }
    
    if err := s.store.CreateGeneratedTest(r.Context(), test); err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to store test")
        return
    }
    
    // Update session status
    if err := s.store.UpdateRecordingSessionStatus(r.Context(), sessionID, "completed"); err != nil {
        writeJSONError(w, http.StatusInternalServerError, "failed to update session")
        return
    }
    
    writeJSON(w, http.StatusCreated, test)
}
```

---

## AI Test Generation from Recordings

### Prompt Template

```go
func (s *SynthesisService) GenerateTestsFromRecording(ctx context.Context, events []RecordedEvent) (string, error) {
    prompt := fmt.Sprintf(`You are an expert test automation engineer. Convert the following recorded user interactions into a Playwright test.

RECORDED EVENTS:
%s

Generate a complete Playwright test that:
1. Reproduces the recorded user interactions
2. Uses robust selectors (data-testid, aria-label, role)
3. Includes appropriate waits for page loads and element visibility
4. Adds assertions to verify expected outcomes
5. Uses descriptive test names and comments

Output a complete Playwright test in TypeScript format.

Example output:
\`\`\`typescript
import { test, expect } from '@playwright/test';

test('user login flow', async ({ page }) => {
  // Navigate to login page
  await page.goto('https://example.com/login');
  
  // Wait for login form to be visible
  await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  
  // Fill in login form
  await page.fill('[data-testid="email-input"]', 'user@example.com');
  await page.fill('[data-testid="password-input"]', 'password123');
  
  // Click login button
  await page.click('[data-testid="login-button"]');
  
  // Wait for navigation
  await page.waitForURL('**/dashboard');
  
  // Verify successful login
  await expect(page.locator('[data-testid="welcome-message"]')).toBeVisible();
});
\`\`\`
`, formatEvents(events))
    
    return s.client.GenerateText(ctx, prompt)
}
```

---

## Implementation Checklist

### Chrome Extension
- [ ] Create manifest.json (Manifest V3)
- [ ] Implement content script (event capture)
- [ ] Implement background script (WebSocket communication)
- [ ] Create popup UI (start/stop recording)
- [ ] Add icons
- [ ] Test extension locally

### Backend API
- [ ] Create database schema (recording_sessions, recorded_events, generated_tests)
- [ ] Implement POST /api/v1/recordings endpoint
- [ ] Implement POST /api/v1/recordings/:id/events endpoint
- [ ] Implement POST /api/v1/recordings/:id/generate endpoint
- [ ] Add WebSocket endpoint for real-time communication
- [ ] Test endpoints

### AI Integration
- [ ] Implement GenerateTestsFromRecording method
- [ ] Create prompt templates for test generation
- [ ] Test AI-generated tests
- [ ] Add confidence scoring

### Test Library
- [ ] Implement test library storage
- [ ] Add search functionality
- [ ] Add export functionality (Playwright, Cypress, Selenium)
- [ ] Test library UI

---

## Next Steps

1. **Complete Chrome extension** - Finish manifest, content script, background script
2. **Implement backend API** - Create database schema and endpoints
3. **Test integration** - Test Chrome extension with backend
4. **Add AI test generation** - Implement AI-powered test generation from recordings
5. **Build test library** - Create UI for managing generated tests

---

## Challenges

### Selector Generation
- **Challenge**: Generating robust selectors that work across different websites
- **Solution**: Multi-level selector strategy (data-testid > ID > aria-label > CSS)
- **Testing**: Test with multiple websites and frameworks

### Event Capture
- **Challenge**: Capturing all relevant events without overwhelming the system
- **Solution**: Filter events, debounce rapid events, batch send
- **Testing**: Test with complex user interactions

### AI Test Generation
- **Challenge**: Generating high-quality tests from limited recording data
- **Solution**: Advanced prompt engineering, multi-stage generation
- **Testing**: Test with various recording scenarios

---

## Success Metrics

- ✅ Chrome extension captures user interactions accurately
- ✅ Backend stores events correctly
- ✅ AI generates high-quality tests (80%+ confidence score)
- ✅ Generated tests are executable without modification
- ✅ Test library is searchable and exportable

---

**Status**: 🔄 In Progress - Chrome extension planning complete, backend API planning complete, ready for implementation
