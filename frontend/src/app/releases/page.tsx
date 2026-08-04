"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Tag, Clock, CheckCircle2, AlertTriangle, ArrowUpRight } from "lucide-react";

import { getReleases, type Release } from "@/lib/api";

export default function ReleasesPage() {
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getReleases()
      .then(setReleases)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Releases</h1>
        <p className="text-sm text-[var(--text-muted)] mt-1">Track test suite versions and deployment history</p>
      </div>

      {/* Release History */}
      <Section title="Release History">
        {releases.length === 0 ? (
          <EmptyState 
            icon={<Tag className="w-8 h-8" />}
            title="No releases yet"
            description="Test releases are automatically created when tests are deployed to production."
          />
        ) : (
          <div className="space-y-4">
            {releases.map((release, index) => (
              <ReleaseCard key={release.id} release={release} isFirst={index === 0} />
            ))}
          </div>
        )}
      </Section>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatBox label="Total Releases" value={releases.length.toString()} />
        <StatBox label="Runs in Latest Release" value={releases[0]?.run_ids?.length?.toString() || "0"} positive />
        <StatBox label="Latest Status" value={releases[0]?.status || "-"} />
      </div>
    </div>
  );
}

function ReleaseCard({ release, isFirst }: { release: Release; isFirst: boolean }) {
  return (
    <div className={`p-4 rounded-lg border ${isFirst ? "bg-blue-50 border-blue-200" : "bg-white border-[var(--border-default)]"}`}>
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className={`shrink-0 p-2 rounded-lg ${isFirst ? "bg-blue-100" : "bg-gray-100"}`}>
            <Tag className={`w-5 h-5 ${isFirst ? "text-blue-600" : "text-[var(--text-muted)]"}`} />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="text-base font-semibold text-[var(--text-primary)]">{release.name}</h3>
              {isFirst && (
                <Badge variant="success" size="sm">Current</Badge>
              )}
            </div>
            <p className="text-sm text-[var(--text-secondary)]">{release.version}{release.status ? ` · ${release.status}` : ""}</p>
          </div>
        </div>
        
        {isFirst && (
          <Link href={`/releases/${release.id}`}>
            <Button variant="ghost" size="sm">
              View Details <ArrowUpRight className="w-3.5 h-3.5 ml-1" />
            </Button>
          </Link>
        )}
      </div>

      {isFirst && (
        <div className="flex items-center gap-4 mt-4 pt-4 border-t border-blue-200">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-green-600" />
            <span className="text-xs text-[var(--text-muted)]">{release.run_ids?.length || 0} runs</span>
          </div>
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-yellow-600" />
            <span className="text-xs text-[var(--text-muted)]">{release.status || "unknown"}</span>
          </div>
          <div className="flex items-center gap-2 ml-auto">
            <Clock className="w-4 h-4 text-[var(--text-muted)]" />
            <span className="text-xs text-[var(--text-muted)] whitespace-nowrap">
              {formatDate(release.created_at)}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

function StatBox({ label, value, positive }: { label: string; value: string; positive?: boolean }) {
  return (
    <div className="bg-white rounded-lg p-4 border border-[var(--border-default)]">
      <p className="text-xs text-[var(--text-muted)] font-medium uppercase tracking-wide">{label}</p>
      <p className={`text-2xl font-semibold mt-2 ${positive ? "text-green-600" : ""}`}>{value}</p>
    </div>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}
