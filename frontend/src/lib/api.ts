const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface TestRun {
  id: string;
  project_path: string;
  requirements: string;
  mode: string;
  test_type?: string;
  test_case_id?: string;
  test_list_id?: string;
  prd?: string;
  api_docs?: string;
  auth_type?: string;
  credentials?: string;
  focus_hints?: string;
  skip_hints?: string;
  feature_map?: FeatureMap;
  state: string;
  code_analysis?: string;
  test_plan?: TestPlan;
  test_files?: TestFile[];
  run_result?: RunResult;
  screenshots?: string[];
  fix_attempts: number;
  error?: string;
  video_url?: string;
  video_status?: string;
  video_duration?: number;
  video_failure_marker_at?: number;
  created_at: string;
  updated_at: string;
  finished_at?: string;
}

export interface TestPlan {
  summary: string;
  scenarios: Scenario[];
}

export interface Scenario {
  name: string;
  priority: string;
  steps: string[];
}

export interface FeatureMap {
  source: string;
  features: { name: string; use_cases: string[] }[];
}

export interface Project {
  id: string;
  name: string;
  test_type: string;
  base_url: string;
  environment?: string;
  spec?: string;
  api_docs?: string;
  auth_type?: string;
  credentials?: string;
  focus_hints?: string;
  skip_hints?: string;
  feature_map?: FeatureMap;
  created_at: string;
  updated_at: string;
}

export interface DraftPlan {
  id: string;
  project_id: string;
  status: string;
  cases: DraftCase[];
  created_at: string;
  updated_at: string;
}

export interface DraftCase {
  id: string;
  title: string;
  type: string;
  feature: string;
  priority: string;
  enabled: boolean;
  steps: string[];
  assertions: string[];
  tags: string[];
  confidence: number;
}

export interface TestCase {
  id: string;
  project_id: string;
  plan_id?: string;
  title: string;
  type: string;
  feature: string;
  priority: string;
  steps: string[];
  assertions: string[];
  tags: string[];
  version: number;
  created_at: string;
  updated_at: string;
}

export interface TestList {
  id: string;
  name: string;
  project_id?: string;
  tags: string[];
  test_case_ids: string[];
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

export interface ChangeProposal {
  id: string;
  test_case_id: string;
  status: string;
  prompt: string;
  rationale: string;
  original: TestCase;
  proposed: TestCase;
  created_at: string;
  updated_at: string;
  reviewed_at?: string;
  reviewer?: string;
  review_comment?: string;
}

export interface MaintenanceItem {
  test_case_id: string;
  title: string;
  type: string;
  category: string;
  severity: string;
  reason: string;
  action: string;
  last_run_at?: string;
}

export interface MonitoringSummary {
  summary: {
    total_lists: number;
    total_cases: number;
    active_runs: number;
    failed_runs: number;
    completed_runs: number;
  };
  lists: {
    id: string;
    name: string;
    pinned: boolean;
    test_count: number;
    pass_rate: number;
    last_status: string;
    last_run_id?: string;
    last_run_at?: string;
    failed: number;
    passed: number;
    newly_failed: string[];
    recovered: string[];
    stable_failed: string[];
  }[];
  recent_runs: TestRun[];
}

export interface TestListHistory {
  test_list_id: string;
  name: string;
  latest?: {
    run_id?: string;
    status?: string;
    created_at?: string;
  };
  counts: {
    passed: number;
    failed: number;
  };
  runs: {
    id: string;
    test_case_id?: string;
    status: string;
    created_at: string;
    failed: number;
    passed: number;
  }[];
  newly_failed: string[];
  recovered: string[];
  stable_failed: string[];
}

export interface TestFile {
  name: string;
  content: string;
}

export interface RunResult {
  passed: number;
  failed: number;
  total: number;
  failures: Failure[];
  duration_ms?: number;
  healed_count?: number;
  retried_count?: number;
}

export interface AuditResult {
  id: string;
  performance: PerformanceMetrics | null;
  accessibility: AccessibilityResult | null;
  visual_regression: VisualRegressionResult | null;
  created_at: string;
}

export interface PerformanceMetrics {
  lcp_ms: number;
  fid_ms: number;
  cls: number;
  fcp_ms: number;
  ttfb_ms: number;
}

export interface AccessibilityResult {
  violations_count: number;
  passes_count: number;
  violations: AccessibilityViolation[];
}

export interface AccessibilityViolation {
  id: string;
  impact: string;
  description: string;
  help_url: string;
  nodes: number;
}

export interface VisualRegressionResult {
  diff_percentage: number;
  diff_count: number;
  total_pixels: number;
  diff_pixels: number;
  baseline_exists: boolean;
}

export interface ExportCodeResult {
  run_id?: string;
  target?: string;
  language: string;
  code: string;
  framework: string;
  count?: number;
  scripts?: Record<string, string>;
}

export interface Failure {
  test: string;
  message: string;
  screenshot_url?: string;
}

export interface FailureAnalysis {
  run_id: string;
  status: string;
  summary: string;
  likely_cause: string;
  next_action: string;
  evidence: string[];
  source: string;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: "include",  // Send httpOnly cookie (JWT) for dashboard auth
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });
  if (!res.ok) {
    // If unauthorized, redirect to login unless already there
    if (res.status === 401 && typeof window !== "undefined" && !window.location.pathname.startsWith("/login")) {
      sessionStorage.setItem("redirect_after_login", window.location.pathname + window.location.search);
      window.location.href = "/login";
      throw new Error("unauthorized — redirecting to login");
    }
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

// login exchanges an API key for a JWT httpOnly cookie session.
// The cookie is automatically sent on subsequent requests via credentials: "include".
export async function login(credentials: { email?: string; password?: string; apiKey?: string }): Promise<{ status: string; redirect?: string }> {
  const body: Record<string, string> = {};
  if (credentials.email && credentials.password) {
    body.email = credentials.email;
    body.password = credentials.password;
  } else if (credentials.apiKey) {
    body.api_key = credentials.apiKey;
  }
  const res = await fetch(`${API_BASE}/api/v1/auth/login`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    return res.json().catch(() => ({ status: "error" }));
  }
  const data = await res.json();
  if (typeof window !== "undefined") {
    sessionStorage.setItem("user_role", data.role || "admin");
    sessionStorage.setItem("user_label", data.role ? (data.role === "admin" ? "Admin" : data.role) : "Admin");
    const redirect = sessionStorage.getItem("redirect_after_login");
    if (redirect) {
      sessionStorage.removeItem("redirect_after_login");
      return { status: "ok", redirect };
    }
  }
  return { status: "ok", redirect: "/" };
}

// ─── User management (multi-user, admin-only) ─────────────────────────

export interface UserEntry {
  id: string;
  email: string;
  name: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export async function listUsers(): Promise<UserEntry[]> {
  return apiFetch<UserEntry[]>("/api/v1/users");
}

export async function createUser(data: { email: string; password: string; name: string; role: string }): Promise<UserEntry> {
  return apiFetch<UserEntry>("/api/v1/users", { method: "POST", body: JSON.stringify(data) });
}

export async function updateUser(id: string, data: { name?: string; role?: string; is_active?: boolean; new_password?: string }): Promise<UserEntry> {
  return apiFetch<UserEntry>(`/api/v1/users/${id}`, { method: "PUT", body: JSON.stringify(data) });
}

export async function deleteUser(id: string): Promise<void> {
  await apiFetch<unknown>(`/api/v1/users/${id}`, { method: "DELETE" });
}

// logout clears the httpOnly cookie session
export async function logout(): Promise<void> {
  await fetch(`${API_BASE}/api/v1/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

export async function getRuns(): Promise<TestRun[]> {
  return apiFetch<TestRun[]>("/api/v1/runs");
}

export async function getRun(id: string): Promise<TestRun> {
  return apiFetch<TestRun>(`/api/v1/runs/${id}`);
}

export async function createRun(data: {
  project_path: string;
  requirements?: string;
  mode?: string;
  test_type?: string;
  prd?: string;
  api_docs?: string;
  auth_type?: string;
  credentials?: string;
  focus_hints?: string;
  skip_hints?: string;
}): Promise<{ run_id: string; state: string }> {
  return apiFetch("/api/v1/runs", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// --- Projects ---
export async function getProjects(): Promise<Project[]> {
  return apiFetch<Project[]>("/api/v1/projects");
}

export async function createProject(data: Partial<Project>): Promise<Project> {
  return apiFetch<Project>("/api/v1/projects", { method: "POST", body: JSON.stringify(data) });
}

export async function updateProject(id: string, data: Partial<Project>): Promise<Project> {
  return apiFetch<Project>(`/api/v1/projects/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function extractProjectFeatures(id: string): Promise<FeatureMap> {
  return apiFetch<FeatureMap>(`/api/v1/projects/${id}/extract-features`, { method: "POST" });
}

export async function uploadApiDocs(id: string, apiDocs: string): Promise<Project> {
  return apiFetch<Project>(`/api/v1/projects/${id}/api-docs`, { method: "POST", body: JSON.stringify({ api_docs: apiDocs }) });
}

export async function parseApiDocs(id: string): Promise<DraftPlan> {
  return apiFetch<DraftPlan>(`/api/v1/projects/${id}/parse-api`, { method: "POST" });
}

export async function generateProjectTestPlan(id: string): Promise<DraftPlan> {
  return apiFetch<DraftPlan>(`/api/v1/projects/${id}/test-plan`, { method: "POST" });
}

export async function updateTestPlanCase(planId: string, caseId: string, data: Partial<DraftCase>): Promise<DraftPlan> {
  return apiFetch<DraftPlan>(`/api/v1/test-plans/${planId}/cases/${caseId}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function regenerateTestPlan(id: string): Promise<DraftPlan> {
  return apiFetch<DraftPlan>(`/api/v1/test-plans/${id}/regenerate`, { method: "POST" });
}

export async function approveTestPlan(id: string): Promise<{ status: string; test_cases: TestCase[] }> {
  return apiFetch<{ status: string; test_cases: TestCase[] }>(`/api/v1/test-plans/${id}/approve`, { method: "POST" });
}

export async function getTestCases(projectId?: string): Promise<TestCase[]> {
  const qs = projectId ? `?project_id=${encodeURIComponent(projectId)}` : "";
  return apiFetch<TestCase[]>(`/api/v1/test-cases${qs}`);
}

export async function updateTestCase(id: string, data: Partial<TestCase>): Promise<TestCase> {
  return apiFetch<TestCase>(`/api/v1/test-cases/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function getTestCaseMaintenance(projectId?: string): Promise<MaintenanceItem[]> {
  const qs = projectId ? `?project_id=${encodeURIComponent(projectId)}` : "";
  return apiFetch<MaintenanceItem[]>(`/api/v1/test-cases/maintenance${qs}`);
}

export async function runTestCase(id: string): Promise<{ run_id: string; state: string; test_case_id: string }> {
  return apiFetch<{ run_id: string; state: string; test_case_id: string }>(`/api/v1/test-cases/${id}/run`, { method: "POST" });
}

export async function refineTestCase(id: string, prompt: string): Promise<ChangeProposal> {
  return apiFetch<ChangeProposal>(`/api/v1/test-cases/${id}/refine`, { method: "POST", body: JSON.stringify({ prompt }) });
}

export async function getTestCaseProposals(id: string): Promise<ChangeProposal[]> {
  return apiFetch<ChangeProposal[]>(`/api/v1/test-cases/${id}/proposals`);
}

export async function getChangeProposals(testCaseId?: string): Promise<ChangeProposal[]> {
  const qs = testCaseId ? `?test_case_id=${encodeURIComponent(testCaseId)}` : "";
  return apiFetch<ChangeProposal[]>(`/api/v1/change-proposals${qs}`);
}

export async function approveChangeProposal(id: string, data: { reviewer?: string; comment?: string } = {}): Promise<{ proposal: ChangeProposal; test_case: TestCase }> {
  return apiFetch<{ proposal: ChangeProposal; test_case: TestCase }>(`/api/v1/change-proposals/${id}/approve`, { method: "POST", body: JSON.stringify(data) });
}

export async function rejectChangeProposal(id: string, data: { reviewer?: string; comment?: string } = {}): Promise<ChangeProposal> {
  return apiFetch<ChangeProposal>(`/api/v1/change-proposals/${id}/reject`, { method: "POST", body: JSON.stringify(data) });
}

export async function getTestLists(projectId?: string): Promise<TestList[]> {
  const qs = projectId ? `?project_id=${encodeURIComponent(projectId)}` : "";
  return apiFetch<TestList[]>(`/api/v1/test-lists${qs}`);
}

export async function createTestList(data: Partial<TestList>): Promise<TestList> {
  return apiFetch<TestList>("/api/v1/test-lists", { method: "POST", body: JSON.stringify(data) });
}

export async function runTestList(id: string): Promise<{ test_list_id: string; run_ids: string[] }> {
  return apiFetch<{ test_list_id: string; run_ids: string[] }>(`/api/v1/test-lists/${id}/run`, { method: "POST" });
}

export async function getTestListHistory(id: string): Promise<TestListHistory> {
  return apiFetch<TestListHistory>(`/api/v1/test-lists/${id}/history`);
}

export async function getMonitoringSummary(): Promise<MonitoringSummary> {
  return apiFetch<MonitoringSummary>("/api/v1/monitoring/summary");
}

// Rerun: buat run baru dari konfigurasi run lama
export async function rerunRun(id: string): Promise<{ run_id: string; state: string }> {
  return apiFetch(`/api/v1/runs/${id}/rerun`, { method: "POST" });
}

export async function analyzeFailure(id: string): Promise<FailureAnalysis> {
  return apiFetch<FailureAnalysis>(`/api/v1/runs/${id}/analyze-failure`, { method: "POST" });
}

export function reportUrl(id: string): string {
  return `${API_BASE}/api/v1/runs/${id}/report`;
}
export function exportUrl(id: string): string {
  return `${API_BASE}/api/v1/runs/${id}/export`;
}
export function exportJunitUrl(id: string): string {
  return `${API_BASE}/api/v1/runs/${id}/export-junit`;
}

// Advanced testing features
export async function runAudit(id: string): Promise<AuditResult> {
  return apiFetch<AuditResult>(`/api/v1/runs/${id}/audit`, { method: "POST" });
}

export async function runExploratory(id: string): Promise<{ pages_visited: number; actions_attempted: number; new_tests_generated: number }> {
  return apiFetch(`/api/v1/runs/${id}/explore`, { method: "POST" });
}

export async function getPerformanceMetrics(id: string): Promise<PerformanceMetrics> {
  return apiFetch<PerformanceMetrics>(`/api/v1/runs/${id}/performance`);
}

export async function getAccessibilityReport(id: string): Promise<AccessibilityResult> {
  return apiFetch<AccessibilityResult>(`/api/v1/runs/${id}/accessibility`);
}

export async function runVisualRegression(id: string): Promise<VisualRegressionResult> {
  return apiFetch<VisualRegressionResult>(`/api/v1/runs/${id}/visual-regression`, { method: "POST" });
}

export async function exportCode(id: string, language: string = "playwright"): Promise<ExportCodeResult> {
  return apiFetch<ExportCodeResult>(`/api/v1/runs/${id}/export-code?language=${encodeURIComponent(language)}`);
}

export async function cancelRun(id: string): Promise<{ status: string; message: string }> {
  return apiFetch(`/api/v1/runs/${id}/cancel`, { method: "POST" });
}

// Urutan fase eksekusi untuk membangun timeline
export const PHASES = [
  "analyzing",
  "plan_generated",
  "writing_tests",
  "running",
  "fixing",
  "done",
] as const;

// isActive: run belum mencapai terminal state (belum done/failed/simulated).
// Dipakai untuk filter "belum selesai" dan SSE subscription.
export function isActive(state: string): boolean {
  return !["done", "failed", "simulated"].includes(state);
}

// isExecuting: run BENAR-BENAR sedang dieksekusi oleh engine.
// State "idle" (dibuat tapi belum/tidak pernah jalan) TIDAK termasuk.
// Dipakai untuk label "Running Now" di dashboard agar tidak menyesatkan.
export function isExecuting(state: string): boolean {
  return ["analyzing", "plan_generated", "writing_tests", "running", "fixing"].includes(state);
}

export function subscribeToRun(
  id: string,
  onEvent: (event: { type: string; data: Record<string, string> }) => void
): () => void {
  const es = new EventSource(`${API_BASE}/api/v1/runs/${id}/stream`);

  es.addEventListener("state_change", (e) => {
    onEvent({ type: "state_change", data: JSON.parse(e.data) });
  });

  es.addEventListener("step", (e) => {
    onEvent({ type: "step", data: JSON.parse(e.data) });
  });

  es.addEventListener("done", (e) => {
    onEvent({ type: "done", data: JSON.parse(e.data) });
    es.close();
  });

  es.onerror = () => {
    es.close();
  };

  return () => es.close();
}

// Step-level events
export interface RunEvent {
  id: string;
  run_id: string;
  type: string;
  phase: string;
  message: string;
  metadata?: Record<string, string>;
  timestamp: string;
}

export async function getRunEvents(id: string): Promise<RunEvent[]> {
  return apiFetch<RunEvent[]>(`/api/v1/runs/${id}/events`);
}

// Recordings
export interface Recording {
  id: string;
  run_id: string;
  test_name: string;
  step_name: string;
  screenshot_url: string;
  start_time: string;
  end_time: string;
  status: string;
}

export async function getRunRecordings(id: string): Promise<Recording[]> {
  return apiFetch<Recording[]>(`/api/v1/runs/${id}/recordings`);
}

export async function getRecordings(): Promise<Recording[]> {
  return apiFetch<Recording[]>("/api/v1/recordings");
}

// Visual artifacts
export interface VisualArtifact {
  id: string;
  run_id: string;
  step_name: string;
  baseline_url?: string;
  current_url?: string;
  diff_url?: string;
  similarity_score: number;
  passed: boolean;
  created_at: string;
}

export async function getRunVisuals(id: string): Promise<VisualArtifact[]> {
  return apiFetch<VisualArtifact[]>(`/api/v1/runs/${id}/visual`);
}

// Compare
export interface CompareResult {
  run_a: string;
  run_b: string;
  summary: string;
  total_delta: number;
  passed_delta: number;
  failed_delta: number;
  new_failures: string[];
  recovered: string[];
  common_failures: string[];
  screenshot_diff: number;
}

export async function compareRuns(idA: string, idB: string): Promise<CompareResult> {
  return apiFetch<CompareResult>(`/api/v1/runs/${idA}/compare/${idB}`);
}

// --- Schedules ---
export interface Schedule {
  id: string;
  project_id?: string;
  test_list_id?: string;
  name: string;
  project_path: string;
  requirements: string;
  mode: string;
  environment: string;
  base_url: string;
  frequency: string;
  enabled: boolean;
  next_run_at: string;
  last_run_at?: string;
  last_run_id?: string;
  last_run_status?: string;
  notify_on_fail: boolean;
  webhook_url?: string;
  created_at: string;
}

export async function getSchedules(): Promise<Schedule[]> {
  return apiFetch<Schedule[]>("/api/v1/schedules");
}

export async function createSchedule(data: Partial<Schedule>): Promise<Schedule> {
  return apiFetch("/api/v1/schedules", { method: "POST", body: JSON.stringify(data) });
}

export async function updateSchedule(id: string, data: Partial<Schedule>): Promise<Schedule> {
  return apiFetch(`/api/v1/schedules/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteSchedule(id: string): Promise<void> {
  await fetch(`${API_BASE}/api/v1/schedules/${id}`, { method: "DELETE" });
}

export async function runScheduleNow(id: string): Promise<{ run_id: string; run_ids?: string[]; test_list_id?: string; state?: string }> {
  return apiFetch(`/api/v1/schedules/${id}/run-now`, { method: "POST" });
}

// --- Releases ---
export interface Release {
  id: string;
  name: string;
  version: string;
  project_id: string;
  status: string;
  run_ids: string[];
  created_at: string;
}

export interface ReleaseSummary {
  release_id: string;
  total_runs: number;
  passed_runs: number;
  failed_runs: number;
  pass_rate: number;
  total_tests: number;
  total_passed: number;
  total_failed: number;
  latest_status: string;
}

export async function getReleases(): Promise<Release[]> {
  return apiFetch<Release[]>("/api/v1/releases");
}

export async function getReleaseSummary(id: string): Promise<ReleaseSummary> {
  return apiFetch<ReleaseSummary>(`/api/v1/releases/${id}/summary`);
}

// --- Alerts (dedicated /api/v1/alerts endpoint with server-side severity/category) ---
export interface Alert {
  id: string;
  run_id: string;
  schedule_id?: string;
  type: string;
  severity?: string; // "critical" | "warning" | "info"
  category?: string; // "failure" | "drift" | "system"
  message: string;
  delivered: boolean;
  acknowledged?: boolean;
  dismissed?: boolean;
  created_at: string;
}

export async function getAlerts(params?: { type?: string; limit?: number; include_dismissed?: boolean }): Promise<Alert[]> {
  const qs = new URLSearchParams();
  if (params?.type) qs.set("type", params.type);
  if (params?.limit) qs.set("limit", String(params.limit));
  if (params?.include_dismissed) qs.set("include_dismissed", "true");
  const q = qs.toString();
  return apiFetch<Alert[]>(`/api/v1/alerts${q ? `?${q}` : ""}`);
}

export async function acknowledgeAlert(id: string): Promise<{ acknowledged: boolean }> {
  return apiFetch(`/api/v1/alerts/${id}/acknowledge`, { method: "POST" });
}

export async function dismissAlert(id: string): Promise<{ dismissed: boolean }> {
  return apiFetch(`/api/v1/alerts/${id}/dismiss`, { method: "POST" });
}

export async function markAllAlertsRead(): Promise<{ updated: number }> {
  return apiFetch("/api/v1/alerts/read-all", { method: "POST" });
}

// --- Notifications ---
export interface Notification {
  id: string;
  run_id: string;
  schedule_id?: string;
  type: string;
  message: string;
  delivered: boolean;
  created_at: string;
}

export async function getNotifications(): Promise<Notification[]> {
  return apiFetch<Notification[]>("/api/v1/notifications");
}

// --- Metrics ---
export interface MetricsSummary {
  total_runs: number;
  pass_rate: number;
  avg_duration_ms: number;
  total_tests: number;
  total_passed: number;
  total_failed: number;
}

export interface Hotspot {
  test_name: string;
  fail_count: number;
  fail_rate: number;
}

export interface FlakyTest {
  test_name: string;
  flip_count: number;
  total_appearances: number;
}

export async function getMetricsSummary(): Promise<MetricsSummary> {
  return apiFetch<MetricsSummary>("/api/v1/metrics/summary");
}

export async function getMetricsHotspots(): Promise<Hotspot[]> {
  return apiFetch<Hotspot[]>("/api/v1/metrics/hotspots");
}

export async function getMetricsFlaky(): Promise<FlakyTest[]> {
  return apiFetch<FlakyTest[]>("/api/v1/metrics/flaky");
}

export interface TrendPoint {
  date: string;
  pass_rate: number;
  fail_count: number;
  total_tests: number;
  duration_ms: number;
}

export async function getMetricsTrend(): Promise<TrendPoint[]> {
  return apiFetch<TrendPoint[]>("/api/v1/metrics/trend");
}

// --- Intelligence ---
export interface RiskItem {
  name: string;
  type: string;
  risk_score: number;
  reason: string;
  environment?: string;
}

export interface Recommendation {
  action: string;
  target: string;
  reason: string;
  priority: number;
}

export interface ConfidenceScore {
  score: number;
  grade: string;
  pass_rate: number;
  risk_score: number;
  freshness: number;
  explanation: string;
}

export async function getMetricsRisk(): Promise<RiskItem[]> {
  return apiFetch<RiskItem[]>("/api/v1/metrics/risk");
}

export async function getRecommendations(): Promise<Recommendation[]> {
  return apiFetch<Recommendation[]>("/api/v1/metrics/recommendations");
}

export async function getReleaseConfidence(id: string): Promise<ConfidenceScore> {
  return apiFetch<ConfidenceScore>(`/api/v1/releases/${id}/confidence`);
}

export async function getReleaseExplanation(id: string): Promise<{ confidence: ConfidenceScore; factors: { factor: string; value: number; impact: string; detail: string }[] }> {
  return apiFetch(`/api/v1/releases/${id}/explanation`);
}

// --- Reviews ---
export interface Review {
  id: string;
  run_id: string;
  type: string;
  status: string;
  reviewer?: string;
  comment?: string;
  created_at: string;
}

export async function getRunReviews(runId: string): Promise<Review[]> {
  return apiFetch<Review[]>(`/api/v1/runs/${runId}/reviews`);
}

export async function getAllReviews(): Promise<Review[]> {
  return apiFetch<Review[]>("/api/v1/reviews");
}

export async function approveReview(id: string, reviewer: string, comment: string): Promise<Review> {
  return apiFetch(`/api/v1/reviews/${id}/approve`, { method: "POST", body: JSON.stringify({ reviewer, comment }) });
}

export async function rejectReview(id: string, reviewer: string, comment: string): Promise<Review> {
  return apiFetch(`/api/v1/reviews/${id}/reject`, { method: "POST", body: JSON.stringify({ reviewer, comment }) });
}

// --- Suites ---
export interface Suite {
  id: string;
  name: string;
  project_id?: string;
  environment?: string;
  tags: string[];
  pinned: boolean;
  run_ids: string[];
  created_at: string;
}

export async function getSuites(): Promise<Suite[]> {
  return apiFetch<Suite[]>("/api/v1/suites");
}

export async function createSuite(data: Partial<Suite>): Promise<Suite> {
  return apiFetch<Suite>("/api/v1/suites", { method: "POST", body: JSON.stringify(data) });
}

// --- Demo data ---
export async function seedDemoData(): Promise<{ message: string; runs: number }> {
  return apiFetch("/api/v1/demo/seed", { method: "POST" });
}

// --- Settings & AI Providers ---
export interface AIProvider {
  id: string;
  name: string;
  models: string[];
}

export async function getAIProviders(): Promise<AIProvider[]> {
  return apiFetch<AIProvider[]>("/api/v1/ai/providers");
}

export async function testAIProvider(data: { provider: string; model: string; api_key: string; base_url: string }, signal?: AbortSignal): Promise<{ success: boolean; error?: string }> {
  return apiFetch<{ success: boolean; error?: string }>("/api/v1/ai/test-provider", { method: "POST", body: JSON.stringify(data), signal });
}

export async function getAIModels(data: { provider: string; api_key: string; base_url: string }): Promise<{ models: string[]; error?: string }> {
  return apiFetch<{ models: string[]; error?: string }>("/api/v1/ai/models", { method: "POST", body: JSON.stringify(data) });
}

export async function getSettings(): Promise<Record<string, string>> {
  return apiFetch<Record<string, string>>("/api/v1/settings");
}

export async function saveSettings(data: Record<string, string>): Promise<Record<string, string>> {
  return apiFetch<Record<string, string>>("/api/v1/settings", { method: "PUT", body: JSON.stringify(data) });
}

export interface LLMProfile {
  id: string;
  name: string;
  provider: string;
  base_url: string;
  api_key: string;
  model: string;
  temperature: string;
  max_tokens: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export async function getLLMProfiles(): Promise<LLMProfile[]> {
  return apiFetch<LLMProfile[]>("/api/v1/ai/profiles");
}

export async function createLLMProfile(data: Partial<LLMProfile>): Promise<LLMProfile> {
  return apiFetch<LLMProfile>("/api/v1/ai/profiles", { method: "POST", body: JSON.stringify(data) });
}

export async function updateLLMProfile(id: string, data: Partial<LLMProfile>): Promise<LLMProfile> {
  return apiFetch<LLMProfile>(`/api/v1/ai/profiles/${id}`, { method: "PUT", body: JSON.stringify(data) });
}

export async function deleteLLMProfile(id: string): Promise<void> {
  await apiFetch<unknown>(`/api/v1/ai/profiles/${id}`, { method: "DELETE" });
}

export async function activateLLMProfile(id: string): Promise<{ status: string }> {
  return apiFetch<{ status: string }>(`/api/v1/ai/profiles/${id}/activate`, { method: "POST" });
}

export async function testLLMProfile(id: string): Promise<{ success: boolean; error?: string }> {
  return apiFetch<{ success: boolean; error?: string }>(`/api/v1/ai/profiles/${id}/test`, { method: "POST" });
}

// --- Recording Sessions ---
export interface RecordingSession {
  id: string;
  name: string;
  project_path: string;
  base_url: string;
  status: string;
  metadata?: Record<string, unknown>;
  event_count?: number;
  created_at: string;
  updated_at: string;
}

export interface RecordedEvent {
  id: string;
  session_id: string;
  event_type: string;
  selector?: string;
  value?: string;
  url?: string;
  timestamp: string;
  sequence_order: number;
}

export async function listRecordingSessions(): Promise<RecordingSession[]> {
  return apiFetch<RecordingSession[]>("/api/v1/recording-sessions");
}

export async function getRecordingSession(id: string): Promise<{ session: RecordingSession; events: RecordedEvent[] }> {
  return apiFetch<{ session: RecordingSession; events: RecordedEvent[] }>(`/api/v1/recording-sessions/${id}`);
}

export async function createRecordingSession(data: { name: string; project_path: string; base_url: string }): Promise<RecordingSession> {
  return apiFetch<RecordingSession>("/api/v1/recording-sessions", { method: "POST", body: JSON.stringify(data) });
}

export async function addRecordingEvent(sessionId: string, event: Omit<RecordedEvent, "id" | "session_id" | "timestamp">): Promise<RecordedEvent> {
  return apiFetch<RecordedEvent>(`/api/v1/recording-sessions/${sessionId}/events`, { method: "POST", body: JSON.stringify(event) });
}

export async function generateRecordingTest(sessionId: string): Promise<{ test_code: string; language: string; framework: string }> {
  return apiFetch<{ test_code: string; language: string; framework: string }>(`/api/v1/recording-sessions/${sessionId}/generate`, { method: "POST" });
}

export async function deleteRecordingSession(id: string): Promise<void> {
  await fetch(`${API_BASE}/api/v1/recording-sessions/${id}`, { method: "DELETE", credentials: "include" });
}

export async function updateRecordingSession(id: string, data: Partial<RecordingSession>): Promise<RecordingSession> {
  return apiFetch<RecordingSession>(`/api/v1/recording-sessions/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

// --- API Key Management (admin-only) ---

export interface APIKeyEntry {
  id: string;
  key?: string;        // only returned on creation
  label: string;
  role: string;
  active: boolean;
  created_at: string;
  created_by?: string;
}

export async function createAPIKey(label: string, role: string): Promise<APIKeyEntry> {
  return apiFetch<APIKeyEntry>("/api/v1/keys", { method: "POST", body: JSON.stringify({ label, role }) });
}

export async function listAPIKeys(): Promise<APIKeyEntry[]> {
  return apiFetch<APIKeyEntry[]>("/api/v1/keys");
}

export async function revokeAPIKey(id: string, active: boolean): Promise<{ status: string }> {
  return apiFetch<{ status: string }>(`/api/v1/keys/${id}/revoke`, { method: "POST", body: JSON.stringify({ active }) });
}

export async function deleteAPIKey(id: string): Promise<void> {
  await fetch(`${API_BASE}/api/v1/keys/${id}`, { method: "DELETE", credentials: "include" });
}

// --- Audit Log (admin-only) ---

export interface AuditEntry {
  id: string;
  actor_id: string;
  actor_role: string;
  action: string;
  resource: string;
  resource_id: string;
  detail: string;
  created_at: string;
}

export async function listAuditLog(): Promise<AuditEntry[]> {
  return apiFetch<AuditEntry[]>("/api/v1/audit-log");
}

// --- User Role (from JWT session) ---

// getUserRole returns the current user's role from sessionStorage.
// Returns "admin" as default for backward compatibility.
export function getUserRole(): string {
  if (typeof window === "undefined") return "admin";
  return sessionStorage.getItem("user_role") || "admin";
}

// getUserLabel returns a human-readable label for the current session.
export function getUserLabel(): string {
  if (typeof window === "undefined") return "";
  return sessionStorage.getItem("user_label") || "";
}

// isAdmin returns true if the current user has admin role.
export function isAdmin(): boolean {
  return getUserRole() === "admin";
}

// isReviewerOrAbove returns true if the user can approve/review.
export function isReviewerOrAbove(): boolean {
  const role = getUserRole();
  return role === "admin" || role === "reviewer";
}

