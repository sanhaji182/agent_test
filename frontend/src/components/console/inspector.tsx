"use client";

import { type TestRun } from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { ExecutionTimeline } from "./timeline";
import { X, ArrowRight, CheckCircle2, XCircle, AlertTriangle } from "lucide-react";
import Link from "next/link";

// Drawer inspector kanan untuk preview run yang dipilih di suite browser
export function RunInspector({
  run,
  onClose,
}: {
  run: TestRun | null;
  onClose: () => void;
}) {
  if (!run) return null;

  return (
    <div className="fixed inset-y-0 right-0 z-40 w-full max-w-md bg-[var(--bg-card)] border-l border-[var(--border)] shadow-xl flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-5 h-14 border-b border-[var(--border)]">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">Run</span>
          <span className="font-mono text-xs text-[var(--accent)]">{run.id.slice(0, 8)}</span>
        </div>
        <button onClick={onClose} className="p-1.5 rounded-md hover:bg-[var(--bg-hover)] text-[var(--text-muted)]">
          <X className="w-4 h-4" />
        </button>
      </div>

      <div className="flex-1 overflow-auto p-5 space-y-5">
        <div className="flex items-center justify-between">
          <StatusBadge state={run.state} />
          <span className="text-xs text-[var(--text-muted)]">
            {new Date(run.created_at).toLocaleString()}
          </span>
        </div>

        {/* Timeline */}
        <div className="p-4 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
          <ExecutionTimeline state={run.state} />
        </div>

        {/* Requirements */}
        {run.requirements && (
          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">Requirements</p>
            <p className="text-sm text-[var(--text-secondary)]">{run.requirements}</p>
          </div>
        )}

        {/* Result */}
        {run.run_result && (
          <div className="grid grid-cols-3 gap-2">
            <MiniStat icon={<CheckCircle2 className="w-3.5 h-3.5" />} value={run.run_result.passed} label="Passed" color="success" />
            <MiniStat icon={<XCircle className="w-3.5 h-3.5" />} value={run.run_result.failed} label="Failed" color="danger" />
            <MiniStat icon={<AlertTriangle className="w-3.5 h-3.5" />} value={run.fix_attempts} label="Fixes" color="warning" />
          </div>
        )}

        {/* Failures preview */}
        {run.run_result?.failures && run.run_result.failures.length > 0 && (
          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">Failures</p>
            <div className="space-y-1.5">
              {run.run_result.failures.slice(0, 3).map((f, i) => (
                <div key={i} className="p-2.5 rounded-md bg-[var(--danger-bg)] border border-[var(--danger)]/10">
                  <p className="text-xs font-medium text-[var(--text-primary)] truncate">{f.test}</p>
                  <p className="text-[11px] text-[var(--danger)]/80 truncate">{f.message}</p>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-[var(--border)]">
        <Link
          href={`/runs/${run.id}`}
          className="flex items-center justify-center gap-1.5 w-full py-2.5 rounded-lg bg-[var(--accent)] text-white text-sm font-medium hover:bg-[var(--accent-hover)] transition-colors"
        >
          Open Console <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}

function MiniStat({ icon, value, label, color }: { icon: React.ReactNode; value: number; label: string; color: string }) {
  const c: Record<string, string> = {
    success: "text-[var(--success)]",
    danger: "text-[var(--danger)]",
    warning: "text-[var(--warning)]",
  };
  return (
    <div className="p-2.5 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-center">
      <div className={`flex items-center justify-center mb-1 ${c[color]}`}>{icon}</div>
      <p className={`text-base font-bold ${c[color]}`}>{value}</p>
      <p className="text-[10px] text-[var(--text-muted)]">{label}</p>
    </div>
  );
}
