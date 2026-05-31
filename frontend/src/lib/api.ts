const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface TestRun {
  id: string;
  project_path: string;
  requirements: string;
  mode: string;
  state: string;
  code_analysis?: string;
  test_plan?: TestPlan;
  test_files?: TestFile[];
  run_result?: RunResult;
  screenshots?: string[];
  fix_attempts: number;
  error?: string;
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

export interface TestFile {
  name: string;
  content: string;
}

export interface RunResult {
  passed: number;
  failed: number;
  total: number;
  failures: Failure[];
}

export interface Failure {
  test: string;
  message: string;
  screenshot_url?: string;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
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
}): Promise<{ run_id: string; state: string }> {
  return apiFetch("/api/v1/runs", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// Rerun: buat run baru dari konfigurasi run lama
export async function rerunRun(id: string): Promise<{ run_id: string; state: string }> {
  return apiFetch(`/api/v1/runs/${id}/rerun`, { method: "POST" });
}

export function reportUrl(id: string): string {
  return `${API_BASE}/api/v1/runs/${id}/report`;
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

// isActive: run masih berjalan (belum done/failed)
export function isActive(state: string): boolean {
  return !["done", "failed", "idle"].includes(state);
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

export async function runScheduleNow(id: string): Promise<{ run_id: string }> {
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
  total_tests: number;
  total_passed: number;
  total_failed: number;
}

export interface Hotspot {
  test_name: string;
  fail_count: number;
  fail_rate: number;
}

export async function getMetricsSummary(): Promise<MetricsSummary> {
  return apiFetch<MetricsSummary>("/api/v1/metrics/summary");
}

export async function getMetricsHotspots(): Promise<Hotspot[]> {
  return apiFetch<Hotspot[]>("/api/v1/metrics/hotspots");
}

export async function getMetricsTrend(): Promise<{ date: string; pass_rate: number; fail_count: number }[]> {
  return apiFetch("/api/v1/metrics/trend");
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
  tags: string[];
  pinned: boolean;
  run_ids: string[];
}

export async function getSuites(): Promise<Suite[]> {
  return apiFetch<Suite[]>("/api/v1/suites");
}
