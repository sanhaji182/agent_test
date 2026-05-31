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

export function subscribeToRun(
  id: string,
  onEvent: (event: { type: string; data: Record<string, string> }) => void
): () => void {
  const es = new EventSource(`${API_BASE}/api/v1/runs/${id}/stream`);

  es.addEventListener("state_change", (e) => {
    onEvent({ type: "state_change", data: JSON.parse(e.data) });
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
