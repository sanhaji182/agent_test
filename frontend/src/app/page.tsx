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
  Flame, ArrowRight, Lightbulb, Shield,
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
  }, []);

  if (loading) return <LoadingSkeleton rows={6} />;

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const activeRuns = runs.filter((r) => isActive(r.state));
  const passRate = total > 0 ? Math.round((passed / total) * 100) : 0;

  // Show onboarding if no data
  if (total === 0) return <OnboardingView />;

  return (
    <div className="space-y-5">
      {/* Page header */}
      <div>
        <h1 className="text-lg font-bold">Overview</h1>
        <p className="text-[13px] text-[var(--text-secondary)]">Your testing health at a glance. Start here to understand what needs attention.</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="Total Runs" value={total} icon={<Activity className="w-4 h-4" />} />
        <StatCard label="Pass Rate" value={`${passRate}%`} color="success" icon={<CheckCircle2 className="w-4 h-4" />} trend={`${passed} passed`} />
        <StatCard label="Failed" value={failed} color="danger" icon={<XCircle className="w-4 h-4" />} />
        <StatCard label="Active" value={activeRuns.length} color="warning" icon={<PlayCircle className="w-4 h-4" />} />
      </div>

      {/* Trend + Recommendations side by side */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Trend */}
        <Section title="Pass Rate Trend">
          {runs.length > 1 ? (
            <PassRateChart trend={buildTrend(runs)} />
          ) : (
            <p className="text-[11px] text-[var(--text-muted)]">More runs needed to show trend.</p>
          )}
        </Section>

        {/* Recommendations */}
        <Section title="Recommendations" action={<Link href="/risk" className="text-[10px] font-medium text-[var(--accent)]">View all →</Link>}>
          {recs.length === 0 ? (
            <p className="text-[11px] text-[var(--text-muted)]">No recommendations yet. Run more tests to generate insights.</p>
          ) : (
            <div className="space-y-1.5">
              {recs.slice(0, 4).map((rec, i) => (
                <div key={i} className="flex items-center gap-2 px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)]">
                  <Lightbulb className="w-3.5 h-3.5 text-[var(--accent)] shrink-0" />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate flex-1">{rec.target}: {rec.reason}</span>
                  <span className="text-[9px] font-bold text-[var(--accent)]">{rec.action.replace("_", " ")}</span>
                </div>
              ))}
            </div>
          )}
        </Section>
      </div>

      {/* Active execution */}
      {activeRuns.length > 0 && (
        <Section title="Running Now">
          <div className="space-y-2">
            {activeRuns.map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="block p-3 rounded-[var(--radius-sm)] border border-[var(--border)] hover:border-[var(--accent-light)] transition-all">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <StatusBadge state={r.state} />
                    <span className="font-mono text-[11px] text-[var(--text-muted)]">{r.id.slice(0, 8)}</span>
                  </div>
                  <ArrowRight className="w-3.5 h-3.5 text-[var(--text-muted)]" />
                </div>
                <ExecutionTimeline state={r.state} />
              </Link>
            ))}
          </div>
        </Section>
      )}

      {/* Recent runs + Risk */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Section title="Recent Runs" action={<Link href="/runs" className="text-[10px] font-medium text-[var(--accent)]">All runs →</Link>}>
          <div className="space-y-0.5">
            {runs.slice(0, 5).map((r) => (
              <Link key={r.id} href={`/runs/${r.id}`} className="flex items-center justify-between px-2.5 py-2 rounded-[var(--radius-sm)] hover:bg-[var(--bg-hover)] transition-colors">
                <div className="flex items-center gap-2 min-w-0">
                  <StatusBadge state={r.state} />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate">{r.requirements || r.id.slice(0, 8)}</span>
                </div>
                {r.run_result && (
                  <span className="text-[10px] shrink-0 ml-2">
                    <span className="text-[var(--success)]">{r.run_result.passed}</span>
                    <span className="text-[var(--text-muted)]">/{r.run_result.total}</span>
                  </span>
                )}
              </Link>
            ))}
          </div>
        </Section>

        <Section title="Top Risks" action={<Link href="/risk" className="text-[10px] font-medium text-[var(--accent)]">Details →</Link>}>
          {risks.length === 0 ? (
            <p className="text-[11px] text-[var(--text-muted)]">No risks detected. This is good!</p>
          ) : (
            <div className="space-y-1.5">
              {risks.slice(0, 5).map((r, i) => (
                <div key={i} className="flex items-center gap-2 px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)]">
                  <Shield className="w-3.5 h-3.5 text-[var(--danger)] shrink-0" />
                  <span className="text-[11px] text-[var(--text-secondary)] truncate flex-1">{r.name}</span>
                  <span className="text-[9px] font-bold text-[var(--danger)]">{Math.round(r.risk_score * 100)}%</span>
                </div>
              ))}
            </div>
          )}
        </Section>
      </div>
    </div>
  );
}

// Onboarding view for first-time users
function OnboardingView() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Welcome to GoTest Agent</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5 max-w-lg">
          An AI-powered testing platform that analyzes your code, generates tests, runs them automatically, and helps you understand what needs attention.
        </p>
      </div>

      <Section title="Get Started in 3 Steps">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <StepCard num={1} title="Seed demo data" desc="See the platform in action with sample runs, schedules, and risk scores." />
          <StepCard num={2} title="Connect your project" desc="Point the agent at your codebase via API or MCP in your IDE." />
          <StepCard num={3} title="Set up monitoring" desc="Create schedules, enable alerts, and track release confidence." />
        </div>
        <div className="mt-4">
          <button
            onClick={() => { fetch((process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080") + "/api/v1/demo/seed", { method: "POST" }).then(() => window.location.reload()); }}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
          >
            <PlayCircle className="w-3.5 h-3.5" /> Seed Demo Data
          </button>
          <span className="ml-3 text-[11px] text-[var(--text-muted)]">Creates 5 sample runs, a schedule, and a release</span>
        </div>
      </Section>

      <Section title="What Each Page Does">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <PageDesc title="Runs" desc="View and inspect test executions with live timelines" />
          <PageDesc title="Risk" desc="See which tests and schedules need attention" />
          <PageDesc title="Monitoring" desc="Manage recurring schedules and their status" />
          <PageDesc title="Releases" desc="Track release readiness with confidence grades" />
          <PageDesc title="Suites" desc="Organize tests with tags and pinning" />
          <PageDesc title="Reviews" desc="Approve or reject generated test plans and fixes" />
          <PageDesc title="Alerts" desc="View notification history for failures" />
          <PageDesc title="Exports" desc="Download test data as JSON for reporting" />
        </div>
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

function PageDesc({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="flex items-start gap-2 px-2.5 py-2">
      <span className="text-[11px] font-semibold text-[var(--text-primary)] w-20 shrink-0">{title}</span>
      <span className="text-[11px] text-[var(--text-muted)]">{desc}</span>
    </div>
  );
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
