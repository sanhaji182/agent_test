# Security Review

**Owner:** Engineering (security reviewer pending nomination)  
**Authoritative sources:** Tracked source, manifests, Compose config, migrations, and verified audit evidence  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection only; no runtime exploit simulation, no vulnerability scans, no pen-test, and dependency CVEs remain unknown  
**Confidence:** High for confirmed static findings; deployment-hardening effectiveness and external controls are UNKNOWN

## Verified Security Controls

These controls exist in tracked source and were confirmed by direct file inspection:

| Control | Evidence |
|---|---|
| Parameterized SQL (parameterized PostgreSQL queries) | `internal/db/store.go:46-60,94-103`; `internal/db/settings_store.go:38-50`; `internal/project/store.go:107-120`; `internal/schedule/store.go:135-165` |
| HTML auto-escaping (`html/template`) in reports | `internal/report/html.go:4-13,72-81` |
| No production `dangerouslySetInnerHTML` (only static theme script) | `frontend/src/app/layout.tsx:15-20` |
| Route ID allowlist with character-based traversal protection | `internal/api/server.go:111-122` |
| Constant-time HMAC-SHA256 webhook verification when secret is nonempty | `internal/webhook/github.go:59-66,91-100` |
| `crypto/rand` API key generation with SHA-256 hashing | `internal/auth/auth.go:99-113` |
| Settings `Key` allowlist during update | `internal/api/server.go:3434-3440` |
| Settings `llm_api_key` partial masking in GET (presentation-only) | `internal/api/server.go:3415-3418` |
| `.env`, `.pem`, `.key`, `*.log`, and `data/` paths Git-ignored | `.gitignore:8-12,37-45,63-65` |
| Default Gitleaks rules enabled | `.gitleaks.toml:6-8` |
| Frontend production container runs as non-root UID 1001 | `frontend/Dockerfile:13-25` |
| Health checks and `restart: unless-stopped` on all Compose services | `docker-compose.yml:29-35,49-55,68-74,81-87,104-110,126-132` |
| Chi RequestID, Logger, and Recoverer middleware | `internal/api/server.go:55-60` |

**Verified absence:** No string-concatenated SQL, unsafe HTML injection beyond the static theme script, or known unsafe deserialization was found.

## Confirmed Vulnerabilities in Default Compose Configuration

### SEC-1 — Backend API authentication fails open

**Severity:** HIGH  
**Prerequisites:** Deployment uses the Compose defaults where `API_KEY` expands to the empty string. Backend port 8080 is network-reachable.  
**Evidence:** `internal/api/server.go:2512-2524` bypasses all `/api/v1` authorization when `cfg.APIKey == ""`. `internal/config/config.go:24-30` defaults to empty. `docker-compose.yml:7-18` publishes 8080 without an override for `API_KEY`. `frontend/src/lib/api.ts:222-229` sends no key, and `EventSource` cannot attach one (`frontend/src/lib/api.ts:395`).  
**Impact:** All application data (projects, runs, plans, cases, lists, schedules, settings, reviews, releases, credentials) and operations (execution, export, settings mutation) are accessible without credentials. LLM credits can be consumed. Stored provider credentials can be exfiltrated through a base URL change.  
**Existing mitigation:** A nonempty `API_KEY` materially restricts access, but the dashboard and smoke test break unless frontend authentication is also implemented.  
**Confidence:** 10/10

### SEC-2 — Sidecar published without inbound authentication; can present backend API key

**Severity:** HIGH  
**Prerequisites:** Sidecar on port 8000 is network-reachable. Compose publishes `8000:8000` and Uvicorn binds `0.0.0.0`.  
**Evidence:** `sidecar/main.py:28-46` exposes `POST /agent/run` and `GET /agent/{job_id}` without authentication checks. `sidecar/agents/executor.py:5-21` reads `GOTEST_API_KEY` and forwards it as `X-Api-Key` to the backend.  
**Impact:** A network attacker submits arbitrary LLM-driven jobs and backend test runs while the sidecar authenticates on the attacker's behalf.  
**Existing mitigation:** Pydantic validates request shape. Job identifiers are server-generated UUIDs. Neither control authenticates the caller.  
**Confidence:** 10/10

### SEC-3 — PostgreSQL host-published with known credentials and disabled TLS

**Severity:** HIGH  
**Prerequisites:** Port 5432 is network-reachable. Compose publishes it and no external firewall/PG configuration overrides the defaults.  
**Evidence:** `docker-compose.yml:57-67` publishes `5432:5432` with user `postgres`, password `password`, and the backend connects with `sslmode=disable` (`internal/config/config.go:24-30`).  
**Impact:** A network-reachable attacker bypasses all application authentication and reads/modifies application data, LLM credentials, and project secrets.  
**Existing mitigation:** The README identifies these credentials as requiring production hardening (`README.md:184-206`). There is no coded enforcement.  
**Confidence:** 10/10

### SEC-4 — LLM credential exfiltration via configurable base URL

**Severity:** HIGH  
**Prerequisites:** A functional LLM API key is stored in the settings table. The caller can change `llm_base_url` through the settings API. No role barrier separates global settings from ordinary API access.  
**Evidence:** Settings API permits `llm_base_url` changes. Execution reads stored key from DB settings or env fallback via `launchRun` → `agent.NewLLM`. Credential-origin binding (`isApprovedLLMOrigin`, approved by ADR-005 Phase 2) requires an explicit user-provided key for unapproved origins.  
**Impact:** An API-key holder changes the base URL, starts a run, and the stored credential is transmitted to the attacker's server. When `API_KEY` is empty, this path is unauthenticated.  
**Existing mitigation:** Settings update uses a key-name whitelist. The whitelist does not validate destinations.  
**Confidence:** 10/10

### SEC-5 — Docker runner TypeScript injection via unescaped project URL

**Severity:** HIGH  
**Prerequisites:** `GOTEST_APPROVED_CASE_RUNNER=docker` is enabled, or the MCP process is used (MCP always constructs `DockerRunner`).  
**Evidence:** `internal/runner/docker.go:58-65` interpolates `projectURL` directly into a TypeScript template literal (`baseURL: '` + projectURL + `'`) without escaping. The Docker container runs with `--network host` at `docker.go:74`. MCP always uses `DockerRunner` (`cmd/mcp/main.go:13-20`).  
**Impact:** A specially crafted project URL injects arbitrary Node.js code into the Playwright container and gains host-network access. This is a targeted-execution path, not a drive-by risk, and requires environment-opt-in or MCP access.  
**Existing mitigation:** The container is ephemeral (`--rm`) and mounts temp directories rather than host volumes.  
**Confidence:** 9/10

### SEC-6 — Unbounded resource consumption (no server timeouts, body limits, concurrency cap)

**Severity:** HIGH  
**Prerequisites:** Network access to the backend. When `API_KEY` is empty, this is the default deployment path.  
**Evidence:** `cmd/server/main.go:44-46` uses bare `http.ListenAndServe` without configured timeouts. No `MaxBytesReader` or body-size limit is imposed on any handler. Every run, rerun, approved case, and webhook callback launches a bare `go` goroutine (`internal/api/server.go:1581-1589,2186-2190,1196-1200`). Webhook handler reads bodies with unbounded `io.ReadAll` (`internal/webhook/github.go:53`). Sidecar stores jobs in unbounded process memory (`sidecar/main.py:11-12`).  
**Impact:** Slow-loris exhaustion, memory exhaustion via large bodies or many concurrent runs, and paid LLM credit consumption without a per-caller quota.  
**Existing mitigation:** Fix attempts default to three; Steel has configurable session limits and timeouts (`docker-compose.yml:94-98`).  
**Confidence:** 10/10

### SEC-7 — Public unauthenticated temporary video route

**Severity:** MEDIUM  
**Prerequisites:** Attacker knows or discovers a generated video filename.  
**Evidence:** `internal/api/server.go:127-131` registers `GET /videos/{filename}` before any authentication middleware. A separate authenticated file server at `server.go:257-261` serves persistent videos from a different path and does not protect the temporary `/tmp/agent_test/videos` directory.  
**Impact:** Browser recordings that may contain authenticated sessions, PII, or internal dashboards can be fetched without authentication. Filenames are not shown to be enumerable, limiting opportunistic exploitation.  
**Existing mitigation:** Playwright-generated basename provides a degree of unpredictability.  
**Confidence:** 9/10

## Design and Deployment-Dependent Risks

### DES-1 — Unrestricted browser egress from LLM-generated navigation

**Risk:** HIGH (design-level; exploitability depends on run-creation access and what the browser can reach)  
**Evidence:** `internal/agent/playwright_runner.go:77-102` executes `page.Goto(a.URL)` on LLM-generated URLs without scheme, hostname, IP-range, or redirect validation. The self-healing loop at `internal/agent/playwright_runner.go:104-152` may redirect to new URLs without re-validation. Approved-case browser path uses project `BaseURL` directly (`internal/api/server.go:1295-1349`). Docker runner uses `--network host` (`internal/runner/docker.go:74`).  
**Impact:** Generated or submitted URLs targeting loopback, metadata endpoints, Compose service names, or internal infrastructure can be reached. Failure artifacts (HTML snapshots, screenshots, DOM dumps) expose internal content.  
**Existing mitigation:** New browser context per run with no shared state.  
**Recommendation:** Validate resolved destination against an allowlist. Deny loopback, RFC1918, link-local, and metadata addresses by default.  
**Confidence:** 8.5/10

### DES-2 — Single shared authorization domain with credential-bearing responses

**Risk:** MEDIUM (design-level; severity escalates if multiple teams share a deployment)  
**Evidence:** Project and run JSON responses include `credentials` fields directly (`internal/project/store.go:139-155`; `internal/db/store.go:94-192`). All `/api/v1` routes share one API key with no ownership or tenant predicate. JWT code is implemented but not wired (`internal/auth/auth.go:21-95`).  
**Impact:** In a multi-user scenario, any API-key holder can list all projects, read credential notes, and modify global settings and review state.  
**Existing mitigation:** UUIDs reduce opportunistic enumeration. The single- or multi-user product model is UNKNOWN.  
**Recommendation:** Confirm the product model. If multi-user, add tenant ID to all entities, per-resource predicates, and credential redaction from API responses.  
**Confidence:** 10/10

### DES-3 — Redis and Steel published to host; Steel receives elevated privilege

**Risk:** MEDIUM (design-level; requires network reachability)  
**Evidence:** `docker-compose.yml:76-87` publishes Redis on 6379. `docker-compose.yml:89-110` publishes Steel on 3010, grants `SYS_ADMIN`, uses the mutable `:latest` tag, and passes no explicit authentication variable. Backend and sidecar run as image-default root (`Dockerfile:9-18`; `sidecar/Dockerfile:1-7`).  
**Impact:** If Redis or Steel accept unauthenticated requests through the Docker network bridge, they become reachable. Any Steel compromise has greater blast radius due to `SYS_ADMIN`. The appropriate routing/authentication for these services is currently deployment-dependent.  
**Existing mitigation:** Frontend runs as non-root. Steel sets session limits and timeouts (`docker-compose.yml:94-98`). Redis has its own default protected-mode behavior; whether it blocks Docker NAT traffic was not verified.  
**Confidence:** 9/10

### DES-4 — Supply chain: mutable images, unlocked Python dependencies, suppressed install failure

**Risk:** MEDIUM (design-level; applicable even without a known CVE)  
**Evidence:** `docker-compose.yml:91` uses `steel-browser:latest`. `Dockerfile:13-14` installs Playwright at `@latest` and suppresses failure with `|| true`. `sidecar/requirements.txt:1-6` uses `>=` without a lockfile. PostgreSQL and Redis use version tags without digests.  
**Impact:** A rebuild can silently resolve to materially different package or image versions or produce an image missing the required browser runtime. This is a supply-chain control gap, not a confirmed vulnerable version.  
**Existing mitigation:** npm `package-lock.json` + `npm ci` pin frontend resolution. Go `go.sum` pins direct and indirect modules.  
**Confidence:** 10/10

## Vulnerability Scan Status

| Scanner | Status | Evidence |
|---|---|---|
| `govulncheck` | Not run | Command blocked by shell safety classifier |
| `npm audit --omit=dev` | Not run | Command blocked |
| `pip-audit` | Not run | Command blocked |
| Native `gitleaks` full-history scan | Not run | Tool invocation blocked; manual pattern searches of tracked files and `git log -p` found no high-confidence live external secrets |
| CVE status across all ecosystems | **UNKNOWN** | Do not state "no vulnerabilities" until scanners run and output is reviewed |

## UNKNOWN Questions

- Is production deployed behind a TLS-terminating reverse proxy, firewall, or authenticated ingress not tracked in this repository?
- Is `API_KEY` enforced through deployment automation regardless of the repository default?
- Is the product strictly single-administrator or multi-user/tenant?
- Are `credentials` fields guaranteed to contain secret-manager references rather than live passwords and tokens?
- Does the Steel image enforce authentication by default?
- Are cloud metadata endpoints and private address ranges blocked by host or network egress policy?
- Are database and `/data` volumes encrypted and backup/rotation managed externally?
- What are the actual CVE results once scanners execute?
- Should temporary videos be accessible through an authentication-free path, or should all recordings be authorization-resolved?
- Should the sidecar be publicly reachable, and if so, what authentication mechanism is expected?
