"use client";

import { useEffect, useState } from "react";
import { getMetricsRisk, getRecommendations, type RiskItem, type Recommendation } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Shield, Lightbulb, AlertTriangle, PlayCircle, Search } from "lucide-react";

export default function RiskPage() {
  const [risks, setRisks] = useState<RiskItem[]>([]);
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getMetricsRisk(), getRecommendations()])
      .then(([r, rec]) => { setRisks(r); setRecs(rec); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Risk</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Tests and schedules ranked by failure risk. Act on the highest-priority items first.</p>
      </div>

      {/* Recommendations */}
      <Section title="Recommendations" action={<span className="text-[11px] text-[var(--text-muted)]">{recs.length} actions</span>}>
        {recs.length === 0 ? (
          <EmptyState icon={<Lightbulb className="w-6 h-6" />} title="No recommendations" description="Recommendations appear when risk patterns are detected across your test runs." />
        ) : (
          <div className="space-y-2">
            {recs.map((rec, i) => (
              <div key={i} className="flex items-start gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                <ActionIcon action={rec.action} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-semibold text-[var(--text-primary)]">{rec.target}</span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-[var(--accent-bg)] text-[var(--accent)]">{rec.action.replace("_", " ")}</span>
                  </div>
                  <p className="text-[11px] text-[var(--text-muted)] mt-0.5">{rec.reason}</p>
                </div>
                <span className="text-xs font-bold text-[var(--warning)]">{(rec.priority * 100).toFixed(0)}%</span>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* Risk Items */}
      <Section title="Risk Scores" action={<span className="text-[11px] text-[var(--text-muted)]">{risks.length} items</span>}>
        {risks.length === 0 ? (
          <EmptyState icon={<Shield className="w-6 h-6" />} title="No risk data" description="Risk scores are computed from test failure patterns, schedule staleness, and environment criticality." />
        ) : (
          <div className="space-y-1.5">
            {risks.map((risk, i) => (
              <div key={i} className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-[var(--bg-subtle)] transition-colors">
                <RiskBar score={risk.risk_score} />
                <div className="flex-1 min-w-0">
                  <span className="text-xs font-medium text-[var(--text-primary)]">{risk.name}</span>
                  <span className="ml-2 text-[10px] text-[var(--text-muted)]">{risk.type}</span>
                </div>
                <span className="text-[11px] text-[var(--text-secondary)]">{risk.reason}</span>
                {risk.environment && (
                  <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-[var(--info-bg)] text-[var(--info)]">{risk.environment}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}

function RiskBar({ score }: { score: number }) {
  const pct = Math.round(score * 100);
  const color = score >= 0.7 ? "bg-[var(--danger)]" : score >= 0.4 ? "bg-[var(--warning)]" : "bg-[var(--success)]";
  return (
    <div className="w-16 flex items-center gap-1.5">
      <div className="flex-1 h-1.5 rounded-full bg-[var(--border)]">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-[10px] font-bold text-[var(--text-secondary)] w-7">{pct}%</span>
    </div>
  );
}

function ActionIcon({ action }: { action: string }) {
  switch (action) {
    case "run_now": return <PlayCircle className="w-4 h-4 text-[var(--accent)] mt-0.5 shrink-0" />;
    case "investigate": return <Search className="w-4 h-4 text-[var(--danger)] mt-0.5 shrink-0" />;
    default: return <AlertTriangle className="w-4 h-4 text-[var(--warning)] mt-0.5 shrink-0" />;
  }
}
