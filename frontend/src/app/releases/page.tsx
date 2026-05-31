"use client";

import { useEffect, useState } from "react";
import { getReleases, type Release } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Tag } from "lucide-react";

export default function ReleasesPage() {
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getReleases().then(setReleases).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="space-y-6"><LoadingSkeleton rows={4} /></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">Releases</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Organize test runs by release version</p>
      </div>

      <Section title="All Releases">
        {releases.length === 0 ? (
          <EmptyState
            icon={<Tag className="w-6 h-6" />}
            title="No releases"
            description="Create a release via POST /api/v1/releases to group runs by version."
          />
        ) : (
          <div className="space-y-2">
            {releases.map((rel) => (
              <div key={rel.id} className="flex items-center justify-between p-4 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                <div>
                  <span className="text-sm font-semibold">{rel.name}</span>
                  <span className="ml-2 text-xs text-[var(--text-muted)]">v{rel.version}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${rel.status === "active" ? "bg-[var(--success-bg)] text-[var(--success)]" : "bg-[var(--bg-subtle)] text-[var(--text-muted)]"}`}>
                    {rel.status}
                  </span>
                  <span className="text-[11px] text-[var(--text-muted)]">{rel.run_ids?.length || 0} runs</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}
