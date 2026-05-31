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
