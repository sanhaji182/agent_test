"use client";

import { useEffect, useMemo, useState } from "react";
import { getRuns, isActive, type TestRun } from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { RunInspector } from "@/components/console/inspector";
import { Search, Inbox, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

type Group = "all" | "active" | "passed" | "failed";
type Sort = "newest" | "oldest";

export default function RunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [group, setGroup] = useState<Group>("all");
  const [sort, setSort] = useState<Sort>("newest");
  const [selected, setSelected] = useState<TestRun | null>(null);

  useEffect(() => {
    getRuns().then((r) => setRuns(r || [])).catch((e) => setError(e.message)).finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    let list = runs.filter(
      (r) =>
        r.id.includes(q) ||
        r.state.includes(q) ||
        (r.requirements || "").toLowerCase().includes(q) ||
        (r.project_path || "").toLowerCase().includes(q)
    );
    if (group === "active") list = list.filter((r) => isActive(r.state));
    if (group === "passed") list = list.filter((r) => r.state === "done");
    if (group === "failed") list = list.filter((r) => r.state === "failed");
    list = [...list].sort((a, b) => {
      const d = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      return sort === "newest" ? -d : d;
    });
    return list;
  }, [runs, query, group, sort]);

  const counts = {
    all: runs.length,
    active: runs.filter((r) => isActive(r.state)).length,
    passed: runs.filter((r) => r.state === "done").length,
    failed: runs.filter((r) => r.state === "failed").length,
  };

  if (loading) {
    return <div className="space-y-6"><div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" /><LoadingSkeleton rows={8} /></div>;
  }
  if (error) {
    return <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5 text-sm text-[var(--danger)]">Error: {error}</div>;
  }

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-xl font-bold">Test Suites</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Browse, inspect, and rerun your test sessions</p>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input
            type="text"
            placeholder="Search by ID, state, requirements, or project..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 rounded-lg bg-[var(--bg-card)] border border-[var(--border)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-2 focus:ring-[var(--accent)]/10"
          />
        </div>
        <div className="relative">
          <select
            value={sort}
            onChange={(e) => setSort(e.target.value as Sort)}
            className="appearance-none pl-3 pr-8 py-2.5 rounded-lg bg-[var(--bg-card)] border border-[var(--border)] text-sm focus:outline-none focus:border-[var(--accent)]/50 cursor-pointer"
          >
            <option value="newest">Newest first</option>
            <option value="oldest">Oldest first</option>
          </select>
          <ChevronDown className="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] pointer-events-none" />
        </div>
      </div>

      {/* Status filter tabs */}
      <div className="flex items-center gap-1 border-b border-[var(--border)]">
        {(["all", "active", "passed", "failed"] as Group[]).map((g) => (
          <button
            key={g}
            onClick={() => setGroup(g)}
            className={cn(
              "px-3 py-2 text-[13px] font-medium border-b-2 -mb-px capitalize transition-colors",
              group === g ? "border-[var(--accent)] text-[var(--accent)]" : "border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            )}
          >
            {g} <span className="text-[var(--text-muted)]">({counts[g]})</span>
          </button>
        ))}
      </div>

      {/* List */}
      {filtered.length === 0 ? (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
          <EmptyState 
            icon={<Inbox className="w-6 h-6" />} 
            title={query ? "No matching runs" : "No runs in this group"} 
            description={query ? "Try a different search." : "Runs will appear here. Create your first test to get started."} 
            action={!query && runs.length === 0 ? (
              <a href="/create" className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm">
                Create First Test
              </a>
            ) : undefined}
          />
        </div>
      ) : (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] overflow-hidden divide-y divide-[var(--border)]">
          {filtered.map((r) => (
            <button
              key={r.id}
              onClick={() => setSelected(r)}
              className={cn(
                "w-full flex items-center gap-4 px-4 py-3 text-left hover:bg-[var(--bg-hover)] transition-colors",
                selected?.id === r.id && "bg-[var(--accent-bg)]"
              )}
            >
              <StatusBadge state={r.state} />
              <span className="font-mono text-xs text-[var(--accent)] w-16 shrink-0">{r.id.slice(0, 8)}</span>
              <span className="text-sm text-[var(--text-secondary)] truncate flex-1">{r.requirements || "—"}</span>
              {r.run_result && (
                <span className="text-[11px] shrink-0 flex items-center gap-2">
                  <span className="text-[var(--success)]">{r.run_result.passed}✓</span>
                  <span className="text-[var(--danger)]">{r.run_result.failed}✗</span>
                </span>
              )}
              <span className="text-[11px] text-[var(--text-muted)] shrink-0 w-28 text-right">
                {new Date(r.created_at).toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}
              </span>
            </button>
          ))}
        </div>
      )}

      {/* Inspector drawer + backdrop */}
      {selected && <div className="fixed inset-0 z-30 bg-black/20" onClick={() => setSelected(null)} />}
      <RunInspector run={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
