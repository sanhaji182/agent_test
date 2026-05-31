"use client";

import { useEffect, useState } from "react";
import { getRuns, type TestRun } from "@/lib/api";
import { StatCard } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, LoadingSkeleton, Section } from "@/components/ui/section";
import Link from "next/link";
import {
  Activity,
  CheckCircle2,
  XCircle,
  PlayCircle,
  Clock,
  Inbox,
} from "lucide-react";

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

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-24 rounded-xl bg-[var(--bg-subtle)] animate-pulse" />
          ))}
        </div>
        <LoadingSkeleton rows={5} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5">
        <p className="text-sm font-medium text-[var(--danger)]">Failed to load dashboard</p>
        <p className="text-xs text-[var(--danger)]/70 mt-1">{error}</p>
      </div>
    );
  }

  const total = runs.length;
  const passed = runs.filter((r) => r.state === "done").length;
  const failed = runs.filter((r) => r.state === "failed").length;
  const active = runs.filter((r) => !["done", "failed", "idle"].includes(r.state)).length;
  const passRate = total > 0 ? Math.round((passed / total) * 100) : 0;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-xl font-bold text-[var(--text-primary)]">Dashboard</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">
          Overview of your testing activity
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Total Runs"
          value={total}
          icon={<Activity className="w-4 h-4" />}
        />
        <StatCard
          label="Pass Rate"
          value={`${passRate}%`}
          color="success"
          icon={<CheckCircle2 className="w-4 h-4" />}
          trend={`${passed} passed`}
        />
        <StatCard
          label="Failed"
          value={failed}
          color="danger"
          icon={<XCircle className="w-4 h-4" />}
        />
        <StatCard
          label="Active"
          value={active}
          color="warning"
          icon={<PlayCircle className="w-4 h-4" />}
        />
      </div>

      {/* Recent Runs */}
      <Section
        title="Recent Activity"
        action={
          runs.length > 0 ? (
            <Link href="/runs" className="text-[11px] font-medium text-[var(--accent)] hover:text-[var(--accent-hover)]">
              View all →
            </Link>
          ) : null
        }
      >
        {runs.length === 0 ? (
          <EmptyState
            icon={<Inbox className="w-6 h-6" />}
            title="No test runs yet"
            description="Create your first run via the API or MCP tool in your IDE."
          />
        ) : (
          <div className="space-y-2">
            {runs.slice(0, 8).map((run) => (
              <Link
                key={run.id}
                href={`/runs/${run.id}`}
                className="flex items-center justify-between px-4 py-3 rounded-lg hover:bg-[var(--bg-subtle)] transition-colors group"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <StatusBadge state={run.state} />
                  <span className="text-xs font-mono text-[var(--text-secondary)] truncate group-hover:text-[var(--text-primary)]">
                    {run.id.slice(0, 8)}
                  </span>
                  {run.requirements && (
                    <span className="text-xs text-[var(--text-muted)] truncate max-w-[200px]">
                      {run.requirements}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-3 text-[var(--text-muted)]">
                  <Clock className="w-3 h-3" />
                  <span className="text-[11px]">
                    {new Date(run.created_at).toLocaleString(undefined, {
                      month: "short",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}
