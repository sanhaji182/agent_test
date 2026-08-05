"use client";

import React, { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  getRecordingSession,
  generateRecordingTest,
  createTestCaseFromRecording,
  deleteRecordingSession,
  type RecordingSession,
  type RecordedEvent,
} from "@/lib/api";
import { StatusBadge } from "@/components/ui/badge";
import { LoadingSkeleton, EmptyState } from "@/components/ui/section";
import {
  ArrowLeft,
  Globe,
  FolderOpen,
  MousePointerClick,
  Type,
  Navigation,
  ArrowUpDown,
  Keyboard,
  MousePointer,
  Sparkles,
  Copy,
  Check,
  CheckCircle2,
  Trash2,
  Clock,
  Hash,
  Monitor,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";

const eventIcons: Record<string, React.ElementType> = {
  click: MousePointerClick,
  dblclick: MousePointerClick,
  fill: Type,
  type: Type,
  navigate: Navigation,
  goto: Navigation,
  scroll: ArrowUpDown,
  keydown: Keyboard,
  keyup: Keyboard,
  hover: MousePointer,
  mousemove: MousePointer,
  screenshot: Monitor,
};

const eventColors: Record<string, string> = {
  click: "text-[var(--accent)]",
  dblclick: "text-[var(--accent)]",
  fill: "text-[var(--info)]",
  type: "text-[var(--info)]",
  navigate: "text-[var(--success)]",
  goto: "text-[var(--success)]",
  scroll: "text-[var(--warning)]",
  keydown: "text-[var(--warning)]",
  keyup: "text-[var(--warning)]",
  hover: "text-[var(--text-muted)]",
  mousemove: "text-[var(--text-muted)]",
  screenshot: "text-[var(--text-secondary)]",
};

export default function RecordingDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [session, setSession] = useState<RecordingSession | null>(null);
  const [events, setEvents] = useState<RecordedEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

	  // Generate test state
	  const [testCode, setTestCode] = useState<string | null>(null);
	  const [testMeta, setTestMeta] = useState<{ language: string; framework: string } | null>(null);
	  const [generating, setGenerating] = useState(false);
	  const [generateError, setGenerateError] = useState<string | null>(null);

	  // Save-as-test-case state (deterministic playback — like Katalon record & playback)
	  const [savingCase, setSavingCase] = useState(false);
	  const [savedCase, setSavedCase] = useState<{ id: string; title: string; steps: string[] } | null>(null);
	  const [saveCaseError, setSaveCaseError] = useState<string | null>(null);

  // Copy state
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    getRecordingSession(id)
      .then((data) => {
        setSession(data.session);
        setEvents(data.events || []);
      })
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoading(false));
  }, [id]);

	  const handleGenerate = async () => {
	    setGenerating(true);
	    setGenerateError(null);
	    setTestCode(null);
	    try {
	      const result = await generateRecordingTest(id);
	      setTestCode(result.test_code);
	      setTestMeta({ language: result.language, framework: result.framework });
	    } catch (e) {
	      setGenerateError((e as Error).message);
	    } finally {
	      setGenerating(false);
	    }
	  };

	  const handleSaveAsTestCase = async () => {
	    setSavingCase(true);
	    setSaveCaseError(null);
	    try {
	      const tc = await createTestCaseFromRecording(id);
	      setSavedCase({ id: tc.id, title: tc.title, steps: tc.steps || [] });
	    } catch (e) {
	      setSaveCaseError((e as Error).message);
	    } finally {
	      setSavingCase(false);
	    }
	  };

  const handleCopy = async () => {
    if (!testCode) return;
    try {
      await navigator.clipboard.writeText(testCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const textarea = document.createElement("textarea");
      textarea.value = testCode;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Hapus sesi rekaman ini beserta semua event-nya? Tindakan ini tidak bisa dibatalkan.")) return;
    try {
      await deleteRecordingSession(id);
      router.push("/recordings");
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const formatTimestamp = (ts: string) => {
    const d = new Date(ts);
    return d.toLocaleTimeString(undefined, {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      fractionalSecondDigits: 3,
    } as Intl.DateTimeFormatOptions);
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 rounded-lg bg-[var(--bg-subtle)] animate-pulse" />
        <LoadingSkeleton rows={8} />
      </div>
    );
  }
  if (error || !session) {
    return (
      <div className="space-y-4">
        <Link
          href="/recordings"
          className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
        >
          <ArrowLeft className="w-3 h-3" /> Kembali ke rekaman
        </Link>
        <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5 text-sm text-[var(--danger)]">
          {error || "Sesi tidak ditemukan"}
        </div>
      </div>
    );
  }

  const EventIcon = (eventType: string) => eventIcons[eventType] || Monitor;
  const eventIconColor = (eventType: string) => eventColors[eventType] || "text-[var(--text-muted)]";

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <Link
          href="/recordings"
          className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
        >
          <ArrowLeft className="w-3 h-3" /> Kembali ke rekaman
        </Link>
        <button
          onClick={handleDelete}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:text-[var(--danger)] hover:border-[var(--danger)]/30 hover:bg-[var(--danger-bg)] transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
          Delete
        </button>
      </div>

      {/* Session info header */}
      <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] p-5">
        <div className="flex items-start justify-between mb-4">
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-lg font-bold">{session.name}</h1>
              <StatusBadge state={session.status} />
            </div>
            <div className="flex items-center gap-4 mt-2 text-xs text-[var(--text-secondary)]">
              <span className="flex items-center gap-1">
                <Globe className="w-3.5 h-3.5 text-[var(--text-muted)]" />
                {session.base_url}
              </span>
              <span className="flex items-center gap-1">
                <FolderOpen className="w-3.5 h-3.5 text-[var(--text-muted)]" />
                <span className="font-mono">{session.project_path}</span>
              </span>
            </div>
          </div>
	          <button
	            onClick={handleGenerate}
	            disabled={generating || events.length === 0}
	            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-xs font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50 transition-colors shadow-sm"
	          >
	            <Sparkles className={cn("w-3.5 h-3.5", generating && "animate-pulse")} />
	            {generating ? "Membuat…" : "Buat Test"}
	          </button>
	        </div>

	        {/* Save as deterministic test case (record & playback — seperti Katalon) */}
	        <div className="mt-4 flex flex-wrap items-center gap-2">
	          <button
	            onClick={handleSaveAsTestCase}
	            disabled={savingCase || events.length === 0 || !!savedCase}
	            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-[var(--success-bg)] border border-[var(--success)]/25 text-[var(--success)] text-xs font-semibold hover:opacity-90 disabled:opacity-50 transition-colors"
	            title="Simpan hasil rekam sebagai test case deterministik — bisa di-run ulang persis tanpa AI"
	          >
	            <CheckCircle2 className="w-3.5 h-3.5" />
	            {savingCase ? "Menyimpan…" : savedCase ? "Tersimpan sebagai Test Case" : "Simpan sebagai Test Case"}
	          </button>
	          {savedCase && (
	            <Link href="/tests" className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] transition-colors">
	              Buka Perpustakaan Test →
	            </Link>
	          )}
	          {saveCaseError && (
	            <span className="text-xs text-[var(--danger)]">{saveCaseError}</span>
	          )}
	        </div>

        {/* Meta grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
          <Meta label="Status" value={session.status} />
          <Meta label="Events" value={String(events.length)} />
          <Meta
            label="Created"
            value={new Date(session.created_at).toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          />
          <Meta label="ID Sesi" value={session.id.slice(0, 8)} mono />
        </div>
      </div>

      {/* Generated test code */}
      {(testCode || generateError) && (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] overflow-hidden">
          <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border)]">
            <div className="flex items-center gap-2">
              <h2 className="text-[13px] font-semibold text-[var(--text-primary)]">
                Test yang Dihasilkan
              </h2>
              {testMeta && (
                <span className="text-[10px] text-[var(--text-muted)] bg-[var(--bg-subtle)] px-2 py-0.5 rounded-full font-mono">
                  {testMeta.language} · {testMeta.framework}
                </span>
              )}
            </div>
            {testCode && (
              <button
                onClick={handleCopy}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[11px] font-medium transition-colors border border-[var(--border)] text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-[var(--success)]" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    Salin kode
                  </>
                )}
              </button>
            )}
          </div>
          <div className="p-5">
            {generateError ? (
              <div className="rounded-lg border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-4 text-xs text-[var(--danger)]">
                {generateError}
              </div>
            ) : testCode ? (
              <pre className="text-[12px] leading-relaxed font-mono text-[var(--text-primary)] bg-[var(--bg-subtle)] rounded-lg p-4 overflow-auto max-h-[500px] border border-[var(--border)]">
                <code>{testCode}</code>
              </pre>
            ) : null}
          </div>
        </div>
      )}

      {/* Events timeline */}
      <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)]">
        <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border)]">
          <h2 className="text-[13px] font-semibold text-[var(--text-primary)]">
            Events
          </h2>
          <span className="text-[11px] text-[var(--text-muted)]">
            {events.length} event{events.length !== 1 ? "s" : ""}
          </span>
        </div>
        {events.length === 0 ? (
          <div className="p-5">
            <EmptyState
              icon={<MousePointerClick className="w-5 h-5" />}
              title="Belum ada event"
              description="Mulai rekam interaksi browser untuk melihat event muncul di sini."
            />
          </div>
        ) : (
          <div className="divide-y divide-[var(--border)] max-h-[600px] overflow-auto">
            {events
              .sort((a, b) => a.sequence_order - b.sequence_order)
              .map((event) => {
                const Icon = EventIcon(event.event_type);
                return (
                  <div
                    key={event.id}
                    className="flex items-start gap-3 px-5 py-3 hover:bg-[var(--bg-hover)] transition-colors"
                  >
                    {/* Event icon */}
                    <div
                      className={cn(
                        "w-8 h-8 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] flex items-center justify-center shrink-0 mt-0.5",
                        eventIconColor(event.event_type)
                      )}
                    >
                      <Icon className="w-4 h-4" />
                    </div>
                    {/* Event details */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-[11px] font-semibold text-[var(--text-primary)] uppercase tracking-wider">
                          {event.event_type}
                        </span>
                        <span className="text-[10px] text-[var(--text-muted)] font-mono">
                          #{event.sequence_order}
                        </span>
                      </div>
                      <div className="mt-0.5 text-xs text-[var(--text-secondary)] space-y-0.5">
                        {event.selector && (
                          <div className="flex items-center gap-1.5">
                            <Hash className="w-3 h-3 text-[var(--text-muted)] shrink-0" />
                            <code className="text-[11px] font-mono text-[var(--accent)] bg-[var(--accent-bg)] px-1 py-px rounded truncate block">
                              {event.selector}
                            </code>
                          </div>
                        )}
                        {event.value && (
                          <div className="flex items-center gap-1.5">
                            <Type className="w-3 h-3 text-[var(--text-muted)] shrink-0" />
                            <span className="truncate">{event.value}</span>
                          </div>
                        )}
                        {event.url && (
                          <div className="flex items-center gap-1.5">
                            <Globe className="w-3 h-3 text-[var(--text-muted)] shrink-0" />
                            <span className="truncate font-mono text-[11px]">{event.url}</span>
                          </div>
                        )}
                      </div>
                    </div>
                    {/* Timestamp */}
                    <div className="flex items-center gap-1 text-[10px] text-[var(--text-muted)] font-mono shrink-0 tabular-nums">
                      <Clock className="w-3 h-3" />
                      {formatTimestamp(event.timestamp)}
                    </div>
                  </div>
                );
              })}
          </div>
        )}
      </div>
    </div>
  );
}

function Meta({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="px-3 py-2 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
      <p className="text-[10px] text-[var(--text-muted)] uppercase tracking-wider font-semibold">
        {label}
      </p>
      <p className={cn("text-xs text-[var(--text-primary)] mt-0.5 truncate", mono && "font-mono")}>
        {value}
      </p>
    </div>
  );
}
