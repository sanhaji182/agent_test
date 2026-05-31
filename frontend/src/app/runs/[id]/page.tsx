"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  getRun, subscribeToRun, rerunRun, reportUrl, isActive,
  getRunEvents, getRunRecordings, getRunVisuals,
  type TestRun, type RunEvent, type Recording, type VisualArtifact,
} from "@/lib/api";
import { StatusBadge, PriorityBadge } from "@/components/ui/badge";
import { LoadingSkeleton } from "@/components/ui/section";
import { ExecutionTimeline } from "@/components/console/timeline";
import { Tabs } from "@/components/console/tabs";
import { ScreenshotStrip } from "@/components/console/screenshot-strip";
import { EmptyState } from "@/components/ui/section";
import {
  ArrowLeft, FileText, RotateCw, FileCode, AlertTriangle,
  CheckCircle2, XCircle, Image as ImageIcon, ListChecks, Eye,
  Radio, Film, GitCompare,
} from "lucide-react";
import Link from "next/link";

export default function RunConsolePage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;
  const [run, setRun] = useState<TestRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [liveState, setLiveState] = useState<string | null>(null);
  const [liveEvents, setLiveEvents] = useState<RunEvent[]>([]);
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [visuals, setVisuals] = useState<VisualArtifact[]>([]);
  const [rerunning, setRerunning] = useState(false);

  useEffect(() => {
    let unsub: (() => void) | undefined;
    getRun(id)
      .then((r) => {
        setRun(r);
        // Load events, recordings, visuals
        getRunEvents(id).then(setLiveEvents).catch(() => {});
        getRunRecordings(id).then(setRecordings).catch(() => {});
        getRunVisuals(id).then(setVisuals).catch(() => {});

        if (isActive(r.state)) {
          unsub = subscribeToRun(id, (event) => {
            if (event.type === "state_change") setLiveState(event.data.state);
            if (event.type === "step") {
              setLiveEvents((prev) => [...prev, event.data as unknown as RunEvent]);
            }
            if (event.type === "done") {
              getRun(id).then(setRun);
              getRunRecordings(id).then(setRecordings).catch(() => {});
              getRunVisuals(id).then(setVisuals).catch(() => {});
            }
          });
        }
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
    return () => unsub?.();
  }, [id]);

  const handleRerun = async () => {
    setRerunning(true);
    try {
      const res = await rerunRun(id);
      router.push(`/runs/${res.run_id}`);
    } catch (e) {
      setError((e as Error).message);
      setRerunning(false);
    }
  };

  if (loading) return <div className="space-y-6"><LoadingSkeleton rows={6} /></div>;
  if (error || !run) return <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-5 text-sm text-[var(--danger)]">{error || "Run not found"}</div>;

  const displayState = liveState || run.state;
  const result = run.run_result;

  return (
    <div className="space-y-5">
      <Link href="/runs" className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)] hover:text-[var(--text-secondary)]">
        <ArrowLeft className="w-3 h-3" /> Back to suites
      </Link>

      {/* Summary bar */}
      <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] p-5">
        <div className="flex items-start justify-between mb-5">
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-lg font-bold">Execution <span className="font-mono text-[var(--accent)]">{run.id.slice(0, 8)}</span></h1>
              <StatusBadge state={displayState} />
            </div>
            <p className="text-xs text-[var(--text-secondary)] mt-1">{run.requirements || "No requirements specified"}</p>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={handleRerun} disabled={rerunning} className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg bg-[var(--accent)] text-white text-xs font-medium hover:bg-[var(--accent-hover)] disabled:opacity-60 transition-colors">
              <RotateCw className={`w-3.5 h-3.5 ${rerunning ? "animate-spin" : ""}`} /> Rerun
            </button>
            <a href={reportUrl(run.id)} target="_blank" className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] transition-colors">
              <FileText className="w-3.5 h-3.5" /> Report
            </a>
            <Link href={`/runs/${run.id}/compare`} className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg border border-[var(--border)] text-xs font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] transition-colors">
              <GitCompare className="w-3.5 h-3.5" /> Compare
            </Link>
          </div>
        </div>
        <div className="p-4 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
          <ExecutionTimeline state={displayState} />
        </div>
        {result && (
          <div className="flex items-center gap-6 mt-4">
            <Stat icon={<CheckCircle2 className="w-4 h-4" />} value={result.passed} label="Passed" color="success" />
            <Stat icon={<XCircle className="w-4 h-4" />} value={result.failed} label="Failed" color="danger" />
            <Stat icon={<ListChecks className="w-4 h-4" />} value={result.total} label="Total" color="default" />
            <Stat icon={<RotateCw className="w-4 h-4" />} value={run.fix_attempts} label="Fixes" color="warning" />
          </div>
        )}
      </div>

      {run.error && (
        <div className="rounded-xl border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-4">
          <p className="text-xs font-semibold text-[var(--danger)] mb-1">Execution Error</p>
          <pre className="text-[11px] text-[var(--danger)]/80 whitespace-pre-wrap">{run.error}</pre>
        </div>
      )}

      {/* Tabbed console */}
      <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] p-5">
        <Tabs tabs={[
          { id: "video", label: "Video", content: <VideoPlayer run={run} /> },
          { id: "events", label: "Live Events", count: liveEvents.length, content: <EventsView events={liveEvents} /> },
          { id: "steps", label: "Steps", count: stepCount(run), content: <StepsView run={run} /> },
          { id: "files", label: "Files", count: run.test_files?.length, content: <FilesView run={run} /> },
          { id: "recordings", label: "Recordings", count: recordings.length, content: <RecordingsView recordings={recordings} /> },
          { id: "shots", label: "Screenshots", count: run.screenshots?.length, content: <ScreenshotStrip screenshots={run.screenshots} /> },
          { id: "failures", label: "Failures", count: result?.failures.length, content: <FailuresView run={run} /> },
          { id: "visual", label: "Visual", count: visuals.length, content: <VisualView artifacts={visuals} /> },
        ]} initial={run.video_url ? "video" : "events"} />
      </div>
    </div>
  );
}

function Stat({ icon, value, label, color }: { icon: React.ReactNode; value: number; label: string; color: string }) {
  const c: Record<string, string> = { success: "text-[var(--success)]", danger: "text-[var(--danger)]", warning: "text-[var(--warning)]", default: "text-[var(--text-primary)]" };
  return (
    <div className="flex items-center gap-2">
      <span className={c[color]}>{icon}</span>
      <span className={`text-lg font-bold ${c[color]}`}>{value}</span>
      <span className="text-[11px] text-[var(--text-muted)]">{label}</span>
    </div>
  );
}

function stepCount(run: TestRun): number {
  return (run.test_plan?.scenarios || []).reduce((n, s) => n + (s.steps?.length || 0), 0);
}

// Live events timeline
function EventsView({ events }: { events: RunEvent[] }) {
  if (events.length === 0) {
    return <EmptyState icon={<Radio className="w-6 h-6" />} title="No events yet" description="Events appear as the run executes. Subscribe to see live updates." />;
  }
  return (
    <div className="space-y-1 max-h-96 overflow-y-auto">
      {events.map((evt, i) => (
        <div key={evt.id || i} className="flex items-start gap-3 px-3 py-2 rounded-lg hover:bg-[var(--bg-subtle)] text-xs">
          <EventDot type={evt.type} />
          <div className="flex-1 min-w-0">
            <span className="font-medium text-[var(--text-primary)]">{evt.message}</span>
            {evt.metadata && Object.keys(evt.metadata).length > 0 && (
              <span className="ml-2 text-[var(--text-muted)]">
                {Object.entries(evt.metadata).map(([k, v]) => `${k}=${v}`).join(" ")}
              </span>
            )}
          </div>
          <span className="text-[10px] text-[var(--text-muted)] shrink-0">
            {evt.timestamp ? new Date(evt.timestamp).toLocaleTimeString() : ""}
          </span>
        </div>
      ))}
    </div>
  );
}

function EventDot({ type }: { type: string }) {
  const colors: Record<string, string> = {
    run_started: "bg-[var(--info)]", run_completed: "bg-[var(--success)]", run_failed: "bg-[var(--danger)]",
    assertion_passed: "bg-[var(--success)]", assertion_failed: "bg-[var(--danger)]",
    screenshot_captured: "bg-purple-500", fix_attempt_started: "bg-[var(--warning)]",
  };
  return <span className={`w-2 h-2 rounded-full mt-1.5 shrink-0 ${colors[type] || "bg-[var(--text-muted)]"}`} />;
}

// Recordings view
function RecordingsView({ recordings }: { recordings: Recording[] }) {
  if (recordings.length === 0) {
    return <EmptyState icon={<Film className="w-6 h-6" />} title="No recordings" description="Recordings are created when screenshots are captured during execution." />;
  }
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  return (
    <div className="space-y-2">
      {recordings.map((rec, i) => (
        <div key={rec.id || i} className="flex items-center gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
          <div className="w-10 h-10 rounded-md bg-[var(--bg-card)] border border-[var(--border)] flex items-center justify-center shrink-0">
            <ImageIcon className="w-4 h-4 text-[var(--text-muted)]" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-[var(--text-primary)] truncate">{rec.test_name}</p>
            <p className="text-[11px] text-[var(--text-muted)]">{rec.step_name} · {rec.status}</p>
          </div>
          {rec.screenshot_url && (
            <a href={rec.screenshot_url.startsWith("http") ? rec.screenshot_url : `${API}${rec.screenshot_url}`} target="_blank" className="text-[11px] text-[var(--accent)] hover:underline shrink-0">View</a>
          )}
        </div>
      ))}
    </div>
  );
}

// Steps view
function StepsView({ run }: { run: TestRun }) {
  const scenarios = run.test_plan?.scenarios || [];
  if (scenarios.length === 0) return <EmptyState icon={<ListChecks className="w-6 h-6" />} title="No steps yet" description="Steps appear once the test plan is generated." />;
  return (
    <div className="space-y-4">
      {run.test_plan?.summary && <p className="text-sm text-[var(--text-secondary)]">{run.test_plan.summary}</p>}
      {scenarios.map((s, i) => (
        <div key={i} className="rounded-lg border border-[var(--border)] overflow-hidden">
          <div className="flex items-center gap-2 px-4 py-2.5 bg-[var(--bg-subtle)] border-b border-[var(--border)]">
            <span className="text-sm font-semibold">{s.name}</span>
            <PriorityBadge priority={s.priority} />
          </div>
          <ol className="divide-y divide-[var(--border)]">
            {s.steps.map((step, j) => (
              <li key={j} className="flex items-start gap-3 px-4 py-2.5">
                <span className="w-5 h-5 rounded-full bg-[var(--bg-subtle)] border border-[var(--border)] flex items-center justify-center text-[10px] font-semibold text-[var(--text-muted)] shrink-0 mt-0.5">{j + 1}</span>
                <span className="text-[13px] text-[var(--text-secondary)]">{step}</span>
              </li>
            ))}
          </ol>
        </div>
      ))}
    </div>
  );
}

function FilesView({ run }: { run: TestRun }) {
  const files = run.test_files || [];
  if (files.length === 0) return <EmptyState icon={<FileCode className="w-6 h-6" />} title="No files generated" />;
  return (
    <div className="space-y-2">
      {files.map((f, i) => (
        <details key={i} className="rounded-lg border border-[var(--border)] overflow-hidden">
          <summary className="flex items-center gap-2 px-4 py-2.5 cursor-pointer text-xs font-mono text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]">
            <FileCode className="w-3.5 h-3.5 text-[var(--text-muted)]" /> {f.name}
          </summary>
          <pre className="px-4 py-3 text-[11px] leading-relaxed overflow-x-auto bg-[var(--bg-subtle)] border-t border-[var(--border)] text-[var(--text-secondary)]">{f.content}</pre>
        </details>
      ))}
    </div>
  );
}

function FailuresView({ run }: { run: TestRun }) {
  const failures = run.run_result?.failures || [];
  if (failures.length === 0) return <EmptyState icon={<CheckCircle2 className="w-6 h-6" />} title="No failures" description="All tests passed or the run hasn't completed." />;
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  return (
    <div className="space-y-3">
      {failures.map((f, i) => (
        <div key={i} className="rounded-lg border border-[var(--danger)]/15 bg-[var(--danger-bg)] p-4">
          <div className="flex items-start gap-2.5">
            <AlertTriangle className="w-4 h-4 text-[var(--danger)] mt-0.5 shrink-0" />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-[var(--text-primary)]">{f.test}</p>
              <pre className="text-[11px] text-[var(--danger)]/80 mt-1 whitespace-pre-wrap break-all">{f.message}</pre>
              {f.screenshot_url && (
                <a href={f.screenshot_url.startsWith("http") ? f.screenshot_url : `${API}${f.screenshot_url}`} target="_blank" className="inline-flex items-center gap-1 text-[11px] text-[var(--accent)] mt-2 hover:underline">
                  <ImageIcon className="w-3 h-3" /> View screenshot
                </a>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

// Visual artifacts view
// Video player for browser recording replay
function VideoPlayer({ run }: { run: TestRun }) {
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  if (!run.video_url || run.video_status === "none") {
    return (
      <EmptyState
        icon={<Film className="w-6 h-6" />}
        title="No video recording"
        description="Video recordings are captured automatically during browser test execution. Run a test with Steel Browser to generate a recording."
      />
    );
  }

  if (run.video_status === "recording") {
    return (
      <div className="flex items-center gap-2 p-4 rounded-[var(--radius-sm)] bg-[var(--warning-bg)] border border-[var(--warning)]/15">
        <div className="w-2 h-2 rounded-full bg-[var(--danger)] animate-pulse" />
        <span className="text-[12px] font-medium text-[var(--warning)]">Recording in progress...</span>
      </div>
    );
  }

  if (run.video_status === "failed") {
    return (
      <EmptyState
        icon={<Film className="w-6 h-6" />}
        title="Recording failed"
        description="The video recording could not be completed. This does not affect test results."
      />
    );
  }

  const videoSrc = run.video_url.startsWith("http") ? run.video_url : `${API}${run.video_url}`;

  return (
    <div className="space-y-3">
      {/* Player */}
      <div className="rounded-[var(--radius)] border border-[var(--border)] overflow-hidden bg-black">
        <video
          src={videoSrc}
          controls
          className="w-full aspect-video"
          preload="metadata"
        >
          Your browser does not support video playback.
        </video>
      </div>

      {/* Metadata */}
      <div className="flex items-center gap-4 text-[11px] text-[var(--text-muted)]">
        {run.video_duration && run.video_duration > 0 && (
          <span>Duration: {run.video_duration.toFixed(1)}s</span>
        )}
        {run.video_failure_marker_at && run.video_failure_marker_at > 0 && (
          <span className="text-[var(--danger)]">Failure at: {run.video_failure_marker_at.toFixed(1)}s</span>
        )}
        <a href={videoSrc} download className="text-[var(--accent)] hover:underline">Download</a>
      </div>
    </div>
  );
}

function VisualView({ artifacts }: { artifacts: VisualArtifact[] }) {
  if (artifacts.length === 0) {
    return <EmptyState icon={<Eye className="w-6 h-6" />} title="No visual artifacts" description="Visual artifacts are created when screenshots are captured. Enable visual regression for baseline comparisons." />;
  }
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  const resolveUrl = (u?: string) => u && (u.startsWith("http") ? u : `${API}${u}`);

  return (
    <div className="space-y-4">
      {artifacts.map((a, i) => (
        <div key={a.id || i} className="rounded-lg border border-[var(--border)] p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-[var(--text-primary)]">{a.step_name}</span>
            <div className="flex items-center gap-2">
              <span className={`text-xs font-bold ${a.passed ? "text-[var(--success)]" : "text-[var(--danger)]"}`}>
                {(a.similarity_score * 100).toFixed(1)}%
              </span>
              <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${a.passed ? "bg-[var(--success-bg)] text-[var(--success)]" : "bg-[var(--danger-bg)] text-[var(--danger)]"}`}>
                {a.passed ? "PASS" : "FAIL"}
              </span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-md border border-[var(--border)] bg-[var(--bg-subtle)] p-3">
              <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">Baseline</p>
              {a.baseline_url ? (
                <a href={resolveUrl(a.baseline_url)} target="_blank" className="block aspect-video rounded border border-[var(--border)] overflow-hidden">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={resolveUrl(a.baseline_url)!} alt="Baseline" className="w-full h-full object-cover" />
                </a>
              ) : (
                <div className="aspect-video rounded border border-dashed border-[var(--border-strong)] flex items-center justify-center text-[var(--text-muted)]">
                  <Eye className="w-5 h-5" />
                </div>
              )}
            </div>
            <div className="rounded-md border border-[var(--border)] bg-[var(--bg-subtle)] p-3">
              <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">Current</p>
              {a.current_url ? (
                <a href={resolveUrl(a.current_url)} target="_blank" className="block aspect-video rounded border border-[var(--border)] overflow-hidden">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={resolveUrl(a.current_url)!} alt="Current" className="w-full h-full object-cover" />
                </a>
              ) : (
                <div className="aspect-video rounded border border-dashed border-[var(--border-strong)] flex items-center justify-center text-[var(--text-muted)]">
                  <Eye className="w-5 h-5" />
                </div>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
