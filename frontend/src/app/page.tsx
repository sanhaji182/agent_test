"use client";

import { useEffect, useState } from "react";
import { getRuns, isActive, type TestRun } from "@/lib/api";
import { StatCard } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, Section, LoadingSkeleton } from "@/components/ui/section";
import { ExecutionTimeline } from "@/components/console/timeline";
import Link from "next/link";
import {
  Activity, CheckCircle2, XCircle, PlayCircle, Inbox,
  Flame, Layers, Film, ArrowRight, Terminal,
} from "lucide-react";

export default function DashboardPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getRuns().then(setRuns).catch((e) => setError(e.message)).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" />
        <div className="grid grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => <div key={i} className="h-24 rounded-xl bg-[var(--bg-subtle)] animate-pulse" />)}
        </div>
        <LoadingSkeleton rows={4} />
      </div>
    );
  }

  if (error) {
    return <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5 text-sm text-[var(--danger)]">Failed to load: {error}</div>;
  }

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const activeRuns = runs.filter((r) => isActive(r.state));
  const passRate = total > 0 ? Math.round((passed / total) * 100) : 0;

  // Failure hotspots: agregasi failure berdasarkan nama test (data nyata)
  const hotspots = aggregateFailures(runs);
  // Recent recordings: screenshot terbaru lintas run (data nyata)
  const recordings = runs.flatMap((r) => (r.screenshots || []).map((s) => ({ runId: r.id, url: s }))).slice(0, 6);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">Control Center</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Live testing activity and execution overview</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total Runs" value={total} icon={<Activity className="w-4 h-4" />} />
        <StatCard label="Pass Rate" value={`${passRate}%`} color="success" icon={<CheckCircle2 className="w-4 h-4" />} trend={`${passed} passed`} />
        <StatCard label="Failed" value={failed} color="danger" icon={<XCircle className="w-4 h-4" />} />
        <StatCard label="Active" value={activeRuns.length} color="warning" icon={<PlayCircle className="w-4 h-4" />} />
      </div>

      {/* Active execution */}
      <Section title="Active Execution" action={<span className="text-[11px] text-[var(--text-muted)]">{activeRuns.length} running</span>}>
        {activeRuns.length === 0 ? (
          <EmptyState icon={<PlayCircle className="w-6 h-6" />} title="No active runs" description="Runs in progress will appear here with a live execution timeline." />
        ) : (
          <div className="space-y-3">
            {activeRuns.map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="block p-4 rounded-lg border border-[var(--border)] hover:border-[var(--accent)]/40 hover:shadow-[var(--shadow-sm)] transition-all">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <StatusBadge state={r.state} />
                    <span className="font-mono text-xs text-[var(--text-secondary)]">{r.id.slice(0, 8)}</span>
                  </div>
                  <ArrowRight className="w-4 h-4 text-[var(--text-muted)]" />
                </div>
                <ExecutionTimeline state={r.state} />
              </Link>
            ))}
          </div>
        )}
      </Section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent suites */}
        <Section title="Recent Suites" action={<Link href="/runs" className="text-[11px] font-medium text-[var(--accent)]">View all →</Link>}>
          {runs.length === 0 ? (
            <EmptyState icon={<Layers className="w-6 h-6" />} title="No suites yet" description="Create a run via API or MCP." />
          ) : (
            <div className="space-y-1">
              {runs.slice(0, 6).map((r) => (
                <Link key={r.id} href={`/runs/${r.id}`} className="flex items-center justify-between px-3 py-2.5 rounded-lg hover:bg-[var(--bg-hover)] transition-colors">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <StatusBadge state={r.state} />
                    <span className="text-xs text-[var(--text-secondary)] truncate max-w-[160px]">{r.requirements || r.id.slice(0, 8)}</span>
                  </div>
                  {r.run_result && (
                    <span className="text-[11px] shrink-0">
                      <span className="text-[var(--success)]">{r.run_result.passed}</span>
                      <span className="text-[var(--text-muted)]">/{r.run_result.total}</span>
                    </span>
                  )}
                </Link>
              ))}
            </div>
          )}
        </Section>

        {/* Failure hotspots */}
        <Section title="Failure Hotspots">
          {hotspots.length === 0 ? (
            <EmptyState icon={<Flame className="w-6 h-6" />} title="No failures" description="Recurring test failures will surface here." />
          ) : (
            <div className="space-y-2">
              {hotspots.map((h, i) => (
                <div key={i} className="flex items-center justify-between px-3 py-2 rounded-lg bg-[var(--danger-bg)] border border-[var(--danger)]/10">
                  <span className="text-xs font-medium text-[var(--text-primary)] truncate">{h.test}</span>
                  <span className="text-[11px] font-semibold text-[var(--danger)] shrink-0 ml-2">{h.count}×</span>
                </div>
              ))}
            </div>
          )}
        </Section>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick actions */}
        <Section title="Quick Actions">
          <div className="grid grid-cols-2 gap-2">
            <QuickAction href="/runs" icon={<Layers className="w-4 h-4" />} label="Browse Suites" />
            <QuickAction href="/projects" icon={<Terminal className="w-4 h-4" />} label="Projects" />
            <QuickAction href="/settings" icon={<Activity className="w-4 h-4" />} label="Settings" />
            <QuickAction href="/runs" icon={<PlayCircle className="w-4 h-4" />} label="Recent Runs" />
          </div>
        </Section>

        {/* Recent recordings */}
        <Section title="Recent Recordings">
          {recordings.length === 0 ? (
            <EmptyState icon={<Film className="w-6 h-6" />} title="No recordings" description="Screenshots and recordings from runs will appear here." />
          ) : (
            <div className="grid grid-cols-3 gap-2">
              {recordings.map((rec, i) => (
                <Link key={i} href={`/runs/${rec.runId}`} className="aspect-video rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)] overflow-hidden hover:border-[var(--accent)] transition-colors">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={rec.url.startsWith("http") ? rec.url : `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}${rec.url}`} alt="recording" className="w-full h-full object-cover" />
                </Link>
              ))}
            </div>
          )}
        </Section>
      </div>

      {runs.length === 0 && (
        <EmptyState icon={<Inbox className="w-6 h-6" />} title="Welcome to GoTest Agent" description="Run your first test via the MCP tool in your IDE or the HTTP API." />
      )}
    </div>
  );
}

function QuickAction({ href, icon, label }: { href: string; icon: React.ReactNode; label: string }) {
  return (
    <Link href={href} className="flex items-center gap-2.5 px-3 py-3 rounded-lg border border-[var(--border)] hover:border-[var(--accent)]/40 hover:bg-[var(--bg-hover)] transition-all">
      <div className="text-[var(--accent)]">{icon}</div>
      <span className="text-xs font-medium">{label}</span>
    </Link>
  );
}

function aggregateFailures(runs: TestRun[]): { test: string; count: number }[] {
  const map = new Map<string, number>();
  for (const r of runs) {
    for (const f of r.run_result?.failures || []) {
      map.set(f.test, (map.get(f.test) || 0) + 1);
    }
  }
  return [...map.entries()]
    .map(([test, count]) => ({ test, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5);
}
