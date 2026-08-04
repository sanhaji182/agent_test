"use client";

import { useEffect, useMemo, useState } from "react";
import { getRuns, isActive, type TestRun } from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { RunInspector } from "@/components/console/inspector";
import { Search, Inbox, Sparkles } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";
import { cn } from "@/lib/utils";

type Group = "all" | "active" | "passed" | "failed";
type Sort = "newest" | "oldest" | "status" | "name";

export default function RunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [group, setGroup] = useState<Group>("all");
  const [sort, setSort] = useState<Sort>("newest");
  const [selected, setSelected] = useState<TestRun | null>(null);

  useEffect(() => {
    getRuns().then((r) => setRuns(r || [])).catch((e) => setError(e.message)).finally(() => setLoading(false));
    
    // Poll for updates
    const interval = setInterval(() => {
      getRuns().then((r) => setRuns(r || [])).catch(() => {});
    }, 10000);
    
    return () => clearInterval(interval);
  }, []);

  const filtered = useMemo(() => {
    let list = [...runs];
    
    // Filter by group
    if (group === "active") list = list.filter((r) => isActive(r.state));
    else if (group === "passed") list = list.filter((r) => r.state === "done");
    else if (group === "failed") list = list.filter((r) => r.state === "failed");
    
    // Search filter
    if (query.trim()) {
      const q = query.toLowerCase();
      list = list.filter(
        (r) =>
          r.id.includes(q) ||
          r.state.toLowerCase().includes(q) ||
          (r.requirements || "").toLowerCase().includes(q) ||
          (r.project_path || "").toLowerCase().includes(q)
      );
    }
    
    // Sorting
    list.sort((a, b) => {
      switch (sort) {
        case "oldest":
          return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
        case "status":
          return isActive(b.state) ? -1 : 1;
        case "name":
          return (b.requirements || "").localeCompare(a.requirements || "");
        case "newest":
        default:
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
      }
    });
    
    return list;
  }, [runs, query, group, sort]);

  const counts = {
    all: runs.length,
    active: runs.filter((r) => isActive(r.state)).length,
    passed: runs.filter((r) => r.state === "done").length,
    failed: runs.filter((r) => r.state === "failed").length,
  };

  if (loading) {
    return <div className="space-y-6"><LoadingSkeleton rows={8} /></div>;
  }
  
  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-6">
        <h2 className="text-sm font-semibold text-red-700 mb-2">Error loading runs</h2>
        <p className="text-sm text-red-600">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Test Runs</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Browse and inspect test execution history</p>
        </div>
      </div>

      {/* Search & Actions */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <Input
            placeholder="Search by ID, state, or requirements..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        
        {/* Status Filters - Modern Segment Control */}
        <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
          {(["all", "active", "passed", "failed"] as Group[]).map((g) => (
            <button
              key={g}
              onClick={() => setGroup(g)}
              className={cn(
                "px-3 py-1.5 text-xs font-medium rounded-md transition-colors",
                group === g 
                  ? "bg-blue-600 text-white shadow-sm" 
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
              )}
            >
              {g.charAt(0).toUpperCase() + g.slice(1)}
              <span className={cn("ml-1.5", group === g ? "text-blue-200" : "text-[var(--text-muted)]")}>({counts[g]})</span>
            </button>
          ))}
        </div>
      </div>

      {/* Results */}
      {filtered.length === 0 ? (
        <div className="rounded-lg border border-[var(--border-default)] bg-white">
          <EmptyState 
            icon={<Inbox className="w-6 h-6" />} 
            title={query ? "No matching runs" : "No runs found"} 
            description={query ? "Try adjusting your search criteria." : "Create your first test run to start monitoring executions."}
            action={!query && runs.length === 0 ? (
              <Link href="/create">
                <Button>Create First Test</Button>
              </Link>
            ) : undefined}
          />
        </div>
      ) : (
        <div className="rounded-lg border border-[var(--border-default)] bg-white overflow-hidden">
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
	              <tr>
	                <Th>Status</Th>
	                <Th>Run Name / ID</Th>
	                <Th>Model</Th>
	                <Th>Coverage</Th>
	                <Th>Duration</Th>
	                <Th>Started</Th>
	                <Th align="right">Actions</Th>
	              </tr>
            </thead>
            <tbody>
              {filtered.map((r) => (
                <Tr 
                  key={r.id} 
                  onClick={() => setSelected(r)}
                  hover
                  className={selected?.id === r.id ? "bg-blue-50" : ""}
                >
                  <Td><StatusBadge state={r.state} size="sm" /></Td>
	                  <Td className="font-medium">
	                    <span className="block truncate max-w-[200px]">{r.requirements || "Untitled test"}</span>
	                    <span className="text-xs text-[var(--text-muted)] font-mono">{r.id.slice(0, 8)}</span>
	                  </Td>
	                  <Td>
	                    {r.llm_model ? (
	                      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-[var(--accent-light)] border border-[var(--accent)]/15 text-[11px] font-medium text-[var(--accent)] whitespace-nowrap" title={r.llm_provider ? `Provider: ${r.llm_provider}` : undefined}>
	                        <Sparkles className="w-3 h-3" />
	                        {r.llm_model}
	                        {r.model_override && <span className="text-[var(--success)] font-semibold">*</span>}
	                        {r.llm_fallback_model && <span className="text-[var(--text-muted)] font-normal">+fb</span>}
	                      </span>
	                    ) : (
	                      <span className="text-sm text-[var(--text-muted)]">-</span>
	                    )}
	                  </Td>
                  <Td>
                    {r.run_result ? (
                      <span className="text-sm text-[var(--text-primary)]">
                        <span className="text-green-600">{r.run_result.passed}✓</span>
                        <span className="mx-1">·</span>
                        <span className="text-red-600">{r.run_result.failed}✗</span>
                      </span>
                    ) : (
                      "-"
                    )}
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">
                    {r.run_result?.duration_ms 
                      ? formatDuration(r.run_result.duration_ms) 
                      : "-"}
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {formatDate(r.created_at)}
                  </Td>
                  <Td>
                    <Link href={`/runs/${r.id}`} className="text-blue-600 hover:text-blue-700 text-xs font-medium text-right">
                      View details
                    </Link>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Inspector Drawer */}
      {selected && <RunInspector run={selected} onClose={() => setSelected(null)} />}
      
      {/* Backdrop */}
      {selected && (
        <div 
          className="fixed inset-0 z-30 bg-black/20 cursor-pointer" 
          onClick={() => setSelected(null)}
        />
      )}
    </div>
  );
}

// === Components ===

function Th({ children, align = "left" }: { children: React.ReactNode; align?: string }) {
  return (
    <th className={`px-4 py-3 text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] ${align === "right" ? "text-right" : ""}`}>
      {children}
    </th>
  );
}

function Tr({ children, onClick, hover, className }: { 
  children: React.ReactNode; 
  onClick?: () => void; 
  hover?: boolean;
  className?: string;
}) {
  return (
    <tr 
      className={`border-b border-[var(--border-default)] ${onClick ? "cursor-pointer" : ""} ${hover && !onClick ? "hover:bg-gray-50 transition-colors" : ""} ${className || ""}`}
      onClick={onClick}
    >
      {children}
    </tr>
  );
}

function Td({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <td className={`px-4 py-3 text-sm ${className || ""}`}>
      {children}
    </td>
  );
}

function Button({ children, variant = "primary" }: { children: React.ReactNode; variant?: "primary" }) {
  const baseStyles = "inline-flex items-center justify-center px-4 py-2 text-sm font-medium rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";
  const variants: Record<string, string> = {
    primary: "bg-blue-600 text-white hover:bg-blue-700 shadow-xs",
  };
  
  return (
    <button className={`${baseStyles} ${variants[variant]}`}>
      {children}
    </button>
  );
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const mins = Math.floor(ms / 60000);
  const secs = ((ms / 1000) % 60).toFixed(0);
  return `${mins}m ${secs}s`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);
  
  if (diffMin < 1) return "Just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}
