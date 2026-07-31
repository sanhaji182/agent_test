# AI Agent Prompts - GoTest Agent Development

**Purpose:** Ready-to-use prompts for AI agents to continue GoTest Agent development  
**Last Updated:** 2026-07-31  
**Status:** Production Ready

---

## 🚀 Quick Start Prompts

### **Prompt 1: General Continuation (RECOMMENDED)**

```
Lanjutkan development GoTest Agent.

Baca semua dokumentasi di docs/:
- FINAL_SUMMARY.md - Current status dan achievements
- PHASE-2-IMPLEMENTATION.md - Record & Playback implementation guide
- PHASE-3-IMPLEMENTATION.md - Continuous Sync implementation guide
- PHASE-4-IMPLEMENTATION.md - Advanced AI features implementation guide

Identifikasi next high-impact feature yang belum implement berdasarkan priority:
1. High impact + High feasibility → implement first
2. High impact + Low feasibility → implement second
3. Low impact + High feasibility → implement third
4. Low impact + Low feasibility → implement last

Ikuti protocol dari MEMORY.md:
- Investigate first (understand current implementation)
- Preserve architecture (jangan break existing features)
- Small atomic commits
- Verify thoroughly (run tests)
- Document di .ai/CHANGELOG_AI.md

Berikan progress report setelah setiap milestone. Jika stuck, tanya sebelum lanjut.
```

### **Prompt 2: Autonomous Full Development**

```
Lanjutkan autonomous development GoTest Agent sampai semua Phase 2-4 features selesai.

Baca semua dokumentasi:
- docs/FINAL_SUMMARY.md
- docs/PHASE-2-IMPLEMENTATION.md
- docs/PHASE-3-IMPLEMENTATION.md
- docs/PHASE-4-IMPLEMENTATION.md
- docs/MANUAL_SETUP_GUIDE.md

Implement semua unfinished features berdasarkan priority matrix:

Phase 2 (Record & Playback):
- Chrome extension (Manifest V3)
- Backend API untuk recorded events
- Test library management

Phase 3 (Continuous Sync):
- GitHub webhook integration
- Drift detection service
- Auto-generation service
- Alert system

Phase 4 (Advanced AI):
- Ruby Rails parser
- Java Spring Boot parser
- C# ASP.NET Core parser
- AI-driven test optimization

Untuk setiap feature:
1. Investigate current state
2. Design solution
3. Implement dengan small atomic commits
4. Test thoroughly (go test ./...)
5. Document di .ai/CHANGELOG_AI.md
6. Commit dan push

Berikan progress report setelah setiap feature. Jika stuck, tanya sebelum lanjut.
```

---

## 🎯 Phase-Specific Prompts

### **Prompt 3: Phase 2 - Record & Playback**

```
Implementasikan Phase 2: Record & Playback untuk GoTest Agent.

Baca dokumentasi lengkap di docs/PHASE-2-IMPLEMENTATION.md

Scope implementasi:
1. Chrome extension (Manifest V3) untuk record user interactions
2. Backend API untuk receive recorded events
3. Service untuk convert recorded events jadi test code
4. Test library management (save, browse, reuse tests)

Implementation checklist:
- [ ] Chrome extension: manifest.json, content.js, background.js
- [ ] Backend: POST /api/v1/recordings endpoint
- [ ] Database: recordings table schema
- [ ] Service: convert recorded events → test code
- [ ] UI: test library browse page
- [ ] Integration: save/load/delete tests

Ikuti protocol dari MEMORY.md. Test setiap component sebelum lanjut ke next.
Berikan progress report setelah setiap milestone.
```

### **Prompt 4: Phase 3 - Continuous Sync**

```
Implementasikan Phase 3: Continuous Sync untuk GoTest Agent.

Baca implementasi guide di docs/PHASE-3-IMPLEMENTATION.md

Scope:
1. GitHub webhook integration untuk auto-detect code changes
2. Drift detection service untuk identify test yang outdated
3. Auto-generation service untuk regenerate affected tests
4. Alert system untuk notify tim tentang drift

Prioritas:
- HIGH: Webhook integration (foundation untuk semua fitur lain)
- MEDIUM: Drift detection (core value prop)
- MEDIUM: Auto-generation (value prop)
- LOW: Alert system (nice to have)

Implementation checklist:
- [ ] POST /api/v1/webhooks/github endpoint
- [ ] Drift detection service
- [ ] Auto-generation service
- [ ] Database: webhooks table schema
- [ ] Alert system (Slack/Email notifications)

Ikuti protocol dari MEMORY.md. Test setiap component sebelum lanjut ke next.
```

### **Prompt 5: Phase 4 - Advanced AI Features**

```
Implementasikan Phase 4: Advanced AI Features untuk GoTest Agent.

Baca implementasi guide di docs/PHASE-4-IMPLEMENTATION.md

Scope:
1. Multi-language parsers (Ruby Rails, Java Spring Boot, C# ASP.NET Core)
2. AI-driven test optimization (identify flaky tests, redundant tests)
3. Code review assistant (suggest improvements)
4. Test coverage insights dan analytics

Prioritas:
- HIGH: Ruby Rails parser (most requested)
- HIGH: Java Spring Boot parser (enterprise market)
- MEDIUM: C# ASP.NET Core parser
- MEDIUM: AI-driven optimization
- LOW: Advanced analytics

Implementation checklist:
- [ ] Ruby Rails parser (routes, models, views)
- [ ] Java Spring Boot parser (controllers, models, services)
- [ ] C# ASP.NET Core parser (controllers, models, services)
- [ ] AI-driven test optimization service
- [ ] Code review assistant
- [ ] Test coverage analytics

Ikuti protocol dari MEMORY.md. Implement one language at a time, test thoroughly.
```

---

## 🔧 Setup & Verification Prompts

### **Prompt 6: Setup dan Verification**

```
Setup dan verify GoTest Agent untuk production.

Langkah-langkah:
1. Install tree-sitter dependencies:
   ./scripts/install-tree-sitter-deps.sh
   
   Atau manual:
   go get github.com/smacker/go-tree-sitter@v0.0.0-20240827094217-dd81d9e9be82
   go get github.com/smacker/go-tree-sitter/javascript@v0.0.0-20240827094217-dd81d9e9be82
   go get github.com/smacker/go-tree-sitter/golang@v0.0.0-20240827094217-dd81d9e9be82
   go get github.com/smacker/go-tree-sitter/python@v0.0.0-20240827094217-dd81d9e9be82
   go get github.com/smacker/go-tree-sitter/php@v0.0.0-20240827094217-dd81d9e9be82
   go mod tidy

2. Run parser tests:
   go test ./internal/parser/... -v

3. Start server:
   go run ./cmd/server

4. Verify semua endpoint working:
   - Health check: GET /health
   - Create run: POST /api/v1/runs
   - Get run: GET /api/v1/runs/{id}
   - Get events: GET /api/v1/runs/{id}/events

5. Test end-to-end flow:
   - Generate test dari requirement
   - Execute test
   - Verify result

Jika ada error, troubleshoot dan fix. Document semua issue di .ai/CHANGELOG_AI.md.
```

### **Prompt 7: Quick Start (New Session)**

```
Lanjutkan GoTest Agent development.

Current status:
- Phase 1 (Multi-language parsers): COMPLETE
- Phase 2 (Record & Playback): Implementation guide ready
- Phase 3 (Continuous Sync): Implementation guide ready
- Phase 4 (Advanced AI): Implementation guide ready

Baca FINAL_SUMMARY.md untuk understand current state.
Baca implementation guides di docs/ untuk understand what to implement.

Next step: Implement Phase 2 (Record & Playback)

Ikuti protocol dari MEMORY.md.
Berikan progress report setelah setiap milestone.
```

---

## 🎨 Feature-Specific Prompts

### **Prompt 8: Specific Feature Request**

```
Implementasikan [NAMA_FEATURE] untuk GoTest Agent.

Context:
- [Jelaskan feature yang diinginkan]
- [Jelaskan kenapa feature ini penting]
- [Jelaskan expected behavior]
- [Jelaskan acceptance criteria]

Baca dokumentasi terkait di docs/ untuk understand current architecture.
Ikuti protocol dari MEMORY.md:
- Investigate first
- Preserve architecture
- Small atomic commits
- Verify thoroughly
- Document di .ai/CHANGELOG_AI.md

Test thoroughly sebelum commit.
```

### **Prompt 9: Implement Specific Parser**

```
Implementasikan [LANGUAGE] parser untuk GoTest Agent.

Baca docs/PHASE-4-IMPLEMENTATION.md untuk implementation guide.

Scope:
- Parse [LANGUAGE] routes/models/controllers
- Extract routes dengan methods (GET, POST, PUT, DELETE)
- Extract models dengan fields dan types
- Extract controllers/handlers dengan methods

Test dengan sample [LANGUAGE] code.
Ikuti protocol dari MEMORY.md.
```

### **Prompt 10: Implement API Endpoint**

```
Implementasikan [ENDPOINT_NAME] API endpoint untuk GoTest Agent.

Baca docs/API.md untuk understand current API structure.

Endpoint specification:
- Method: [GET/POST/PUT/DELETE]
- Path: /api/v1/[path]
- Request body: [describe]
- Response: [describe]
- Authentication: [required/optional]

Ikuti protocol dari MEMORY.md.
Test endpoint dengan curl atau Postman.
Document di .ai/CHANGELOG_AI.md.
```

---

## 🐛 Debugging & Troubleshooting Prompts

### **Prompt 11: Debug Issue**

```
Debug dan fix issue di GoTest Agent.

Issue description: [DESCRIBE ISSUE]

Langkah-langkah:
1. Reproduce issue
2. Check logs: [specify logs to check]
3. Identify root cause
4. Fix issue
5. Test fix
6. Document fix di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

### **Prompt 12: Test Failure**

```
Fix test failure di GoTest Agent.

Test yang gagal: [TEST_NAME]
Error message: [ERROR_MESSAGE]

Langkah-langkah:
1. Run test: go test ./[package] -v -run [TEST_NAME]
2. Analyze error message
3. Identify root cause
4. Fix issue
5. Re-run test untuk verify fix
6. Document fix di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

---

## 📊 Analytics & Optimization Prompts

### **Prompt 13: Performance Optimization**

```
Optimize performance GoTest Agent.

Focus areas:
- [ ] Database query optimization
- [ ] API response time
- [ ] Memory usage
- [ ] Test execution speed

Langkah-langkah:
1. Profile current performance
2. Identify bottlenecks
3. Implement optimizations
4. Measure improvements
5. Document optimizations di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

### **Prompt 14: Code Quality Improvement**

```
Improve code quality GoTest Agent.

Focus areas:
- [ ] Code duplication
- [ ] Error handling
- [ ] Code comments
- [ ] Test coverage

Langkah-langkah:
1. Run go vet ./...
2. Run golint ./...
3. Run go test -cover ./...
4. Identify issues
5. Fix issues
6. Document improvements di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

---

## 📚 Documentation Prompts

### **Prompt 15: Update Documentation**

```
Update dokumentasi GoTest Agent.

Scope:
- [ ] Update FINAL_SUMMARY.md dengan latest achievements
- [ ] Update API.md dengan new endpoints
- [ ] Update ARCHITECTURE.md dengan new components
- [ ] Update implementation guides dengan latest progress
- [ ] Add new sections jika perlu

Ikuti protocol dari MEMORY.md.
Document semua perubahan di .ai/CHANGELOG_AI.md.
```

### **Prompt 16: Create New Documentation**

```
Buat dokumentasi baru untuk [TOPIC].

Scope:
- [Jelaskan scope dokumentasi]
- [Jelaskan target audience]
- [Jelaskan expected content]

Format:
- Use Markdown
- Include code examples
- Include diagrams jika perlu
- Include troubleshooting section

Ikuti protocol dari MEMORY.md.
```

---

## 🚀 Advanced Prompts

### **Prompt 17: Autonomous Development with Priority Matrix**

```
Lanjutkan autonomous development GoTest Agent dengan priority matrix.

Baca semua dokumentasi di docs/:
- FINAL_SUMMARY.md
- PHASE-2-IMPLEMENTATION.md
- PHASE-3-IMPLEMENTATION.md
- PHASE-4-IMPLEMENTATION.md

Priority matrix:
1. HIGH impact + HIGH feasibility → implement immediately
2. HIGH impact + LOW feasibility → implement second
3. LOW impact + HIGH feasibility → implement third
4. LOW impact + LOW feasibility → implement last

Untuk setiap feature:
1. Read documentation
2. Identify current state
3. Design solution
4. Implement dengan small atomic commits
5. Test thoroughly (go test ./...)
6. Document di .ai/CHANGELOG_AI.md
7. Commit dan push

Berikan progress report setelah setiap feature:
- Feature name
- Status (done/in progress/blocked)
- What was implemented
- What's left to do
- Next steps

Jika stuck, tanya sebelum lanjut.
```

### **Prompt 18: Feature Comparison dengan TestSprite**

```
Compare GoTest Agent dengan TestSprite dan identify gaps.

Baca docs/FEATURES_COMPARISON.md untuk current comparison.

Langkah-langkah:
1. Review current features
2. Identify missing features vs TestSprite
3. Prioritize missing features berdasarkan impact
4. Implement top 3 missing features
5. Update comparison document
6. Document di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

### **Prompt 19: Production Readiness**

```
Verify dan improve production readiness GoTest Agent.

Checklist:
- [ ] All tests pass (go test ./...)
- [ ] Code quality check (go vet, golint)
- [ ] Security audit (check for vulnerabilities)
- [ ] Performance optimization
- [ ] Documentation complete
- [ ] Error handling robust
- [ ] Logging comprehensive
- [ ] Monitoring in place
- [ ] Deployment guide complete

Langkah-langkah:
1. Run all checks
2. Identify issues
3. Fix issues
4. Re-run checks
5. Document improvements di .ai/CHANGELOG_AI.md

Ikuti protocol dari MEMORY.md.
```

---

## 📋 Template Summary

### Quick Reference

| Prompt | Purpose | When to Use |
|--------|---------|-------------|
| Prompt 1 | General continuation | New session, continue development |
| Prompt 2 | Autonomous full development | Implement all phases |
| Prompt 3 | Phase 2 implementation | Record & Playback |
| Prompt 4 | Phase 3 implementation | Continuous Sync |
| Prompt 5 | Phase 4 implementation | Advanced AI features |
| Prompt 6 | Setup & verification | First time setup |
| Prompt 7 | Quick start | New session |
| Prompt 8 | Specific feature | Implement specific feature |
| Prompt 9 | Implement parser | New language parser |
| Prompt 10 | Implement API endpoint | New API endpoint |
| Prompt 11 | Debug issue | Fix bugs |
| Prompt 12 | Test failure | Fix test failures |
| Prompt 13 | Performance optimization | Optimize performance |
| Prompt 14 | Code quality | Improve code quality |
| Prompt 15 | Update documentation | Update docs |
| Prompt 16 | Create documentation | Create new docs |
| Prompt 17 | Autonomous with priority | Implement by priority |
| Prompt 18 | TestSprite comparison | Compare features |
| Prompt 19 | Production readiness | Production check |

---

## 🎯 Rekomendasi

### **Untuk New Session (RECOMMENDED):**
```
Lanjutkan GoTest Agent development. Baca FINAL_SUMMARY.md dan semua implementation guides di docs/, 
lalu lanjutkan dari phase yang belum selesai. Ikuti protocol dari MEMORY.md. 
Berikan progress report setelah setiap milestone.
```

### **Untuk Autonomous Development:**
```
Lanjutkan autonomous development GoTest Agent. Implement semua unfinished features dari Phase 2-4 
berdasarkan priority. Test thoroughly dan document semua perubahan.
```

### **Untuk Specific Feature:**
```
Implementasikan [feature_name] untuk GoTest Agent. Baca dokumentasi terkait, 
ikuti protocol MEMORY.md, test thoroughly.
```

### **Untuk Quick Fix:**
```
Fix issue: [DESCRIBE_ISSUE]. Investigate, fix, test, dan document di .ai/CHANGELOG_AI.md.
```

---

## 📝 Notes

### **Protocol dari MEMORY.md:**
- ✅ Investigate first (understand before implement)
- ✅ Preserve architecture (don't break existing features)
- ✅ Small atomic commits
- ✅ Verify thoroughly (run tests)
- ✅ Document di .ai/CHANGELOG_AI.md

### **Best Practices:**
1. **Always read documentation first** - understand current state
2. **Follow protocol from MEMORY.md** - maintain consistency
3. **Small atomic commits** - easier to review and rollback
4. **Test thoroughly** - run tests before commit
5. **Document everything** - document in .ai/CHANGELOG_AI.md
6. **Progress reports** - report after each milestone
7. **Ask when stuck** - don't waste time on blockers

### **Current Status:**
- ✅ Phase 1 (Multi-language parsers): COMPLETE
- 🔄 Phase 2 (Record & Playback): Implementation guide ready
- 🔄 Phase 3 (Continuous Sync): Implementation guide ready
- 🔄 Phase 4 (Advanced AI): Implementation guide ready

### **Next Steps:**
1. Implement Phase 2 (Record & Playback) - HIGH priority
2. Implement Phase 3 (Continuous Sync) - HIGH priority
3. Implement Phase 4 (Advanced AI) - MEDIUM priority

---

**Last Updated:** 2026-07-31  
**Status:** Production Ready  
**Total Prompts:** 19 templates
