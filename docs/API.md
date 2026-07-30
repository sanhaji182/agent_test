# GoTest Agent API Documentation

Base URL: `http://localhost:8080`

All API endpoints are prefixed with `/api/v1/` unless otherwise noted.

## Authentication

### API Key Authentication

Set the `API_KEY` environment variable to enable API key authentication. Pass the key in the `Authorization` header:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8080/api/v1/runs
```

If `API_KEY` is empty, authentication is disabled (development mode).

## Endpoints

### Test Runs

#### Create a Test Run

Create and execute a new test run.

```http
POST /api/v1/runs
```

**Request Body:**
```json
{
  "project_path": "/path/to/project",
  "requirements": "test login and checkout flows",
  "mode": "full",
  "test_type": "e2e",
  "browser": "chromium",
  "viewport": "desktop",
  "parallel": true,
  "test_data": {
    "username": "test@example.com",
    "password": "testpass123"
  },
  "tags": ["smoke", "regression"],
  "webhook_url": "https://hooks.slack.com/..."
}
```

**Response:** `202 Accepted`
```json
{
  "id": "run_abc123",
  "status": "pending",
  "created_at": "2026-01-15T10:30:00Z"
}
```

**Parameters:**
- `project_path` (required): Path to the project to test
- `requirements` (optional): Natural language requirements for test generation
- `mode` (optional): `simple` or `full` (default: `full`)
- `test_type` (optional): `unit`, `integration`, or `e2e` (default: `e2e`)
- `browser` (optional): `chromium`, `firefox`, or `webkit` (default: `chromium`)
- `viewport` (optional): `desktop`, `mobile`, or custom (default: `desktop`)
- `parallel` (optional): Run tests in parallel (default: `false`)
- `test_data` (optional): Test data for parameterized tests
- `tags` (optional): Tags for categorizing tests
- `webhook_url` (optional): Webhook URL for notifications

#### List Test Runs

```http
GET /api/v1/runs
```

**Query Parameters:**
- `limit` (optional): Maximum number of runs to return (default: 50)
- `status` (optional): Filter by status (`pending`, `running`, `passed`, `failed`)
- `tag` (optional): Filter by tag

**Response:** `200 OK`
```json
{
  "runs": [
    {
      "id": "run_abc123",
      "status": "passed",
      "created_at": "2026-01-15T10:30:00Z",
      "duration": 45.2,
      "total_tests": 10,
      "passed": 10,
      "failed": 0,
      "tags": ["smoke"]
    }
  ]
}
```

#### Get Test Run Details

```http
GET /api/v1/runs/:id
```

**Response:** `200 OK`
```json
{
  "id": "run_abc123",
  "status": "passed",
  "project_path": "/path/to/project",
  "requirements": "test login flows",
  "mode": "full",
  "test_type": "e2e",
  "browser": "chromium",
  "viewport": "desktop",
  "parallel": false,
  "created_at": "2026-01-15T10:30:00Z",
  "completed_at": "2026-01-15T10:31:30Z",
  "duration": 90.5,
  "total_tests": 15,
  "passed": 14,
  "failed": 1,
  "healed": 2,
  "tags": ["smoke"],
  "test_plan": {
    "summary": "Test critical user flows",
    "scenarios": [
      {
        "name": "Login flow",
        "priority": "high",
        "steps": ["Navigate to login", "Enter credentials", "Click login button"]
      }
    ]
  },
  "test_files": [
    {
      "name": "login.spec.ts",
      "content": "import { test } from '@playwright/test';\n..."
    }
  ],
  "run_result": {
    "passed": 14,
    "failed": 1,
    "total": 15,
    "duration_ms": 90500,
    "healed": 2,
    "failures": [
      {
        "test": "should complete checkout",
        "message": "Timeout waiting for element",
        "screenshot": "http://localhost:8080/api/v1/runs/run_abc123/screenshots/checkout-failure.png"
      }
    ]
  },
  "video_url": "http://localhost:8080/api/v1/runs/run_abc123/video",
  "report_url": "http://localhost:8080/api/v1/runs/run_abc123/report"
}
```

#### Get Test Run Report

```http
GET /api/v1/runs/:id/report
```

**Response:** `200 OK` (HTML)

Returns an HTML report with test results, screenshots, and execution details.

#### Get Test Run Video

```http
GET /api/v1/runs/:id/video
```

**Response:** `200 OK` (video/mp4)

Returns the recorded video of the test execution.

#### Get Test Run Screenshots

```http
GET /api/v1/runs/:id/screenshots
```

**Response:** `200 OK`
```json
{
  "screenshots": [
    {
      "name": "login-success.png",
      "url": "http://localhost:8080/api/v1/runs/run_abc123/screenshots/login-success.png",
      "timestamp": "2026-01-15T10:30:45Z"
    }
  ]
}
```

#### Cancel a Test Run

```http
POST /api/v1/runs/:id/cancel
```

**Response:** `200 OK`
```json
{
  "status": "cancelled"
}
```

#### Delete a Test Run

```http
DELETE /api/v1/runs/:id
```

**Response:** `204 No Content`

### Webhooks

#### Receive GitHub Webhook

```http
POST /api/v1/webhooks/github
```

**Headers:**
- `X-Hub-Signature-256`: HMAC signature (required if `GITHUB_WEBHOOK_SECRET` is set)
- `X-GitHub-Event`: Event type (e.g., `push`, `pull_request`)

**Request Body:** GitHub webhook payload

**Response:** `200 OK`
```json
{
  "status": "accepted",
  "message": "Webhook processed"
}
```

#### Register Webhook

```http
POST /api/v1/webhooks/register
```

**Request Body:**
```json
{
  "repository_url": "https://github.com/owner/repo",
  "github_token": "ghp_...",
  "secret": "your-webhook-secret"
}
```

**Response:** `200 OK`
```json
{
  "status": "registered",
  "webhook_id": "wh_abc123"
}
```

### Test Plans

#### Get Test Plan

```http
GET /api/v1/runs/:run_id/plan
```

**Response:** `200 OK`
```json
{
  "summary": "Test critical user flows",
  "scenarios": [
    {
      "name": "Login flow",
      "priority": "high",
      "steps": ["Navigate to login", "Enter credentials", "Click login button"],
      "expected_result": "User is logged in and redirected to dashboard"
    }
  ]
}
```

#### Update Test Plan

```http
PUT /api/v1/runs/:run_id/plan
```

**Request Body:**
```json
{
  "summary": "Updated test plan summary",
  "scenarios": [
    {
      "name": "Updated scenario",
      "priority": "high",
      "steps": ["Updated steps"]
    }
  ]
}
```

**Response:** `200 OK`

### Test Generation

#### Generate Tests from Code

```http
POST /api/v1/generate
```

**Request Body:**
```json
{
  "project_path": "/path/to/project",
  "requirements": "test all API endpoints",
  "test_type": "e2e"
}
```

**Response:** `202 Accepted`
```json
{
  "status": "generating",
  "message": "Test generation started"
}
```

#### Get Generation Status

```http
GET /api/v1/generate/:id
```

**Response:** `200 OK`
```json
{
  "status": "completed",
  "test_plan": {
    "summary": "Generated test plan",
    "scenarios": [...]
  },
  "test_files": [
    {
      "name": "api.spec.ts",
      "content": "..."
    }
  ]
}
```

### Schedules

#### Create Schedule

```http
POST /api/v1/schedules
```

**Request Body:**
```json
{
  "name": "nightly-tests",
  "project_path": "/path/to/project",
  "requirements": "run full regression suite",
  "frequency": "daily",
  "time": "02:00",
  "timezone": "UTC",
  "enabled": true,
  "webhook_url": "https://hooks.slack.com/..."
}
```

**Response:** `201 Created`
```json
{
  "id": "sched_abc123",
  "name": "nightly-tests",
  "frequency": "daily",
  "next_run": "2026-01-16T02:00:00Z",
  "enabled": true
}
```

#### List Schedules

```http
GET /api/v1/schedules
```

**Response:** `200 OK`
```json
{
  "schedules": [
    {
      "id": "sched_abc123",
      "name": "nightly-tests",
      "project_path": "/path/to/project",
      "frequency": "daily",
      "time": "02:00",
      "timezone": "UTC",
      "enabled": true,
      "next_run": "2026-01-16T02:00:00Z",
      "last_run": "2026-01-15T02:00:00Z"
    }
  ]
}
```

#### Update Schedule

```http
PUT /api/v1/schedules/:id
```

**Request Body:**
```json
{
  "enabled": false,
  "frequency": "weekly",
  "day_of_week": "monday"
}
```

**Response:** `200 OK`

#### Delete Schedule

```http
DELETE /api/v1/schedules/:id
```

**Response:** `204 No Content`

### Codebase Analysis

#### Analyze Codebase

```http
POST /api/v1/analyze
```

**Request Body:**
```json
{
  "project_path": "/path/to/project"
}
```

**Response:** `200 OK`
```json
{
  "language": "javascript",
  "framework": "express",
  "routes": [
    {
      "method": "GET",
      "path": "/api/users",
      "handler": "getUsers",
      "file": "src/routes/users.js",
      "line": 15
    }
  ],
  "models": [
    {
      "name": "User",
      "table": "users",
      "fields": [
        {"name": "id", "type": "integer"},
        {"name": "email", "type": "string"}
      ]
    }
  ],
  "handlers": [
    {
      "name": "getUsers",
      "controller": "UserController",
      "file": "src/controllers/user.controller.js",
      "line": 10
    }
  ]
}
```

### Metrics & Insights

#### Get Metrics Summary

```http
GET /api/v1/metrics/summary
```

**Response:** `200 OK`
```json
{
  "total_runs": 150,
  "total_tests": 1250,
  "total_passed": 1180,
  "total_failed": 70,
  "total_healed": 45,
  "pass_rate": 94.4,
  "avg_duration_ms": 45000,
  "success_rate": 96.7
}
```

#### Get Hotspots

```http
GET /api/v1/metrics/hotspots
```

**Response:** `200 OK`
```json
{
  "hotspots": [
    {
      "file": "src/components/LoginForm.tsx",
      "failure_count": 12,
      "last_failure": "2026-01-15T10:30:00Z"
    }
  ]
}
```

#### Get Flaky Tests

```http
GET /api/v1/metrics/flaky
```

**Response:** `200 OK`
```json
{
  "flaky_tests": [
    {
      "name": "should complete checkout",
      "flakiness_score": 0.35,
      "recent_results": ["passed", "failed", "passed", "passed", "failed"]
    }
  ]
}
```

#### Get Risk Assessment

```http
GET /api/v1/metrics/risk
```

**Response:** `200 OK`
```json
{
  "risks": [
    {
      "name": "Login flow",
      "type": "test",
      "risk_score": 0.85,
      "reason": "High failure rate (35%)"
    }
  ]
}
```

#### Get Recommendations

```http
GET /api/v1/metrics/recommendations
```

**Response:** `200 OK`
```json
{
  "recommendations": [
    {
      "action": "Fix flaky test",
      "target": "Login flow test",
      "reason": "High failure rate affecting reliability",
      "priority": "high"
    }
  ]
}
```

#### Get Test Trends

```http
GET /api/v1/metrics/trends
```

**Query Parameters:**
- `days` (optional): Number of days to include (default: 30)

**Response:** `200 OK`
```json
{
  "trends": [
    {
      "date": "2026-01-15",
      "total_runs": 5,
      "total_tests": 50,
      "passed": 48,
      "failed": 2,
      "pass_rate": 96.0
    }
  ]
}
```

#### Suite Selection

```http
POST /api/v1/suite-selection
```

**Request Body:**
```json
{
  "mode": "high_risk",
  "all_tests": ["login", "checkout", "profile", "settings"]
}
```

**Response:** `200 OK`
```json
{
  "selected_tests": ["login", "checkout"],
  "reason": "Selected high-risk tests"
}
```

### Reviews

#### Create Review

```http
POST /api/v1/reviews
```

**Request Body:**
```json
{
  "run_id": "run_abc123",
  "type": "test_plan",
  "reviewer": "alice"
}
```

**Response:** `201 Created`
```json
{
  "id": "rev_abc123",
  "status": "pending",
  "created_at": "2026-01-15T10:30:00Z"
}
```

#### Approve Review

```http
POST /api/v1/reviews/:id/approve
```

**Request Body:**
```json
{
  "reviewer": "alice",
  "comment": "Looks good!"
}
```

**Response:** `200 OK`
```json
{
  "status": "approved",
  "approved_at": "2026-01-15T11:00:00Z"
}
```

#### Reject Review

```http
POST /api/v1/reviews/:id/reject
```

**Request Body:**
```json
{
  "reviewer": "alice",
  "comment": "Needs more test coverage"
}
```

**Response:** `200 OK`
```json
{
  "status": "rejected",
  "rejected_at": "2026-01-15T11:00:00Z"
}
```

### Monitoring

#### Prometheus Metrics

```http
GET /metrics
```

**Response:** `200 OK` (text/plain)

Returns Prometheus-formatted metrics:
```
# HELP gotest_runs_total Total number of test runs
# TYPE gotest_runs_total counter
gotest_runs_total 150

# HELP gotest_tests_total Total number of tests executed
# TYPE gotest_tests_total counter
gotest_tests_total 1250

# HELP gotest_tests_passed Total number of tests passed
# TYPE gotest_tests_passed counter
gotest_tests_passed 1180

# HELP gotest_tests_failed Total number of tests failed
# TYPE gotest_tests_failed counter
gotest_tests_failed 70

# HELP gotest_tests_healed Total number of tests healed
# TYPE gotest_tests_healed counter
gotest_tests_healed 45
```

### Health Check

```http
GET /health
```

**Response:** `200 OK`
```json
{
  "status": "ok"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Invalid or missing API key
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Rate Limiting

The API currently does not implement rate limiting. For production use, consider implementing rate limiting middleware.

## CORS

By default, CORS is disabled. Set the `CORS_ALLOWED_ORIGINS` environment variable to enable CORS for specific origins:

```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-domain.com
```
