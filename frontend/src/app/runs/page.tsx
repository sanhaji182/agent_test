"use client";

import { useEffect, useState } from "react";
import { getRuns, type TestRun } from "@/lib/api";
import { StateBadge } from "@/components/state-badge";
import Link from "next/link";

export default function RunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getRuns()
      .then(setRuns)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-6">Test Runs</h1>
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="h-12 bg-zinc-200 dark:bg-zinc-800 rounded animate-pulse"
            />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-6">Test Runs</h1>
        <div className="border border-red-200 bg-red-50 rounded-lg p-4">
          <p className="text-red-700">Error: {error}</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Test Runs</h1>

      {runs.length === 0 ? (
        <div className="text-center py-12 text-zinc-500">
          <p>No runs found.</p>
        </div>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-zinc-50 dark:bg-zinc-800">
              <tr>
                <th className="text-left px-4 py-3">Run ID</th>
                <th className="text-left px-4 py-3">State</th>
                <th className="text-left px-4 py-3">Result</th>
                <th className="text-left px-4 py-3">Fixes</th>
                <th className="text-left px-4 py-3">Created</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <tr
                  key={run.id}
                  className="border-t hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
                >
                  <td className="px-4 py-3">
                    <Link
                      href={`/runs/${run.id}`}
                      className="text-blue-600 hover:underline font-mono text-xs"
                    >
                      {run.id.slice(0, 8)}...
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <StateBadge state={run.state} />
                  </td>
                  <td className="px-4 py-3">
                    {run.run_result ? (
                      <span>
                        <span className="text-green-600">
                          {run.run_result.passed}✓
                        </span>{" "}
                        <span className="text-red-600">
                          {run.run_result.failed}✗
                        </span>{" "}
                        / {run.run_result.total}
                      </span>
                    ) : (
                      <span className="text-zinc-400">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-zinc-500">
                    {run.fix_attempts}
                  </td>
                  <td className="px-4 py-3 text-zinc-500 text-xs">
                    {new Date(run.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
