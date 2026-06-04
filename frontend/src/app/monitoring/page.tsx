"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getMonitoringSummary, type MonitoringSummary } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { StatusBadge } from "@/components/ui/badge";
import { Activity, CheckCircle2, Clock, Layers, PlayCircle, RotateCcw, TrendingDown, XCircle } from "lucide-react";

export default function MonitoringPage() {
  const [data, setData] = useState<MonitoringSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMonitoringSummary().then(setData).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="space-y-6"><LoadingSkeleton rows={7} /></div>;

  const summary = data?.summary || { total_lists: 0, total_cases: 0, active_runs: 0, failed_runs: 0, completed_runs: 0 };
  const lists = data?.lists || [];
  const recent = data?.recent_runs || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-lg font-bold">Monitoring</h1>
          <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Track Test List health, recent execution history, and current run status.</p>
        </div>
        <Link href="/suites" className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)]">
          <PlayCircle className="w-3.5 h-3.5" /> Test Lists
        </Link>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
        <Stat label="Lists" value={summary.total_lists} icon={<Layers className="w-4 h-4" />} />
        <Stat label="Cases" value={summary.total_cases} icon={<CheckCircle2 className="w-4 h-4" />} />
        <Stat label="Active" value={summary.active_runs} icon={<Activity className="w-4 h-4" />} color="warning" />
        <Stat label="Failed" value={summary.failed_runs} icon={<XCircle className="w-4 h-4" />} color="danger" />
        <Stat label="Completed" value={summary.completed_runs} icon={<CheckCircle2 className="w-4 h-4" />} color="success" />
      </div>

      <Section title="Test List Health" action={<span className="text-[11px] text-[var(--text-muted)]">{lists.length} lists</span>}>
        {lists.length === 0 ? (
          <EmptyState icon={<Layers className="w-6 h-6" />} title="No monitored lists" description="Create and run a Test List to populate monitoring health." />
        ) : (
          <div className="space-y-2">
            {lists.map((list) => (
              <div key={list.id} className="flex items-center gap-4 p-4 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold">{list.name}</span>
                    {list.pinned && <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-[var(--accent-bg)] text-[var(--accent)]">Pinned</span>}
                    <StatusBadge state={list.last_status} />
                  </div>
                  <div className="flex flex-wrap items-center gap-4 mt-1 text-[11px] text-[var(--text-muted)]">
                    <span>{list.test_count} tests</span>
                    <span>{Math.round((list.pass_rate || 0) * 100)}% pass rate</span>
                    <span>{list.passed} passed</span>
                    <span>{list.failed} failed</span>
                    {list.last_run_at && <span>Last: {new Date(list.last_run_at).toLocaleString()}</span>}
                  </div>
                  {(list.newly_failed?.length > 0 || list.recovered?.length > 0 || list.stable_failed?.length > 0) && (
                    <div className="flex flex-wrap items-center gap-1.5 mt-2">
                      {list.newly_failed?.length > 0 && <SignalChip tone="danger" icon={<TrendingDown className="w-3 h-3" />} label={`${list.newly_failed.length} newly failed`} />}
                      {list.recovered?.length > 0 && <SignalChip tone="success" icon={<RotateCcw className="w-3 h-3" />} label={`${list.recovered.length} recovered`} />}
                      {list.stable_failed?.length > 0 && <SignalChip tone="warning" icon={<XCircle className="w-3 h-3" />} label={`${list.stable_failed.length} stable failed`} />}
                    </div>
                  )}
                </div>
                {list.last_run_id && (
                  <Link href={`/runs/${list.last_run_id}`} className="text-[11px] font-semibold text-[var(--accent)] hover:underline">
                    Open latest
                  </Link>
                )}
              </div>
            ))}
          </div>
        )}
      </Section>

      <Section title="Recent Runs" action={<Link href="/runs" className="text-[11px] font-semibold text-[var(--accent)]">All runs</Link>}>
        {recent.length === 0 ? (
          <EmptyState icon={<Clock className="w-6 h-6" />} title="No run history" description="Runs appear here after executing cases or lists." />
        ) : (
          <div className="space-y-1">
            {recent.slice(0, 12).map((run) => (
              <Link key={run.id} href={`/runs/${run.id}`} className="flex items-center gap-3 px-3 py-2 rounded-[var(--radius-sm)] hover:bg-[var(--bg-hover)]">
                <StatusBadge state={run.state} />
                <span className="font-mono text-[11px] text-[var(--accent)] w-16 shrink-0">{run.id.slice(0, 8)}</span>
                <span className="text-[12px] text-[var(--text-secondary)] truncate flex-1">{run.test_plan?.summary || run.requirements || "Untitled run"}</span>
                <span className="text-[11px] text-[var(--text-muted)] shrink-0">{new Date(run.created_at).toLocaleString()}</span>
              </Link>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}

function SignalChip({ tone, icon, label }: { tone: "danger" | "success" | "warning"; icon: React.ReactNode; label: string }) {
  const cls = {
    danger: "border-[var(--danger)]/25 bg-[var(--danger)]/10 text-[var(--danger)]",
    success: "border-[var(--success)]/25 bg-[var(--success)]/10 text-[var(--success)]",
    warning: "border-[var(--warning)]/25 bg-[var(--warning)]/10 text-[var(--warning)]",
  }[tone];
  return <span className={`inline-flex items-center gap-1 rounded-[var(--radius-sm)] border px-1.5 py-0.5 text-[10px] font-semibold ${cls}`}>{icon}{label}</span>;
}

function Stat({ label, value, icon, color = "default" }: { label: string; value: number; icon: React.ReactNode; color?: string }) {
  const cls: Record<string, string> = {
    default: "text-[var(--text-primary)]",
    success: "text-[var(--success)]",
    warning: "text-[var(--warning)]",
    danger: "text-[var(--danger)]",
  };
  return (
    <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] p-3">
      <div className={`flex items-center gap-2 ${cls[color]}`}>{icon}<span className="text-lg font-bold">{value}</span></div>
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mt-1">{label}</p>
    </div>
  );
}
