"use client";

import { useEffect, useState } from "react";
import { getRuns, isExecuting, getMetricsRisk, getRecommendations, type TestRun, type RiskItem, type Recommendation } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import Link from "next/link";
import { 
  Activity, CheckCircle2, XCircle, PlayCircle,
  Lightbulb, AlertTriangle, Clock, ArrowUpRight,
  Shield
} from "lucide-react";

export default function DashboardPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [risks, setRisks] = useState<RiskItem[]>([]);
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    Promise.all([getRuns(), getMetricsRisk().catch(() => []), getRecommendations().catch(() => [])])
      .then(([r, ri, re]) => { 
        setRuns(r || []); 
        setRisks(ri || []); 
        setRecs(re || []); 
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    let es: EventSource | null = null;
    let fallbackInterval: NodeJS.Timeout | null = null;

    const connectSSE = () => {
      es = new EventSource(`${API}/api/v1/stream`);
      es.addEventListener("update", (e) => {
        const data = JSON.parse(e.data);
        if (["run_started", "run_completed", "run_failed", "analysis_completed", "plan_generated"].includes(data.type)) {
          getRuns().then((r) => setRuns(r || [])).catch(() => {});
        }
        if (data.failed) {
          const label = data.metadata?.test || data.message?.slice(0, 60) || data.run_id?.slice(0, 8);
          setToast(`Test failed: ${label}`);
          setTimeout(() => setToast(null), 5000);
        }
      });
      es.onopen = () => { setConnected(true); if (fallbackInterval) { clearInterval(fallbackInterval); fallbackInterval = null; } };
      es.onerror = () => {
        setConnected(false);
        es?.close();
        if (!fallbackInterval) {
          fallbackInterval = setInterval(() => { getRuns().then((r) => setRuns(r || [])).catch(() => {}); }, 10000);
        }
        setTimeout(connectSSE, 5000);
      };
    };

    connectSSE();
    return () => { es?.close(); if (fallbackInterval) clearInterval(fallbackInterval); };
  }, []);

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const activeCount = runs.filter((r) => isExecuting(r.state)).length;
  const passRate = total > 0 ? Math.round((passed / total) * 100) : 0;
  const recentRuns = runs.slice(0, 10);
  const activeRunsList = runs.filter((r) => isExecuting(r.state));
  
  const hasActive = activeCount > 0;
  const hasFailed = failed > 0;

  if (loading) {
    return <div className="space-y-6"><LoadingSkeleton rows={6} /></div>;
  }

  if (total === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
        <h1 className="text-2xl font-semibold tracking-tight mb-2">No tests yet</h1>
        <p className="text-sm text-[var(--text-muted)] max-w-sm leading-relaxed mb-6">
          Start testing by creating your first test run. The AI will analyze your codebase and generate comprehensive test coverage.
        </p>
        <Link href="/create">
          <Button>Create First Test</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Toast */}
      {toast && (
        <div className="fixed top-4 right-4 z-50 px-4 py-3 rounded-lg bg-red-50 border border-red-200 shadow-md animate-slide-in">
          <p className="text-sm font-medium text-red-700">{toast}</p>
        </div>
      )}

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1.5">
            {hasActive 
              ? `${activeCount} test${activeCount === 1 ? ' run' : 's'} currently executing`
              : 'Overview of your test execution history'
            }
          </p>
        </div>
        <Link href="/create">
          <Button>New Run</Button>
        </Link>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatBox value={total.toString()} label="Total Runs" />
        <StatBox value={`${passRate}%`} label="Pass Rate" success={passRate >= 80} />
        <StatBox value={failed.toString()} label="Failures" danger={failed > 0} />
        <StatBox value={activeCount.toString()} label="Running" color={activeCount > 0 ? "blue" : "gray"} />
      </div>

      {/* Active Runs Table */}
      {hasActive && (
        <Section title="Running Tests">
          <TableContainer>
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-[var(--border-default)]">
                  <Th>Run Name</Th>
                  <Th>Status</Th>
                  <Th>Progress</Th>
                  <Th>Started</Th>
                  <Th align="right">Actions</Th>
                </tr>
              </thead>
              <tbody>
                {activeRunsList.map((r) => (
                  <Tr key={r.id} onClick={() => window.location.href = `/runs/${r.id}`} hover>
                    <Td className="font-medium">
                      <span className="truncate block max-w-[200px]">{r.requirements || "Untitled"}</span>
                    </Td>
                    <Td>
                      <StatusBadge state={r.state} />
                    </Td>
                    <Td>
                      <ProgressBar state={r.state} />
                    </Td>
                    <Td className="text-[var(--text-muted)] text-xs">
                      {formatDate(r.created_at)}
                    </Td>
                    <Td align="right">
                      <Link href={`/runs/${r.id}`} className="text-blue-600 hover:text-blue-700 text-xs font-medium">View</Link>
                    </Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </TableContainer>
        </Section>
      )}

      {/* Failed Runs */}
      {hasFailed && (
        <Section title="Recent Failures" action={<Link href="/runs?filter=failed" className="text-xs text-blue-600 hover:text-blue-700">All failures →</Link>}>
          <div className="space-y-2">
            {runs.filter((r) => r.state === "failed").slice(0, 3).map((r) => (
              <FailedRunCard key={r.id} run={r} />
            ))}
          </div>
        </Section>
      )}

      {/* Two Columns */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Section title="Recommended Actions" action={<Link href="/risk" className="text-xs text-blue-600 hover:text-blue-700">See all →</Link>}>
          {recs.length === 0 ? (
            <p className="text-sm text-[var(--text-muted)]">No actions needed.</p>
          ) : (
            <div className="space-y-2">
              {recs.slice(0, 4).map((rec, i) => (
                <RecommendationCard key={i} rec={rec} />
              ))}
            </div>
          )}
        </Section>

        <Section title="Highest Risk Areas" action={<Link href="/risk" className="text-xs text-blue-600 hover:text-blue-700">Analyze →</Link>}>
          {risks.length === 0 ? (
            <p className="text-sm text-[var(--text-muted)]">No elevated risks detected.</p>
          ) : (
            <div className="space-y-2">
              {risks.slice(0, 4).map((risk, i) => (
                <RiskCard key={i} risk={risk} />
              ))}
            </div>
          )}
        </Section>
      </div>

      {/* Recent Runs Table */}
      <Section title="Recent Runs" action={<Link href="/runs" className="text-xs text-blue-600 hover:text-blue-700">View all →</Link>}>
        <TableContainer>
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[var(--border-default)]">
                <Th>Run Name</Th>
                <Th>Status</Th>
                <Th>Coverage</Th>
                <Th>Duration</Th>
                <Th>Started</Th>
              </tr>
            </thead>
            <tbody>
              {recentRuns.map((r) => (
                <Tr key={r.id} onClick={() => window.location.href = `/runs/${r.id}`} hover>
                  <Td className="font-medium">
                    <span className="truncate block max-w-[200px]">{r.requirements || "Untitled"}</span>
                  </Td>
                  <Td><StatusBadge state={r.state} /></Td>
                  <Td>
                    {r.run_result ? (
                      <span className="text-xs text-[var(--text-muted)]">
                        {r.run_result.passed}/{r.run_result.total}
                      </span>
                    ) : "-"}
                  </Td>
                  <Td className="text-xs text-[var(--text-muted)]">
                    {r.run_result?.duration_ms ? formatDurationMs(r.run_result.duration_ms) : "-"}
                  </Td>
                  <Td className="text-[var(--text-muted)] text-xs">
                    {formatDate(r.created_at)}
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      </Section>
    </div>
  );
}

// === Components ===

function StatBox({ value, label, success, danger, color = "default" }: { 
  value: string; 
  label: string; 
  success?: boolean;
  danger?: boolean;
  color?: string;
}) {
  const textColor = danger ? "text-red-600" : success ? "text-green-600" : color === "blue" ? "text-blue-600" : "";
  const bgColor = danger ? "bg-red-50" : success ? "bg-green-50" : color === "blue" ? "bg-blue-50" : "bg-gray-50";
  
  return (
    <div className={`bg-white rounded-lg p-4 border border-[var(--border-default)] ${bgColor}`}>
      <p className="text-xs text-[var(--text-muted)] font-medium uppercase tracking-wide">{label}</p>
      <p className={`mt-2 text-2xl font-semibold ${textColor}`}>{value}</p>
    </div>
  );
}

function StatusBadge({ state, size = "md" }: { state: string; size?: "sm" | "md" }) {
  const colors: Record<string, string> = {
    done: "bg-green-100 text-green-700 border-green-200",
    failed: "bg-red-100 text-red-700 border-red-200",
    running: "bg-blue-100 text-blue-700 border-blue-200",
    analyzing: "bg-yellow-100 text-yellow-700 border-yellow-200",
    writing_tests: "bg-purple-100 text-purple-700 border-purple-200",
    fixing: "bg-orange-100 text-orange-700 border-orange-200",
  };
  const color = colors[state] || "bg-gray-100 text-gray-700 border-gray-200";
  
  return (
    <span className={`inline-flex items-center font-medium rounded ${color} ${size === "sm" ? "px-1.5 py-0.5 text-xs" : "px-2 py-0.5 text-xs"}`}>
      {state.replace("_", " ").replace(/\b\w/g, l => l.toUpperCase())}
    </span>
  );
}

function ProgressBar({ state }: { state: string }) {
  const progressMap: Record<string, number> = {
    idle: 0,
    analyzing: 15,
    plan_generated: 25,
    writing_tests: 45,
    running: 70,
    fixing: 85,
    done: 100,
    failed: 100,
  };
  
  const progress = progressMap[state] || 0;
  
  return (
    <div className="w-full max-w-[100px]">
      <div className="flex items-center gap-2">
        <div className="flex-1 h-1.5 bg-gray-200 rounded-full overflow-hidden">
          <div 
            className="h-full bg-blue-600 rounded-full transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
        <span className="text-xs text-[var(--text-muted)] w-8 text-right">{progress}%</span>
      </div>
    </div>
  );
}

function FailedRunCard({ run }: { run: TestRun }) {
  const failure = run.run_result?.failures?.[0];
  
  return (
    <Link href={`/runs/${run.id}`} className="block p-3 rounded-lg border border-red-200 bg-red-50 hover:border-red-300 transition-colors">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <AlertTriangle className="w-4 h-4 text-red-600 shrink-0" />
            <span className="font-medium text-sm truncate">{run.requirements || "Untitled test"}</span>
          </div>
          {failure && (
            <p className="text-xs text-red-700 truncate ml-6">
              {failure.test}: {failure.message?.slice(0, 80)}...
            </p>
          )}
        </div>
        <span className="text-xs text-[var(--text-muted)] whitespace-nowrap ml-2">
          {formatDate(run.created_at)}
        </span>
      </div>
    </Link>
  );
}

function RecommendationCard({ rec }: { rec: Recommendation }) {
  return (
    <div className="flex items-start gap-3 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
      <Lightbulb className="w-4 h-4 text-yellow-600 shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        <p className="text-sm text-[var(--text-primary)]">{rec.target}</p>
        <p className="text-xs text-[var(--text-muted)] mt-0.5 capitalize">{rec.action.replace("_", " ")}</p>
      </div>
      <ArrowUpRight className="w-4 h-4 text-[var(--text-muted)] shrink-0" />
    </div>
  );
}

function RiskCard({ risk }: { risk: RiskItem }) {
  const score = risk.risk_score * 100;
  const isHigh = score > 70;
  
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg bg-gray-50">
      <Shield className={`w-4 h-4 shrink-0 ${isHigh ? "text-red-600" : "text-blue-600"}`} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium truncate">{risk.name}</p>
        <p className="text-xs text-[var(--text-muted)]">{risk.reason}</p>
      </div>
      <span className={`text-xs font-semibold ${isHigh ? "text-red-600" : "text-blue-600"}`}>
        {Math.round(score)}%
      </span>
    </div>
  );
}

// === Helpers ===

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);
  
  if (diffMin < 1) return "Just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

function formatDurationMs(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const mins = Math.floor(ms / 60000);
  const secs = ((ms / 1000) % 60).toFixed(0);
  return `${mins}m ${secs}s`;
}
