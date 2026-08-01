"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Video, PlayCircle, Trash2, Search, Download, Eye, Plus } from "lucide-react";

export default function RecordingsPage() {
  const [recordings, setRecordings] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [filterStatus, setFilterStatus] = useState("all");

  useEffect(() => {
    // TODO: Fetch real recordings data
    setTimeout(() => setLoading(false), 500);
  }, []);

  const filteredRecordings = recordings.filter(rec => {
    if (filterStatus !== "all" && rec.status !== filterStatus) return false;
    if (query && !rec.name.toLowerCase().includes(query.toLowerCase())) return false;
    return true;
  });

  const getCounts = () => ({
    all: recordings.length,
    recording: recordings.filter(r => r.status === "recording").length,
    completed: recordings.filter(r => r.status === "completed").length,
    archived: recordings.filter(r => r.status === "archived").length,
  });

  if (loading) return <LoadingSkeleton rows={6} />;

  const counts = getCounts();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Recordings</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Browser recordings and test generation sessions</p>
        </div>
        <Link href="/create">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            New Recording
          </Button>
        </Link>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <Input
            placeholder="Search recordings..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        {/* Status Filter - Modern Segment Control */}
        <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
          {[
            { value: "all", label: "All" },
            { value: "recording", label: "Active" },
            { value: "completed", label: "Completed" },
            { value: "archived", label: "Archived" },
          ].map((f) => (
            <button
              key={f.value}
              onClick={() => setFilterStatus(f.value)}
              className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                filterStatus === f.value
                  ? "bg-blue-600 text-white shadow-sm"
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
              }`}
            >
              {f.label} ({counts[f.value as keyof typeof counts]})
            </button>
          ))}
        </div>
      </div>

      {/* Recordings List */}
      {filteredRecordings.length === 0 ? (
        <Section title="No recordings found">
          <EmptyState 
            icon={<Video className="w-8 h-8" />}
            title={!query ? "No recordings yet" : "No matching recordings"}
            description={!query ? "Start recording your browser interactions to generate automated tests." : "Try adjusting your search criteria."}
            action={!query && (
              <Link href="/create">
                <Button>New Recording</Button>
              </Link>
            )}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Name</Th>
                <Th>Status</Th>
                <Th>Events</Th>
                <Th>Duration</Th>
                <Th>Date</Th>
                <Th align="right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {filteredRecordings.map((rec) => (
                <Tr key={rec.id} hover>
                  <Td className="font-medium">
                    <span className="block truncate max-w-[180px]">{rec.name}</span>
                    <span className="text-xs text-[var(--text-muted)]">{rec.description || ""}</span>
                  </Td>
                  <Td>
                    <Badge 
                      variant={rec.status === "recording" ? "warning" : rec.status === "completed" ? "success" : "default"}
                      size="sm"
                    >
                      {rec.status}
                    </Badge>
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">{rec.events?.length || 0} events</Td>
                  <Td className="text-sm text-[var(--text-muted)]">{formatDuration(rec.duration)}</Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">{formatDate(rec.date)}</Td>
                  <Td align="right" className="space-x-2">
                    <Link href={`/recordings/${rec.id}`}>
                      <Button variant="ghost" size="sm">
                        <Eye className="w-3.5 h-3.5" />
                      </Button>
                    </Link>
                    <Button variant="ghost" size="sm">
                      <Download className="w-3.5 h-3.5" />
                    </Button>
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50">
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      )}
    </div>
  );
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}m ${secs}s`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);
  
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}
