"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getRun, subscribeToRun, type TestRun } from "@/lib/api";
import { StatusBadge, PriorityBadge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Section, LoadingSkeleton } from "@/components/ui/section";
import {
  FileCode,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Image,
  FileText,
  Clock,
  ArrowLeft,
} from "lucide-react";
import Link from "next/link";

export default function RunDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const [run, setRun] = useState<TestRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [liveState, setLiveState] = useState<string | null>(null);

  useEffect(() => {
    let unsub: (() => void) | undefined;
    getRun(id)
      .then((r) => {
        setRun(r);
        if (!["done", "failed"].includes(r.state)) {
          unsub = subscribeToRun(id, (event) => {
            if (event.type === "state_change") setLiveState(event.data.state);
            if (event.type === "done") getRun(id).then(setRun);
          });
        }
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
    return () => unsub?.();
  }, [id]);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-6 w-32 rounded bg-[var(--bg-subtle)] animate-pulse" />
        <LoadingSkeleton rows={6} />
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5">
        <p className="text-sm font-medium text-[var(--danger)]">
          {error || "Run not found"}
        </p>
      </div>
    );
  }

  const displayState = liveState || run.state;
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  return (
    <div className="space-y-6">
      {/* Back + Header */}
      <div className="flex items-start justify-between">
        <div>
          <Link href="/runs" className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-secondary)] mb-2">
            <ArrowLeft className="w-3 h-3" /> Back to runs
          </Link>
          <h1 className="text-lg font-bold text-[var(--text-primary)]">
            Run <span className="font-mono text-[var(--accent)]">{run.id.slice(0, 8)}</span>
          </h1>
          <p className="text-xs font-mono text-[var(--text-muted)] mt-0.5">{run.id}</p>
        </div>
        <div className="flex items-center gap-3">
          <StatusBadge state={displayState} />
          <a
            href={`${API}/api/v1/runs/${run.id}/report`}
            target="_blank"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:border-[var(--accent)]/30 transition-colors"
          >
            <FileText className="w-3.5 h-3.5" /> Report
          </a>
        </div>
      </div>

      {/* Meta Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <MetaCard label="Mode" value={run.mode || "simple"} />
        <MetaCard label="Fix Attempts" value={String(run.fix_attempts)} />
        <MetaCard label="Created" value={new Date(run.created_at).toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })} />
        <MetaCard label="Finished" value={run.finished_at ? new Date(run.finished_at).toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }) : "—"} />
      </div>

      {/* Results */}
      {run.run_result && (
        <Section title="Results">
          <div className="flex items-center gap-8 mb-4">
            <ResultStat icon={<CheckCircle2 className="w-4 h-4" />} value={run.run_result.passed} label="Passed" color="success" />
            <ResultStat icon={<XCircle className="w-4 h-4" />} value={run.run_result.failed} label="Failed" color="danger" />
            <ResultStat icon={<Clock className="w-4 h-4" />} value={run.run_result.total} label="Total" color="default" />
          </div>

          {run.run_result.failures.length > 0 && (
            <div className="space-y-2 mt-4">
              {run.run_result.failures.map((f, i) => (
                <div key={i} className="flex items-start gap-3 p-3 rounded-lg bg-[var(--danger-bg)] border border-[var(--danger)]/10">
                  <AlertTriangle className="w-4 h-4 text-[var(--danger)] mt-0.5 shrink-0" />
                  <div className="min-w-0">
                    <p className="text-xs font-medium text-[var(--text-primary)]">{f.test}</p>
                    <p className="text-[11px] text-[var(--danger)]/80 mt-0.5 break-all">{f.message}</p>
                    {f.screenshot_url && (
                      <a href={f.screenshot_url} target="_blank" className="inline-flex items-center gap-1 text-[10px] text-[var(--accent)] mt-1 hover:underline">
                        <Image className="w-3 h-3" /> Screenshot
                      </a>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </Section>
      )}

      {/* Test Plan */}
      {run.test_plan && (
        <Section title="Test Plan">
          <p className="text-sm text-[var(--text-secondary)] mb-4">{run.test_plan.summary}</p>
          <div className="space-y-2">
            {run.test_plan.scenarios.map((s, i) => (
              <div key={i} className="p-3 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-medium text-[var(--text-primary)]">{s.name}</span>
                  <PriorityBadge priority={s.priority} />
                </div>
                <ul className="text-[11px] text-[var(--text-muted)] space-y-0.5 ml-3">
                  {s.steps.map((step, j) => (
                    <li key={j} className="list-disc">{step}</li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </Section>
      )}

      {/* Generated Files */}
      {run.test_files && run.test_files.length > 0 && (
        <Section title="Generated Test Files">
          <div className="space-y-2">
            {run.test_files.map((f, i) => (
              <details key={i} className="group rounded-lg border border-[var(--border)] overflow-hidden">
                <summary className="flex items-center gap-2 px-4 py-2.5 cursor-pointer text-xs font-mono text-[var(--text-secondary)] hover:bg-[var(--bg-subtle)] transition-colors">
                  <FileCode className="w-3.5 h-3.5 text-[var(--text-muted)]" />
                  {f.name}
                </summary>
                <pre className="px-4 py-3 text-[11px] leading-relaxed overflow-x-auto bg-[var(--bg-subtle)] border-t border-[var(--border)] text-[var(--text-secondary)]">
                  {f.content}
                </pre>
              </details>
            ))}
          </div>
        </Section>
      )}

      {/* Screenshots */}
      {run.screenshots && run.screenshots.length > 0 && (
        <Section title="Screenshots">
          <div className="grid grid-cols-2 gap-2">
            {run.screenshots.map((url, i) => (
              <a key={i} href={url} target="_blank" className="flex items-center gap-2 p-3 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] hover:border-[var(--accent)]/30 transition-colors">
                <Image className="w-4 h-4 text-[var(--text-muted)]" />
                <span className="text-xs text-[var(--accent)] truncate">{url.split("/").pop()}</span>
              </a>
            ))}
          </div>
        </Section>
      )}

      {/* Error */}
      {run.error && (
        <Section title="Error">
          <pre className="text-xs text-[var(--danger)] bg-[var(--danger-bg)] p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">
            {run.error}
          </pre>
        </Section>
      )}
    </div>
  );
}

function MetaCard({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <p className="text-[10px] font-medium uppercase tracking-wider text-[var(--text-muted)]">{label}</p>
      <p className="text-sm font-semibold text-[var(--text-primary)] mt-0.5">{value}</p>
    </Card>
  );
}

function ResultStat({ icon, value, label, color }: { icon: React.ReactNode; value: number; label: string; color: string }) {
  const colors: Record<string, string> = {
    success: "text-[var(--success)]",
    danger: "text-[var(--danger)]",
    default: "text-[var(--text-primary)]",
  };
  return (
    <div className="flex items-center gap-2">
      <div className={colors[color]}>{icon}</div>
      <div>
        <p className={`text-lg font-bold ${colors[color]}`}>{value}</p>
        <p className="text-[10px] text-[var(--text-muted)]">{label}</p>
      </div>
    </div>
  );
}
