"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getRuns, compareRuns, type TestRun, type CompareResult } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { StatusBadge } from "@/components/ui/badge";
import { ArrowLeft, GitCompare, TrendingUp, TrendingDown, Minus } from "lucide-react";
import Link from "next/link";

export default function ComparePage() {
  const params = useParams();
  const id = params.id as string;
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [result, setResult] = useState<CompareResult | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    getRuns().then((r) => setRuns(r.filter((run) => run.id !== id))).catch(() => {});
  }, [id]);

  const handleCompare = async (otherId: string) => {
    setSelectedId(otherId);
    setLoading(true);
    try {
      const res = await compareRuns(id, otherId);
      setResult(res);
    } catch {
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <Link href={`/runs/${id}`} className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-secondary)]">
        <ArrowLeft className="w-3 h-3" /> Back to run
      </Link>

      <div>
        <h1 className="text-lg font-bold">Compare Run <span className="font-mono text-[var(--accent)]">{id.slice(0, 8)}</span></h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Select another run to compare against</p>
      </div>

      {/* Run selector */}
      <Section title="Select Run to Compare">
        {runs.length === 0 ? (
          <EmptyState icon={<GitCompare className="w-6 h-6" />} title="No other runs" description="Create more runs to enable comparison." />
        ) : (
          <div className="space-y-1 max-h-48 overflow-y-auto">
            {runs.slice(0, 10).map((r) => (
              <button
                key={r.id}
                onClick={() => handleCompare(r.id)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors ${selectedId === r.id ? "bg-[var(--accent-bg)] border border-[var(--accent)]/20" : "hover:bg-[var(--bg-hover)]"}`}
              >
                <StatusBadge state={r.state} />
                <span className="font-mono text-xs text-[var(--text-secondary)]">{r.id.slice(0, 8)}</span>
                <span className="text-xs text-[var(--text-muted)] truncate flex-1">{r.requirements || "—"}</span>
              </button>
            ))}
          </div>
        )}
      </Section>

      {/* Results */}
      {loading && <LoadingSkeleton rows={4} />}
      {result && !loading && (
        <Section title="Comparison Result">
          <div className="space-y-4">
            <p className="text-sm font-medium text-[var(--text-primary)]">{result.summary}</p>

            <div className="grid grid-cols-3 gap-3">
              <DeltaCard label="Total" value={result.total_delta} />
              <DeltaCard label="Passed" value={result.passed_delta} />
              <DeltaCard label="Failed" value={result.failed_delta} invert />
            </div>

            {result.new_failures.length > 0 && (
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--danger)] mb-2">New Failures ({result.new_failures.length})</p>
                <div className="space-y-1">
                  {result.new_failures.map((t, i) => (
                    <div key={i} className="px-3 py-2 rounded-md bg-[var(--danger-bg)] text-xs text-[var(--danger)]">{t}</div>
                  ))}
                </div>
              </div>
            )}

            {result.recovered.length > 0 && (
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--success)] mb-2">Recovered ({result.recovered.length})</p>
                <div className="space-y-1">
                  {result.recovered.map((t, i) => (
                    <div key={i} className="px-3 py-2 rounded-md bg-[var(--success-bg)] text-xs text-[var(--success)]">{t}</div>
                  ))}
                </div>
              </div>
            )}

            {result.common_failures.length > 0 && (
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">Still Failing ({result.common_failures.length})</p>
                <div className="space-y-1">
                  {result.common_failures.map((t, i) => (
                    <div key={i} className="px-3 py-2 rounded-md bg-[var(--bg-subtle)] text-xs text-[var(--text-secondary)]">{t}</div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </Section>
      )}
    </div>
  );
}

function DeltaCard({ label, value, invert }: { label: string; value: number; invert?: boolean }) {
  const positive = invert ? value < 0 : value > 0;
  const negative = invert ? value > 0 : value < 0;
  return (
    <div className="p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)] text-center">
      <p className="text-[10px] font-semibold uppercase text-[var(--text-muted)]">{label}</p>
      <div className="flex items-center justify-center gap-1 mt-1">
        {positive && <TrendingUp className="w-3.5 h-3.5 text-[var(--success)]" />}
        {negative && <TrendingDown className="w-3.5 h-3.5 text-[var(--danger)]" />}
        {value === 0 && <Minus className="w-3.5 h-3.5 text-[var(--text-muted)]" />}
        <span className={`text-lg font-bold ${positive ? "text-[var(--success)]" : negative ? "text-[var(--danger)]" : "text-[var(--text-primary)]"}`}>
          {value > 0 ? "+" : ""}{value}
        </span>
      </div>
    </div>
  );
}
