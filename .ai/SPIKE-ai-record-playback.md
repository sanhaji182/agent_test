# Technical Spike Plan: AI-Powered Record & Playback

**Status:** In Progress  
**Started:** 2026-07-30  
**Ends:** 2026-08-13 (2 weeks)  
**Owner:** Engineering Team  
**Budget:** ~$15k (2 engineer-weeks + LLM API costs)

---

## 1. Purpose

Validate the 3 riskiest assumptions of the AI Record & Playback feature **before** committing to the 16-week, $200k implementation.

### Riskiest Assumptions

1. **AI can accurately classify user intent** from recorded actions
2. **Selector stability scoring algorithm** picks truly stable selectors
3. **AI-inferred assertions** are actually useful (not noise)

### Kill Criteria

If ANY of these metrics are <60%, we **pivot or kill** the feature:
- Intent classification accuracy: target ≥85%
- Selector stability matches human judgment: target ≥80%
- Inferred assertions rated "useful+": target ≥70%
- Generated Playwright test runs successfully first try: target ≥70%

### Success Criteria

If metrics meet targets → **Green light for Phase 1** (full 16-week build).

---

## 2. Scope

### In Scope (Spike)

- [ ] Minimal Chrome extension (record click/input/navigate, no UI polish)
- [ ] 10 recorded user flows on real websites
- [ ] AI processing pipeline (intent, NLG, selector optimization, assertions)
- [ ] Scoring rubric + human review
- [ ] End-to-end prototype: record → AI process → export Playwright → run
- [ ] LLM comparison: GPT-4 vs Claude vs Llama (quality + cost)
- [ ] Technical approach validation: Chrome ext vs Playwright Inspector

### Out of Scope (Spike)

- [ ] Full frontend UI (recorder page, test viewer, library)
- [ ] Backend API (recording sessions, WebSocket, test storage)
- [ ] Self-healing, test composition, parameterization
- [ ] Team sharing, version control, export formats (Cypress, Selenium)
- [ ] Security hardening, production deployment

---

## 3. Week 1 Schedule (AI Quality Validation)

### Day 1 (Monday): Setup

**Tasks:**
- [ ] Read spec-kit from /tmp/spec-kit/ (user will clone)
- [ ] Create Chrome extension project (Manifest V3)
- [ ] Setup event listeners (click, input, navigate)
- [ ] Basic selector generation (data-testid > aria-label > ID > CSS)
- [ ] Test on localhost: record 1 flow, verify events captured

**Deliverables:**
- Chrome extension loads in Chrome
- Can record 1 simple flow (navigate + click + input)
- Events logged to console

**Time estimate:** 4 hours

---

### Day 2 (Tuesday): Build Recorder

**Tasks:**
- [ ] Improve selector generation (handle edge cases)
- [ ] Capture more event types (scroll, hover, select, wait)
- [ ] Export events to JSON file
- [ ] Test on 3 different websites
- [ ] Record 3 flows manually

**Deliverables:**
- 3 recorded flows (JSON format)
- Selector generation works on diverse DOM structures

**Time estimate:** 6 hours

---

### Day 3 (Wednesday): Record 7 More Flows

**Tasks:**
- [ ] Record 7 additional flows (total: 10)
- [ ] Aim for diversity:
  - 3x authentication (login, signup, password reset)
  - 2x search (product search, filter)
  - 2x form submission (contact, checkout)
  - 2x navigation (menu, breadcrumbs)
  - 1x payment flow
- [ ] Document each flow (what user did, expected intent, expected assertions)

**Deliverables:**
- 10 recorded flows (JSON)
- Ground truth document (expected intent, selectors, assertions per flow)

**Time estimate:** 6 hours

---

### Day 4 (Thursday): AI Pipeline

**Tasks:**
- [ ] Setup LLM API calls (GPT-4, Claude, Llama)
- [ ] Implement 4 AI prompts from design doc:
  - Intent classification
  - Natural language generation
  - Selector optimization (scoring algorithm)
  - Assertion inference
- [ ] Process all 10 flows through AI
- [ ] Store results in JSON

**Deliverables:**
- AI outputs for all 10 flows (3 LLMs each)
- Cost tracking (tokens, $ per LLM)

**Time estimate:** 8 hours

---

### Day 5 (Friday): Scoring + Analysis

**Tasks:**
- [ ] Score AI outputs using rubric (see Section 4)
- [ ] Compare LLMs (accuracy, cost, latency)
- [ ] Analyze failures (where did AI get it wrong?)
- [ ] Calculate metrics:
  - Intent accuracy
  - Selector stability accuracy
  - Assertion usefulness
- [ ] Draft spike report (preliminary)

**Deliverables:**
- Scored results (spreadsheet)
- Preliminary spike report
- LLM recommendation

**Time estimate:** 6 hours

---

## 4. Week 2 Schedule (Technical Validation)

### Day 6 (Monday): Backend Prototype

**Tasks:**
- [ ] Setup Go project (use existing `gotest-agent` structure)
- [ ] Database schema (recording_sessions, recording_events)
- [ ] REST API: POST /recordings/start, POST /recordings/:id/stop
- [ ] WebSocket server: /recordings/:id/stream
- [ ] Event storage logic

**Deliverables:**
- Backend API running locally
- Can start/stop recording via curl
- Events stored in PostgreSQL

**Time estimate:** 6 hours

---

### Day 7 (Tuesday): Chrome Extension Integration

**Tasks:**
- [ ] Connect Chrome extension to backend (WebSocket)
- [ ] Real-time event streaming
- [ ] Backend stores events as they arrive
- [ ] Test end-to-end: record in Chrome → events in DB

**Deliverables:**
- Working Chrome → Backend pipeline
- 1 flow recorded end-to-end

**Time estimate:** 6 hours

---

### Day 8 (Wednesday): AI Integration

**Tasks:**
- [ ] Trigger AI processing when recording stops
- [ ] Store structured test (JSONB) in database
- [ ] Generate Playwright code from structured test
- [ ] Save exported code to file

**Deliverables:**
- Full pipeline: record → AI → export
- 1 Playwright test file generated

**Time estimate:** 6 hours

---

### Day 9 (Thursday): End-to-End Test

**Tasks:**
- [ ] Record 5 flows on real websites (using full pipeline)
- [ ] Export to Playwright
- [ ] Actually run the generated tests
- [ ] Measure:
  - Tests that pass on first try
  - Tests that need manual fixes
  - Common failure modes

**Deliverables:**
- 5 recorded flows
- 5 exported Playwright tests
- Execution results (pass/fail)
- Failure analysis

**Time estimate:** 6 hours

---

### Day 10 (Friday): Final Report + Decision

**Tasks:**
- [ ] Compile all metrics
- [ ] Write final spike report
- [ ] Present findings to stakeholders
- [ ] Make decision: proceed / pivot / kill

**Deliverables:**
- Final spike report (10-15 pages)
- Decision document
- Next steps (if proceeding)

**Time estimate:** 6 hours

---

## 5. Scoring Rubric

### Intent Classification (0-5 scale)

| Score | Description | Example |
|-------|-------------|---------|
| 5 | Perfect — correctly identified intent + all nuances | User did login + remember me → AI: "authentication with persistent session" |
| 4 | Good — correct intent, minor details missed | User did login → AI: "authentication" (missed remember me) |
| 3 | Acceptable — general intent correct, vague | User did login → AI: "form submission" |
| 2 | Poor — wrong intent, some context | User did login → AI: "search" |
| 1 | Very poor — completely wrong | User did login → AI: "navigation" |
| 0 | Nonsensical | User did login → AI: "payment" |

**Target:** ≥85% of flows score 4 or 5

### Selector Stability (Match Human Judgment)

**Method:** For each recorded action, human reviewer picks the most stable selector. Compare to AI's top-ranked selector.

**Metric:** % of cases where AI's top choice matches human's top choice

**Target:** ≥80% match rate

### Assertion Usefulness (0-5 scale)

| Score | Description | Example |
|-------|-------------|---------|
| 5 | Critical — would definitely include in real test | "Verify redirect to /dashboard" after login |
| 4 | Useful — good to have | "Verify welcome message appears" |
| 3 | Nice to have — optional | "Verify page title changes" |
| 2 | Low value — obvious or redundant | "Verify button was clicked" |
| 1 | Noise — not meaningful | "Verify mouse moved" |
| 0 | Wrong — incorrect assertion | "Verify user logged out" (after login) |

**Target:** ≥70% of assertions score 3 or higher

### Generated Test Quality (Binary)

| Outcome | Description |
|---------|-------------|
| Pass | Generated Playwright test runs successfully on first try |
| Fail | Test fails (wrong selector, timeout, assertion error, etc) |

**Target:** ≥70% pass rate

---

## 6. LLM Comparison Framework

### Test 3 LLMs

1. **OpenAI GPT-4** (via API)
2. **Anthropic Claude 3.5 Sonnet** (via API)
3. **Meta Llama 3.1 405B** (via Groq or Together AI)

### Metrics

| Metric | How to Measure |
|--------|----------------|
| **Accuracy** | Scoring rubric (Section 5) |
| **Latency** | Time from prompt to response (ms) |
| **Cost** | $ per 1000 tokens (input + output) |
| **Consistency** | Run same prompt 5x, measure variance |

### Decision Matrix

| LLM | Accuracy | Latency | Cost | Consistency | Winner? |
|-----|----------|---------|------|-------------|---------|
| GPT-4 | TBD | TBD | TBD | TBD | |
| Claude | TBD | TBD | TBD | TBD | |
| Llama | TBD | TBD | TBD | TBD | |

**Decision rule:** Pick highest accuracy if cost <$10/month per user. Otherwise, pick best accuracy/cost ratio.

---

## 7. Critical Decisions to Make

### Q1: Chrome Extension vs Playwright Inspector

**Decision criteria:**
- Development effort (hours)
- Control over UX
- Selector generation quality
- Integration with our backend

**How to decide:** Try both during Week 1 (Day 1-2). Pick the one that:
1. Records accurately
2. Is easier to integrate
3. Gives us more control

**Fallback:** If both work equally well, pick Chrome extension (more control).

### Q5: LLM Choice

**Decision criteria:**
- Accuracy (scoring rubric)
- Cost ($ per user per month)
- Latency (user experience)
- Consistency (reliability)

**How to decide:** Run all 3 LLMs on all 10 flows. Compare metrics.

**Fallback:** Hybrid approach (GPT-4 for complex tasks, smaller model for simple).

### Q6: Storage Format

**Decision criteria:**
- Flexibility (can store diverse test structures)
- Queryability (can search/filter tests)
- Performance (read/write speed)

**Likely answer:** JSONB in PostgreSQL (already in design doc, spike will confirm).

**How to decide:** If JSONB works well in Week 2, keep it. Otherwise, evaluate MongoDB or separate tables.

---

## 8. Test Websites

### Candidates (to be confirmed by user)

**Simple (1):**
- [ ] example.com (or similar minimal site)

**Medium (1):**
- [ ] E-commerce public page (Tokopedia/Bukalapak product listing)

**Complex (1):**
- [ ] SaaS dashboard demo (Vercel/Linear/Notion public page)

**Specific flows to record:**
1. Login flow (if available)
2. Search + filter
3. Add to cart / wishlist
4. Checkout (if public)
5. Contact form submission
6. Newsletter signup
7. Comment/review submission
8. Navigation (menu, breadcrumbs)
9. Pagination
10. Modal/popup interaction

---

## 9. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| AI accuracy too low (<60%) | Medium | High (kill feature) | Pivot to simpler approach (rule-based) or kill feature |
| LLM costs too high | Medium | Medium | Use hybrid approach (small model for simple tasks) |
| Chrome extension too hard to build | Low | Medium | Switch to Playwright Inspector wrapper |
| Generated tests don't run | Medium | High | Improve selector generation, add fallback selectors |
| Not enough time (2 weeks too short) | Low | Medium | Extend by 1 week if needed (still cheaper than 16 weeks) |

---

## 10. Deliverables

### Week 1 Deliverables

- [ ] Minimal Chrome extension (functional, not polished)
- [ ] 10 recorded user flows (JSON format)
- [ ] Ground truth document (expected outputs)
- [ ] AI outputs for all 10 flows (3 LLMs)
- [ ] Scored results (spreadsheet)
- [ ] Preliminary spike report

### Week 2 Deliverables

- [ ] Backend prototype (API + DB + WebSocket)
- [ ] Chrome extension integrated with backend
- [ ] AI processing pipeline integrated
- [ ] 5 end-to-end recorded flows
- [ ] 5 exported Playwright tests
- [ ] Execution results (pass/fail)
- [ ] **Final spike report** (10-15 pages)
- [ ] **Decision document** (proceed / pivot / kill)

---

## 11. Success Metrics Summary

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Intent classification accuracy | ≥85% | TBD | ⏳ |
| Selector stability match | ≥80% | TBD | ⏳ |
| Assertion usefulness (score ≥3) | ≥70% | TBD | ⏳ |
| Generated test pass rate | ≥70% | TBD | ⏳ |
| LLM cost per user per month | <$10 | TBD | ⏳ |
| Spike completed on time | 2 weeks | TBD | ⏳ |

---

## 12. Budget

| Item | Cost |
|------|------|
| Engineering time (2 weeks) | $10,000 |
| LLM API costs (GPT-4, Claude, Llama) | $500 |
| Chrome Web Store developer account | $5 |
| Test infrastructure (localhost, free tier) | $0 |
| **Total** | **~$10,505** |

**Compared to:** $150-200k for full implementation without validation.

---

## 13. Decision Framework

### After Week 1 (AI Quality)

**If all 3 AI metrics ≥ target:**
- ✅ Proceed to Week 2 (technical validation)

**If 1-2 metrics below target but >60%:**
- ⚠️ Proceed to Week 2, but adjust design based on findings

**If any metric <60%:**
- 🛑 Stop. Pivot or kill feature.
- Options:
  - Pivot to rule-based approach (less AI, more manual)
  - Pivot to different feature (e.g., test generation from PRD only, no recording)
  - Kill feature, explore other directions

### After Week 2 (Technical Validation)

**If generated test pass rate ≥70%:**
- ✅ Green light for Phase 1 (full 16-week build)

**If pass rate 50-70%:**
- ⚠️ Proceed with Phase 1, but allocate extra time for selector generation improvements

**If pass rate <50%:**
- 🛑 Stop. Re-evaluate technical approach.
- Options:
  - Switch to Playwright Inspector wrapper
  - Improve selector generation algorithm
  - Add manual review step before export

---

## 14. Communication Plan

### Daily Updates

**Format:** 3-bullet Slack message to stakeholders
1. What I did today
2. What I'm doing tomorrow
3. Blockers/risks

### End of Week 1

**Deliverable:** Preliminary spike report (email + Slack)
- AI quality metrics
- LLM comparison
- Key findings
- Go/no-go for Week 2

### End of Week 2

**Deliverable:** Final spike report + decision meeting
- 30-minute presentation to stakeholders
- Full metrics
- Recommendation (proceed/pivot/kill)
- Next steps (if proceeding)

---

## 15. Approval

**Spike approved by:** [Your Name]  
**Date:** 2026-07-30  
**Budget authorized:** $10,505  
**Timeline:** 2 weeks (2026-07-30 to 2026-08-13)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-07-30 | Engineering Team | Initial spike plan |

---

**End of Spike Plan**
