"use client";

import { useEffect, useState } from "react";
import { getRuns, isActive, getMetricsRisk, getRecommendations, type TestRun, type RiskItem, type Recommendation } from "@/lib/api";
import { StatCard } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, Section, LoadingSkeleton } from "@/components/ui/section";
import { ExecutionTimeline } from "@/components/console/timeline";
import { PassRateChart } from "@/components/ui/chart";
import Link from "next/link";
import {
  Activity, CheckCircle2, XCircle, PlayCircle,
  ArrowRight, Lightbulb, Shield, AlertTriangle, Clock, Film,
} from "lucide-react";

export default function OverviewPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [risks, setRisks] = useState<RiskItem[]>([]);
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getRuns(), getMetricsRisk().catch(() => []), getRecommendations().catch(() => [])])
      .then(([r, ri, re]) => { setRuns(r); setRisks(ri); setRecs(re); })
      .catch(() => {})
      .finally(() => setLoading(false));
    // Auto-refresh every 10s for live feel
    const interval = setInterval(() => { getRuns().then(setRuns).catch(() => {}); }, 10000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <LoadingSkeleton rows={8} />;

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const activeRuns = runs.filter((r) => isActive(r.state));
  const failedRuns = runs.filter((r) => r.state === "failed").slice(0, 3);
  const passRate = total > 0 ? Math.round((passed / total) * 100) : 0;

  if (total === 0) return <OnboardingView />;

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold">Control Room</h1>
          <p className="text-[12px] text-[var(--text-secondary)]">Live test execution status and audit overview</p>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-[var(--success)] animate-pulse" />
          <span className="text-[10px] text-[var(--text-muted)]">Live</span>
        </div>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="Total" value={total} icon={<Activity className="w-4 h-4" />} />
        <StatCard label="Pass Rate" value={`${passRate}%`} color="success" icon={<CheckCircle2 className="w-4 h-4" />} />
        <StatCard label="Failed" value={failed} color="danger" icon={<XCircle className="w-4 h-4" />} />
        <StatCard label="Running" value={activeRuns.length} color="warning" icon={<PlayCircle className="w-4 h-4" />} />
      </div>

      {/* === SECTION 1: What is running now? === */}
      {activeRuns.length > 0 && (
        <Section title="🔴 Running Now" action={<span className="text-[10px] text-[var(--warning)] font-semibold">{activeRuns.length} active</span>}>
          <div className="space-y-2">
            {activeRuns.map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="block p-3 rounded-[var(--radius-sm)] border border-[var(--warning)]/20 bg-[var(--warning-bg)] hover:border-[var(--warning)]/40 transition-all">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <StatusBadge state={r.state} />
                    <span className="text-[11px] font-medium text-[var(--text-primary)]">{r.requirements || r.id.slice(0, 8)}</span>
                  </div>
                  <span className="text-[10px] text-[var(--text-muted)]">{timeAgo(r.created_at)}</span>
                </div>
                <ExecutionTimeline state={r.state} />
              </Link>
            ))}
          </div>
        </Section>
      )}

      {/* === SECTION 2: What failed and why? === */}
      {failedRuns.length > 0 && (
        <Section title="⚠️ Failed — Needs Attention" action={<Link href="/runs" className="text-[10px] font-medium text-[var(--accent)]">All runs →</Link>}>
          <div className="space-y-2">
            {failedRuns.map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="block p-3 rounded-[var(--radius-sm)] border border-[var(--danger)]/15 bg-[var(--danger-bg)] hover:border-[var(--danger)]/30 transition-all">
                <div className="flex items-center justify-between mb-1.5">
                  <div className="flex items-center gap-2">
                    <AlertTriangle className="w-3.5 h-3.5 text-[var(--danger)]" />
                    <span className="text-[11px] font-semibold text-[var(--text-primary)]">{r.requirements || r.id.slice(0, 8)}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {r.video_url && <Film className="w-3 h-3 text-[var(--text-muted)]" title="Video available" />}
                    <span className="text-[10px] text-[var(--text-muted)]">{timeAgo(r.created_at)}</span>
                  </div>
                </div>
                {/* Failure reason */}
                {r.run_result?.failures?.[0] && (
                  <div className="mt-1 pl-5">
                    <p className="text-[11px] text-[var(--danger)]/80 truncate">{r.run_result.failures[0].test}: {r.run_result.failures[0].message?.slice(0, 100)}</p>
                  </div>
                )}
                {/* Result summary */}
                {r.run_result && (
                  <div className="mt-2 pl-5 flex items-center gap-3 text-[10px]">
                    <span className="text-[var(--success)]">{r.run_result.passed} passed</span>
                    <span className="text-[var(--danger)]">{r.run_result.failed} failed</span>
                    <span className="text-[var(--text-muted)]">{r.fix_attempts} fix attempts</span>
                  </div>
                )}
              </Link>
            ))}
          </div>
        </Section>
      )}

      {/* === SECTION 3: What should I do next? === */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Recommendations */}
        <Section title="💡 Recommended Actions" action={<Link href="/risk" className="text-[10px] font-medium text-[var(--accent)]">Risk →</Link>}>
          {recs.length === 0 ? (
            <p className="text-[11px] text-[var(--text-muted)]">No actions needed. Your tests are healthy.</p>
          ) : (
            <div className="space-y-1.5">
              {recs.slice(0, 5).map((rec, i) => (
                <div key={i} className="flex items-center gap-2 px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] hover:bg-[var(--bg-hover)] transition-colors">
                  <ActionDot action={rec.action} />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate flex-1">{rec.target}</span>
                  <span className="text-[9px] font-bold text-[var(--accent)] uppercase">{rec.action.replace("_", " ")}</span>
                </div>
              ))}
            </div>
          )}
        </Section>

        {/* Top risks */}
        <Section title="🛡️ Highest Risk" action={<Link href="/risk" className="text-[10px] font-medium text-[var(--accent)]">Details →</Link>}>
          {risks.length === 0 ? (
            <p className="text-[11px] text-[var(--text-muted)]">No elevated risks detected.</p>
          ) : (
            <div className="space-y-1.5">
              {risks.slice(0, 5).map((r, i) => (
                <div key={i} className="flex items-center gap-2 px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)]">
                  <RiskDot score={r.risk_score} />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate flex-1">{r.name}</span>
                  <span className="text-[9px] font-bold text-[var(--text-muted)]">{Math.round(r.risk_score * 100)}%</span>
                </div>
              ))}
            </div>
          )}
        </Section>
      </div>

      {/* Trend + Recent */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Section title="Trend">
          {runs.length > 1 ? <PassRateChart trend={buildTrend(runs)} /> : <p className="text-[11px] text-[var(--text-muted)]">More data needed.</p>}
        </Section>

        <Section title="Recent Completed" action={<Link href="/runs" className="text-[10px] font-medium text-[var(--accent)]">All →</Link>}>
          <div className="space-y-0.5">
            {runs.filter((r) => r.state === "done" || r.state === "failed").slice(0, 5).map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="flex items-center justify-between px-2.5 py-2 rounded-[var(--radius-sm)] hover:bg-[var(--bg-hover)] transition-colors">
                <div className="flex items-center gap-2 min-w-0">
                  <StatusBadge state={r.state} />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate">{r.requirements || r.id.slice(0, 8)}</span>
                </div>
                <span className="text-[10px] text-[var(--text-muted)] shrink-0">{timeAgo(r.created_at)}</span>
              </Link>
            ))}
          </div>
        </Section>
      </div>
    </div>
  );
}

// --- Helpers ---

function ActionDot({ action }: { action: string }) {
  const c = action === "run_now" ? "bg-[var(--accent)]" : action === "investigate" ? "bg-[var(--danger)]" : "bg-[var(--warning)]";
  return <span className={`w-2 h-2 rounded-full ${c} shrink-0`} />;
}

function RiskDot({ score }: { score: number }) {
  const c = score >= 0.7 ? "bg-[var(--danger)]" : score >= 0.4 ? "bg-[var(--warning)]" : "bg-[var(--success)]";
  return <span className={`w-2 h-2 rounded-full ${c} shrink-0`} />;
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function buildTrend(runs: TestRun[]): { date: string; pass_rate: number }[] {
  const byDate = new Map<string, { passed: number; total: number }>();
  for (const r of runs) {
    const d = r.created_at.slice(0, 10);
    const entry = byDate.get(d) || { passed: 0, total: 0 };
    entry.total++;
    if (r.state === "done") entry.passed++;
    byDate.set(d, entry);
  }
  return [...byDate.entries()]
    .map(([date, { passed, total }]) => ({ date, pass_rate: total > 0 ? passed / total : 0 }))
    .sort((a, b) => a.date.localeCompare(b.date));
}

// --- Onboarding ---

function OnboardingView() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Welcome to GoTest Agent</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5 max-w-lg">
          An AI-powered testing platform that analyzes your code, generates tests, runs them automatically, and helps you understand what needs attention.
        </p>
      </div>
      <Section title="Get Started">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-4">
          <StepCard num={1} title="Seed demo data" desc="See the platform in action with sample runs and risk scores." />
          <StepCard num={2} title="Connect your project" desc="Point the agent at your codebase via API or MCP." />
          <StepCard num={3} title="Set up monitoring" desc="Create schedules and enable failure alerts." />
        </div>
        <button
          onClick={() => { fetch((process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080") + "/api/v1/demo/seed", { method: "POST" }).then(() => window.location.reload()); }}
          className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
        >
          <PlayCircle className="w-3.5 h-3.5" /> Seed Demo Data
        </button>
      </Section>
    </div>
  );
}

function StepCard({ num, title, desc }: { num: number; title: string; desc: string }) {
  return (
    <div className="p-3 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)]">
      <div className="w-5 h-5 rounded-full bg-[var(--accent-bg)] text-[var(--accent)] flex items-center justify-center text-[10px] font-bold mb-2">{num}</div>
      <p className="text-[12px] font-semibold text-[var(--text-primary)]">{title}</p>
      <p className="text-[11px] text-[var(--text-muted)] mt-0.5 leading-relaxed">{desc}</p>
    </div>
  );
}
