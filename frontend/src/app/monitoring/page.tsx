"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  getMonitoringSummary,
  getMetricsSummary,
  getMetricsTrend,
  getMetricsRisk,
  getMetricsFlaky,
  isExecuting,
  type MonitoringSummary,
  type MetricsSummary,
  type TrendPoint,
  type RiskItem,
  type FlakyTest,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { CheckCircle2, AlertTriangle, XCircle, Clock, RefreshCw, Activity } from "lucide-react";

export default function MonitoringPage() {
  const [monitoring, setMonitoring] = useState<MonitoringSummary | null>(null);
  const [metrics, setMetrics] = useState<MetricsSummary | null>(null);
  const [trend, setTrend] = useState<TrendPoint[]>([]);
  const [risks, setRisks] = useState<RiskItem[]>([]);
  const [flaky, setFlaky] = useState<FlakyTest[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const [mon, sum, tr, rk, fl] = await Promise.all([
        getMonitoringSummary(),
        getMetricsSummary(),
        getMetricsTrend(),
        getMetricsRisk(),
        getMetricsFlaky(),
      ]);
      setMonitoring(mon);
      setMetrics(sum);
      setTrend(tr || []);
      setRisks(rk || []);
      setFlaky(fl || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    load().finally(() => setLoading(false));
  }, [load]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await load();
    setRefreshing(false);
  };

  if (loading) return <LoadingSkeleton rows={6} />;

  const summary = monitoring?.summary;
  const passRate = metrics ? `${Math.round(metrics.pass_rate)}%` : "-";
  const avgDuration = metrics ? formatDuration(metrics.avg_duration_ms) : "-";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Monitoring</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">System health & performance metrics</p>
        </div>
        <Button variant="secondary" onClick={handleRefresh} disabled={refreshing}>
          <RefreshCw className={`w-4 h-4 mr-2 ${refreshing ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Test Tingkat Berhasil" value={passRate} positive={metrics != null && metrics.pass_rate >= 90} danger={metrics != null && metrics.pass_rate < 50} />
        <StatCard label="Avg Run Duration" value={avgDuration} color="blue" />
        <StatCard label="Active Runs" value={String(summary?.active_runs ?? 0)} color="blue" />
        <StatCard label="Failed Test" value={String(metrics?.total_failed ?? 0)} danger={(metrics?.total_failed ?? 0) > 0} />
      </div>

      {/* Secondary counts */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard label="Total Runs" value={String(metrics?.total_runs ?? 0)} />
        <StatCard label="Test Passed" value={String(metrics?.total_passed ?? 0)} positive={(metrics?.total_passed ?? 0) > 0} />
        <StatCard label="Completed Runs" value={String(summary?.completed_runs ?? 0)} />
      </div>

      {/* Tingkat Berhasil Trend */}
      <Section title="Tingkat Berhasil Trend">
        {trend.length === 0 ? (
          <EmptyState
            icon={<Activity className="w-8 h-8" />}
            title="No trend data yet"
            description="Run some tests to see the pass rate trend over time."
          />
        ) : (
          <div className="flex items-end gap-2 h-48 px-2">
            {trend.map((t) => (
              <div key={t.date} className="flex-1 flex flex-col items-center gap-1 min-w-0">
                <span className="text-xs text-[var(--text-muted)]">{Math.round(t.pass_rate)}%</span>
                <div
                  className={`w-full rounded-t ${t.pass_rate >= 90 ? "bg-green-500" : t.pass_rate >= 50 ? "bg-yellow-500" : "bg-red-500"}`}
                  style={{ height: `${Math.max(4, (t.pass_rate / 100) * 140)}px` }}
                  title={`${t.date}: ${Math.round(t.pass_rate)}% pass rate, ${t.fail_count} failed of ${t.total_tests} tests`}
                />
                <span className="text-[10px] text-[var(--text-muted)] whitespace-nowrap">{t.date.slice(5)}</span>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* Test Lists Health */}
      <Section title="Test Lists">
        {!monitoring || monitoring.lists.length === 0 ? (
          <EmptyState
            icon={<Activity className="w-8 h-8" />}
            title="No test lists"
            description="Create a test list to monitor its health here."
          />
        ) : (
          <TableContainer>
            <table className="w-full text-left">
              <thead className="bg-gray-50 border-b border-[var(--border-default)]">
                <tr>
                  <Th>List</Th>
                  <Th>Test</Th>
                  <Th>Tingkat Berhasil</Th>
                  <Th>Status Terakhir</Th>
                  <Th>Run Terakhir</Th>
                </tr>
              </thead>
              <tbody>
                {monitoring.lists.map((l) => (
                  <Tr key={l.id} hover>
                    <Td className="font-medium">{l.name}</Td>
                    <Td className="text-sm text-[var(--text-muted)]">{l.test_count}</Td>
                    <Td className="text-sm">
                      <span className={l.pass_rate >= 90 ? "text-green-600" : l.pass_rate >= 50 ? "text-yellow-600" : "text-red-600"}>
                        {Math.round(l.pass_rate)}%
                      </span>
                    </Td>
                    <Td className="text-sm text-[var(--text-muted)] capitalize">{l.last_status || "-"}</Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                      {l.last_run_at ? formatDate(l.last_run_at) : "Never"}
                    </Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </TableContainer>
        )}
      </Section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Runs */}
        <Section title="Recent Runs">
          {!monitoring || monitoring.recent_runs.length === 0 ? (
            <EmptyState
              icon={<Clock className="w-8 h-8" />}
              title="No runs yet"
              description="Recent test runs will appear here."
            />
          ) : (
            <div className="space-y-2">
              {monitoring.recent_runs.slice(0, 8).map((run) => (
                <RunActivityItem key={run.id} id={run.id} state={run.state} mode={run.mode} time={run.created_at} />
              ))}
            </div>
          )}
        </Section>

        {/* Risk & Flaky */}
        <div className="space-y-6">
          <Section title="At-Risk Test">
            {risks.length === 0 ? (
              <EmptyState
                icon={<CheckCircle2 className="w-8 h-8" />}
                title="No risks detected"
                description="No tests are currently flagged as at-risk."
              />
            ) : (
              <div className="space-y-2">
                {risks.slice(0, 5).map((r, i) => (
                  <ActivityItem
                    key={`${r.name}-${i}`}
                    icon={<AlertTriangle className="w-4 h-4 text-yellow-600" />}
                    title={r.name}
                    subtitle={`${r.reason} (risk ${Math.round(r.risk_score)})`}
                    warning
                  />
                ))}
              </div>
            )}
          </Section>

          <Section title="Flaky Test">
            {flaky.length === 0 ? (
              <EmptyState
                icon={<CheckCircle2 className="w-8 h-8" />}
                title="No flaky tests"
                description="No pass/fail flapping detected across recent runs."
              />
            ) : (
              <div className="space-y-2">
                {flaky.slice(0, 5).map((f) => (
                  <ActivityItem
                    key={f.test_name}
                    icon={<AlertTriangle className="w-4 h-4 text-yellow-600" />}
                    title={f.test_name}
                    subtitle={`${f.flip_count} pass/fail flips across ${f.total_appearances} appearances`}
                    warning
                  />
                ))}
              </div>
            )}
          </Section>
        </div>
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  positive,
  danger,
  color = "default"
}: {
  label: string;
  value: string;
  positive?: boolean;
  danger?: boolean;
  color?: string;
}) {
  const textColor = danger ? "text-red-600" : positive ? "text-green-600" : color === "blue" ? "text-blue-600" : "";

  return (
    <div className="bg-white rounded-lg p-4 border border-[var(--border-default)]">
      <p className="text-xs text-[var(--text-muted)] font-medium uppercase tracking-wide">{label}</p>
      <div className="mt-2 flex items-baseline gap-2">
        <span className={`text-2xl font-semibold ${textColor}`}>{value}</span>
      </div>
    </div>
  );
}

function RunActivityItem({ id, state, mode, time }: { id: string; state: string; mode: string; time: string }) {
  const executing = isExecuting(state);
  const icon = state === "done"
    ? <CheckCircle2 className="w-4 h-4 text-green-600" />
    : state === "failed"
      ? <XCircle className="w-4 h-4 text-red-600" />
      : executing
        ? <Clock className="w-4 h-4 text-blue-600" />
        : <AlertTriangle className="w-4 h-4 text-yellow-600" />;
  const variant = state === "done" ? undefined : state === "failed" ? "danger" : executing ? undefined : "warning";

  return (
    <Link href={`/runs/${id}`} className="block">
      <ActivityItem
        icon={icon}
        title={`Run ${id.slice(0, 8)} — ${mode || "test"}`}
        subtitle={`${state} · ${formatDate(time)}`}
        warning={variant === "warning"}
        danger={variant === "danger"}
      />
    </Link>
  );
}

function ActivityItem({
  icon,
  title,
  subtitle,
  warning,
  danger
}: {
  icon: React.ReactNode;
  title: string;
  subtitle?: string;
  warning?: boolean;
  danger?: boolean;
}) {
  const bgColor = warning ? "bg-yellow-50" : danger ? "bg-red-50" : "bg-gray-50";
  const borderColor = warning ? "border-yellow-200" : danger ? "border-red-200" : "border-gray-200";

  return (
    <div className={`flex items-start gap-3 p-3 rounded-lg border ${bgColor} ${borderColor}`}>
      <div className="shrink-0 mt-0.5">{icon}</div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-[var(--text-primary)] truncate">{title}</p>
        {subtitle && <p className="text-xs text-[var(--text-muted)] mt-0.5">{subtitle}</p>}
      </div>
    </div>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function formatDuration(ms: number): string {
  if (!ms) return "-";
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.round(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}
