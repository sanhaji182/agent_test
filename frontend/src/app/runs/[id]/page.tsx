"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getRun, subscribeToRun, type TestRun } from "@/lib/api";
import { StateBadge } from "@/components/state-badge";

export default function RunDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const [run, setRun] = useState<TestRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [liveState, setLiveState] = useState<string | null>(null);

  useEffect(() => {
    getRun(id)
      .then((r) => {
        setRun(r);
        // Subscribe to SSE if run is still in progress
        if (!["done", "failed"].includes(r.state)) {
          const unsub = subscribeToRun(id, (event) => {
            if (event.type === "state_change") {
              setLiveState(event.data.state);
            }
            if (event.type === "done") {
              getRun(id).then(setRun);
            }
          });
          return unsub;
        }
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-64 bg-zinc-200 dark:bg-zinc-800 rounded animate-pulse" />
        <div className="h-40 bg-zinc-200 dark:bg-zinc-800 rounded animate-pulse" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="border border-red-200 bg-red-50 rounded-lg p-4">
        <p className="text-red-700">Error loading run: {error}</p>
      </div>
    );
  }

  if (!run) return null;

  const displayState = liveState || run.state;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Run {run.id.slice(0, 8)}</h1>
          <p className="text-sm text-zinc-500 font-mono">{run.id}</p>
        </div>
        <div className="flex items-center gap-3">
          <StateBadge state={displayState} />
          <a
            href={`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/v1/runs/${run.id}/report`}
            target="_blank"
            className="text-sm text-blue-600 hover:underline"
          >
            HTML Report ↗
          </a>
        </div>
      </div>

      {/* Summary */}
      <Section title="Summary">
        <dl className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
          <DT label="Mode" value={run.mode || "simple"} />
          <DT label="Fix Attempts" value={String(run.fix_attempts)} />
          <DT
            label="Created"
            value={new Date(run.created_at).toLocaleString()}
          />
          <DT
            label="Finished"
            value={
              run.finished_at
                ? new Date(run.finished_at).toLocaleString()
                : "—"
            }
          />
        </dl>
      </Section>

      {/* Results */}
      {run.run_result && (
        <Section title="Results">
          <div className="flex gap-6 mb-4">
            <div className="text-center">
              <p className="text-2xl font-bold text-green-600">
                {run.run_result.passed}
              </p>
              <p className="text-xs text-zinc-500">Passed</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-red-600">
                {run.run_result.failed}
              </p>
              <p className="text-xs text-zinc-500">Failed</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold">{run.run_result.total}</p>
              <p className="text-xs text-zinc-500">Total</p>
            </div>
          </div>

          {run.run_result.failures.length > 0 && (
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-zinc-50 dark:bg-zinc-800">
                  <tr>
                    <th className="text-left px-4 py-2">Test</th>
                    <th className="text-left px-4 py-2">Error</th>
                    <th className="text-left px-4 py-2">Screenshot</th>
                  </tr>
                </thead>
                <tbody>
                  {run.run_result.failures.map((f, i) => (
                    <tr key={i} className="border-t">
                      <td className="px-4 py-2 font-mono text-xs">{f.test}</td>
                      <td className="px-4 py-2 text-red-600 text-xs">
                        {f.message}
                      </td>
                      <td className="px-4 py-2">
                        {f.screenshot_url ? (
                          <a
                            href={f.screenshot_url}
                            target="_blank"
                            className="text-blue-600 hover:underline text-xs"
                          >
                            View
                          </a>
                        ) : (
                          "—"
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Section>
      )}

      {/* Test Plan */}
      {run.test_plan && (
        <Section title="Test Plan">
          <p className="text-sm text-zinc-600 mb-3">{run.test_plan.summary}</p>
          <ul className="space-y-2">
            {run.test_plan.scenarios.map((s, i) => (
              <li key={i} className="border rounded p-3">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm">{s.name}</span>
                  <span className="text-xs px-1.5 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800">
                    {s.priority}
                  </span>
                </div>
                <ul className="mt-1 text-xs text-zinc-500 list-disc list-inside">
                  {s.steps.map((step, j) => (
                    <li key={j}>{step}</li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </Section>
      )}

      {/* Test Files */}
      {run.test_files && run.test_files.length > 0 && (
        <Section title="Generated Test Files">
          <div className="space-y-3">
            {run.test_files.map((f, i) => (
              <details key={i} className="border rounded">
                <summary className="px-4 py-2 cursor-pointer text-sm font-mono hover:bg-zinc-50 dark:hover:bg-zinc-800">
                  {f.name}
                </summary>
                <pre className="px-4 py-3 text-xs overflow-x-auto bg-zinc-50 dark:bg-zinc-900 border-t">
                  {f.content}
                </pre>
              </details>
            ))}
          </div>
        </Section>
      )}

      {/* Screenshots */}
      {run.screenshots && run.screenshots.length > 0 && (
        <Section title="Screenshots">
          <div className="grid grid-cols-2 gap-3">
            {run.screenshots.map((url, i) => (
              <a
                key={i}
                href={url}
                target="_blank"
                className="border rounded p-2 text-xs text-blue-600 hover:underline"
              >
                {url.split("/").pop()}
              </a>
            ))}
          </div>
        </Section>
      )}

      {/* Error */}
      {run.error && (
        <Section title="Error">
          <pre className="text-sm text-red-600 bg-red-50 dark:bg-red-900/20 p-3 rounded overflow-x-auto">
            {run.error}
          </pre>
        </Section>
      )}
    </div>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="border rounded-lg p-4">
      <h2 className="text-sm font-semibold text-zinc-500 uppercase tracking-wide mb-3">
        {title}
      </h2>
      {children}
    </section>
  );
}

function DT({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-zinc-500">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}
