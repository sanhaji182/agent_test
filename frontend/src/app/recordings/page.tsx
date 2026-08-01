"use client";

import { useEffect, useMemo, useState } from "react";
import {
  listRecordingSessions,
  createRecordingSession,
  deleteRecordingSession,
  type RecordingSession,
} from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Search, Plus, Trash2, X, Video, Globe, FolderOpen } from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export default function RecordingsPage() {
  const [sessions, setSessions] = useState<RecordingSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [showModal, setShowModal] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  // Form state
  const [formName, setFormName] = useState("");
  const [formProjectPath, setFormProjectPath] = useState("");
  const [formBaseUrl, setFormBaseUrl] = useState("");

  const fetchSessions = () => {
    listRecordingSessions()
      .then((s) => setSessions(s || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchSessions();
  }, []);

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    if (!q) return sessions;
    return sessions.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.base_url.toLowerCase().includes(q) ||
        s.status.toLowerCase().includes(q) ||
        (s.project_path || "").toLowerCase().includes(q)
    );
  }, [sessions, query]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim() || !formProjectPath.trim() || !formBaseUrl.trim()) return;
    setCreating(true);
    try {
      await createRecordingSession({
        name: formName.trim(),
        project_path: formProjectPath.trim(),
        base_url: formBaseUrl.trim(),
      });
      setShowModal(false);
      setFormName("");
      setFormProjectPath("");
      setFormBaseUrl("");
      fetchSessions();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this recording session? This cannot be undone.")) return;
    setDeleting(id);
    try {
      await deleteRecordingSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setDeleting(null);
    }
  };

  const resetForm = () => {
    setFormName("");
    setFormProjectPath("");
    setFormBaseUrl("");
    setShowModal(false);
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" />
        <LoadingSkeleton rows={8} />
      </div>
    );
  }
  if (error && sessions.length === 0) {
    return (
      <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5 text-sm text-[var(--danger)]">
        Error: {error}
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold">Recording Sessions</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-0.5">
            Record browser interactions and generate Playwright tests
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-sm font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
        >
          <Plus className="w-4 h-4" />
          New Recording
        </button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <input
          type="text"
          placeholder="Search sessions by name, URL, or project path..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 rounded-lg bg-[var(--bg-card)] border border-[var(--border)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-2 focus:ring-[var(--accent)]/10"
        />
      </div>

      {/* Error banner (non-blocking) */}
      {error && sessions.length > 0 && (
        <div className="rounded-lg border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-3 text-xs text-[var(--danger)] flex items-center justify-between">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="hover:text-[var(--danger)]/70">
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Sessions list */}
      {filtered.length === 0 ? (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
          <EmptyState
            icon={<Video className="w-6 h-6" />}
            title={query ? "No matching sessions" : "No recording sessions"}
            description={
              query
                ? "Try a different search term."
                : "Record browser interactions to generate Playwright tests automatically."
            }
            action={
              !query ? (
                <button
                  onClick={() => setShowModal(true)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
                >
                  <Plus className="w-3.5 h-3.5" />
                  Create First Recording
                </button>
              ) : undefined
            }
          />
        </div>
      ) : (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] overflow-hidden">
          {/* Table header */}
          <div className="grid grid-cols-[1fr_180px_100px_80px_160px_40px] gap-3 px-4 py-2.5 border-b border-[var(--border)] bg-[var(--bg-subtle)] text-[11px] font-semibold text-[var(--text-muted)] uppercase tracking-wider">
            <span>Session</span>
            <span>Base URL</span>
            <span>Events</span>
            <span>Status</span>
            <span>Created</span>
            <span />
          </div>
          {/* Table rows */}
          <div className="divide-y divide-[var(--border)]">
            {filtered.map((session) => (
              <div
                key={session.id}
                className="grid grid-cols-[1fr_180px_100px_80px_160px_40px] gap-3 px-4 py-3 items-center hover:bg-[var(--bg-hover)] transition-colors"
              >
                {/* Name + project path */}
                <div className="min-w-0">
                  <Link
                    href={`/recordings/${session.id}`}
                    className="text-sm font-medium text-[var(--text-primary)] hover:text-[var(--accent)] transition-colors block truncate"
                  >
                    {session.name}
                  </Link>
                  <div className="flex items-center gap-1 mt-0.5 text-[11px] text-[var(--text-muted)] truncate">
                    <FolderOpen className="w-3 h-3 shrink-0" />
                    <span className="truncate">{session.project_path}</span>
                  </div>
                </div>
                {/* Base URL */}
                <div className="flex items-center gap-1 text-xs text-[var(--text-secondary)] truncate">
                  <Globe className="w-3 h-3 shrink-0 text-[var(--text-muted)]" />
                  <span className="truncate">{session.base_url}</span>
                </div>
                {/* Event count */}
                <span className="text-xs text-[var(--text-muted)] tabular-nums">
                  {session.event_count ?? "—"}
                </span>
                {/* Status */}
                <StatusBadge state={session.status} />
                {/* Created */}
                <span className="text-[11px] text-[var(--text-muted)] tabular-nums">
                  {new Date(session.created_at).toLocaleString(undefined, {
                    month: "short",
                    day: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </span>
                {/* Delete */}
                <button
                  onClick={() => handleDelete(session.id)}
                  disabled={deleting === session.id}
                  className="flex items-center justify-center w-8 h-8 rounded-md text-[var(--text-muted)] hover:text-[var(--danger)] hover:bg-[var(--danger-bg)] transition-colors disabled:opacity-50"
                  title="Delete session"
                >
                  <Trash2
                    className={cn("w-4 h-4", deleting === session.id && "animate-pulse")}
                  />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Create modal */}
      {showModal && (
        <>
          <div className="fixed inset-0 z-40 bg-black/30" onClick={resetForm} />
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 pointer-events-none">
            <div className="pointer-events-auto w-full max-w-md rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-xl">
              <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
                <h2 className="text-sm font-bold">New Recording Session</h2>
                <button
                  onClick={resetForm}
                  className="w-7 h-7 flex items-center justify-center rounded-md text-[var(--text-muted)] hover:bg-[var(--bg-hover)] transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              <form onSubmit={handleCreate} className="p-5 space-y-4">
                <div>
                  <label className="block text-[11px] font-semibold text-[var(--text-secondary)] mb-1.5 uppercase tracking-wider">
                    Session Name
                  </label>
                  <input
                    type="text"
                    value={formName}
                    onChange={(e) => setFormName(e.target.value)}
                    placeholder="e.g. Login Flow Recording"
                    required
                    className="w-full px-3 py-2 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-2 focus:ring-[var(--accent)]/10"
                  />
                </div>
                <div>
                  <label className="block text-[11px] font-semibold text-[var(--text-secondary)] mb-1.5 uppercase tracking-wider">
                    Project Path
                  </label>
                  <input
                    type="text"
                    value={formProjectPath}
                    onChange={(e) => setFormProjectPath(e.target.value)}
                    placeholder="e.g. /path/to/project"
                    required
                    className="w-full px-3 py-2 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-2 focus:ring-[var(--accent)]/10"
                  />
                </div>
                <div>
                  <label className="block text-[11px] font-semibold text-[var(--text-secondary)] mb-1.5 uppercase tracking-wider">
                    Base URL
                  </label>
                  <input
                    type="url"
                    value={formBaseUrl}
                    onChange={(e) => setFormBaseUrl(e.target.value)}
                    placeholder="https://example.com"
                    required
                    className="w-full px-3 py-2 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]/50 focus:ring-2 focus:ring-[var(--accent)]/10"
                  />
                </div>
                <div className="flex justify-end gap-2 pt-2">
                  <button
                    type="button"
                    onClick={resetForm}
                    className="px-4 py-2 rounded-lg border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={creating}
                    className="px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-xs font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-60 transition-colors"
                  >
                    {creating ? "Creating..." : "Create Session"}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
