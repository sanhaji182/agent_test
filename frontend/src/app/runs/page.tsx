"use client";

import { useEffect, useState } from "react";
import { getRuns, type TestRun } from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import Link from "next/link";
import { Search, Inbox, ExternalLink } from "lucide-react";

export default function RunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState("");

  useEffect(() => {
    getRuns()
      .then(setRuns)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const filtered = runs.filter(
    (r) =>
      r.id.includes(filter) ||
      r.state.includes(filter) ||
      (r.requirements || "").toLowerCase().includes(filter.toLowerCase())
  );

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" />
        <LoadingSkeleton rows={8} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5">
        <p className="text-sm font-medium text-[var(--danger)]">Error: {error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Test Runs</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-0.5">
            {runs.length} total runs
          </p>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <input
          type="text"
          placeholder="Filter by ID, state, or requirements..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-1 focus:ring-[var(--accent)]/20"
        />
      </div>

      {/* Table */}
      {filtered.length === 0 ? (
        <EmptyState
          icon={<Inbox className="w-6 h-6" />}
          title={filter ? "No matching runs" : "No test runs yet"}
          description={filter ? "Try a different search term." : "Create a run via the API or MCP."}
        />
      ) : (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[var(--border)] bg-[var(--bg-subtle)]">
                <th className="text-left px-4 py-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Run ID</th>
                <th className="text-left px-4 py-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Status</th>
                <th className="text-left px-4 py-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Result</th>
                <th className="text-left px-4 py-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Fixes</th>
                <th className="text-left px-4 py-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Created</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((run) => (
                <tr
                  key={run.id}
                  className="border-b border-[var(--border)] last:border-0 hover:bg-[var(--bg-hover)] transition-colors"
                >
                  <td className="px-4 py-3">
                    <Link
                      href={`/runs/${run.id}`}
                      className="font-mono text-xs text-[var(--accent)] hover:text-[var(--accent-hover)]"
                    >
                      {run.id.slice(0, 8)}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge state={run.state} />
                  </td>
                  <td className="px-4 py-3 text-xs">
                    {run.run_result ? (
                      <span className="flex items-center gap-2">
                        <span className="text-[var(--success)] font-medium">{run.run_result.passed}✓</span>
                        <span className="text-[var(--danger)] font-medium">{run.run_result.failed}✗</span>
                        <span className="text-[var(--text-muted)]">/ {run.run_result.total}</span>
                      </span>
                    ) : (
                      <span className="text-[var(--text-muted)]">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-xs text-[var(--text-secondary)]">
                    {run.fix_attempts > 0 ? run.fix_attempts : "—"}
                  </td>
                  <td className="px-4 py-3 text-[11px] text-[var(--text-muted)]">
                    {new Date(run.created_at).toLocaleString(undefined, {
                      month: "short", day: "numeric", hour: "2-digit", minute: "2-digit",
                    })}
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      href={`/runs/${run.id}`}
                      className="p-1.5 rounded-md hover:bg-[var(--bg-subtle)] text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                    </Link>
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
