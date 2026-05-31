"use client";

import { useEffect, useState } from "react";
import { getSuites, type Suite } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Layers, Pin } from "lucide-react";

export default function SuitesPage() {
  const [suites, setSuites] = useState<Suite[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getSuites().then(setSuites).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSkeleton rows={5} />;

  const pinned = suites.filter((s) => s.pinned);
  const others = suites.filter((s) => !s.pinned);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">Test Suites</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Organize and manage test groups with tags and pinning</p>
      </div>

      {suites.length === 0 ? (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
          <EmptyState icon={<Layers className="w-6 h-6" />} title="No suites configured" description="Create suites via POST /api/v1/suites to organize tests by project, environment, or custom tags." />
        </div>
      ) : (
        <>
          {pinned.length > 0 && (
            <Section title="Pinned Suites" action={<Pin className="w-3.5 h-3.5 text-[var(--accent)]" />}>
              <div className="space-y-2">
                {pinned.map((s) => <SuiteRow key={s.id} suite={s} />)}
              </div>
            </Section>
          )}
          <Section title="All Suites" action={<span className="text-[11px] text-[var(--text-muted)]">{suites.length} total</span>}>
            <div className="space-y-2">
              {others.map((s) => <SuiteRow key={s.id} suite={s} />)}
            </div>
          </Section>
        </>
      )}
    </div>
  );
}

function SuiteRow({ suite }: { suite: Suite }) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">{suite.name}</span>
          {suite.pinned && <Pin className="w-3 h-3 text-[var(--accent)]" />}
        </div>
        {suite.tags && suite.tags.length > 0 && (
          <div className="flex gap-1 mt-1">
            {suite.tags.map((tag) => (
              <span key={tag} className="px-1.5 py-0.5 rounded text-[10px] bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-muted)]">{tag}</span>
            ))}
          </div>
        )}
      </div>
      <span className="text-[11px] text-[var(--text-muted)]">{suite.run_ids?.length || 0} runs</span>
    </div>
  );
}
