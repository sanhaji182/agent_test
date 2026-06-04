"use client";

import { useEffect, useState } from "react";
import { getRuns, type TestRun } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Download, FileJson } from "lucide-react";

const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function ExportsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getRuns().then((r) => setRuns(r || [])).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSkeleton rows={4} />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Exports</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Download test data as JSON for reporting, CI integration, or archival.</p>
      </div>

      <Section title="Available Exports">
        <div className="space-y-2">
          <ExportRow
            label="Risk Report"
            desc="Current risk scores for all tests and schedules"
            href={`${API}/api/v1/metrics/risk/export`}
          />
          {runs.slice(0, 5).map((run) => (
            <ExportRow
              key={run.id}
              label={`Run ${run.id.slice(0, 8)}`}
              desc={run.requirements || "Test run data"}
              href={`${API}/api/v1/runs/${run.id}/export`}
            />
          ))}
        </div>
        {runs.length === 0 && (
          <EmptyState
            icon={<FileJson className="w-6 h-6" />}
            title="No runs to export"
            description="Create a test run first. Exports will appear here once you have data."
          />
        )}
      </Section>
    </div>
  );
}

function ExportRow({ label, desc, href }: { label: string; desc: string; href: string }) {
  return (
    <a
      href={href}
      target="_blank"
      className="flex items-center gap-3 p-3 rounded-[var(--radius-sm)] border border-[var(--border)] hover:border-[var(--accent-light)] hover:shadow-[var(--shadow-sm)] transition-all group"
    >
      <div className="p-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] text-[var(--text-muted)] group-hover:text-[var(--accent)] transition-colors">
        <Download className="w-4 h-4" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-[12px] font-semibold text-[var(--text-primary)]">{label}</p>
        <p className="text-[11px] text-[var(--text-muted)] truncate">{desc}</p>
      </div>
      <span className="text-[10px] font-medium text-[var(--text-muted)] uppercase">JSON</span>
    </a>
  );
}
