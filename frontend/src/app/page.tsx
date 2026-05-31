"use client";

import { useEffect, useState } from "react";
import { getRuns, type TestRun } from "@/lib/api";
import { StateBadge } from "@/components/state-badge";
import Link from "next/link";

export default function DashboardPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getRuns()
      .then(setRuns)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSkeleton />;
  if (error) return <ErrorState message={error} />;

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const running = runs.filter(
    (r) => !["done", "failed", "idle"].includes(r.state)
  ).length;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Runs" value={total} />
        <StatCard label="Passed" value={passed} className="text-green-600" />
        <StatCard label="Failed" value={failed} className="text-red-600" />
        <StatCard label="Running" value={running} className="text-yellow-600" />
      </div>

      <h2 className="text-lg font-semibold mb-3">Recent Runs</h2>
      {runs.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-zinc-50 dark:bg-zinc-800">
              <tr>
                <th className="text-left px-4 py-2">ID</th>
                <th className="text-left px-4 py-2">State</th>
                <th className="text-left px-4 py-2">Created</th>
              </tr>
            </thead>
            <tbody>
              {runs.slice(0, 10).map((run) => (
                <tr
                  key={run.id}
                  className="border-t hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
                >
                  <td className="px-4 py-2">
                    <Link
                      href={`/runs/${run.id}`}
                      className="text-blue-600 hover:underline font-mono text-xs"
                    >
                      {run.id.slice(0, 8)}
                    </Link>
                  </td>
                  <td className="px-4 py-2">
                    <StateBadge state={run.state} />
                  </td>
                  <td className="px-4 py-2 text-zinc-500">
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

function StatCard({
  label,
  value,
  className,
}: {
  label: string;
  value: number;
  className?: string;
}) {
  return (
    <div className="border rounded-lg p-4">
      <p className="text-sm text-zinc-500">{label}</p>
      <p className={`text-3xl font-bold ${className || ""}`}>{value}</p>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-8 w-48 bg-zinc-200 dark:bg-zinc-800 rounded animate-pulse" />
      <div className="grid grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => (
          <div
            key={i}
            className="h-20 bg-zinc-200 dark:bg-zinc-800 rounded animate-pulse"
          />
        ))}
      </div>
    </div>
  );
}

function EmptyState() {
  return (
    <div className="text-center py-12 text-zinc-500">
      <p className="text-lg">No test runs yet</p>
      <p className="text-sm mt-1">
        Create a run via the API or MCP to get started.
      </p>
    </div>
  );
}

function ErrorState({ message }: { message: string }) {
  return (
    <div className="border border-red-200 bg-red-50 dark:bg-red-900/20 rounded-lg p-4">
      <p className="text-red-700 dark:text-red-400 font-medium">
        Failed to load dashboard
      </p>
      <p className="text-sm text-red-600 dark:text-red-300 mt-1">{message}</p>
    </div>
  );
}
