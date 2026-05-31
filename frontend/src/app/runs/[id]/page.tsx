"use client";

import React, { useEffect, useState } from "react";
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
          { id: "video", label: "Video", content: <VideoPlayer run={run} events={liveEvents} /> },
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
// Video player with failure highlight, step markers, and inspection controls
function VideoPlayer({ run, events: runEvents }: { run: TestRun; events: RunEvent[] }) {
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  const videoRef = React.useRef<HTMLVideoElement>(null);
  const [currentTime, setCurrentTime] = React.useState(0);

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
  const hasFailure = run.run_result && run.run_result.failed > 0;
  const failureAt = run.video_failure_marker_at || 0;
  const duration = run.video_duration || 0;

  // Use precise step timestamps from events if available, otherwise approximate
  const stepMarkers = React.useMemo(() => {
    // Try precise timestamps from step_started events
    const preciseMarkers = runEvents
      .filter((e) => e.type === "step_started" && e.metadata?.timestamp_ms)
      .map((e) => ({
        label: e.metadata?.step || e.message,
        time: parseInt(e.metadata!.timestamp_ms!) / 1000,
        precise: true,
      }));

    if (preciseMarkers.length > 0) return preciseMarkers;

    // Fallback: approximate from test plan
    const steps = run.test_plan?.scenarios?.flatMap((s) => s.steps) || [];
    if (duration > 0 && steps.length > 0) {
      return steps.map((step, i) => ({ label: step, time: (duration / (steps.length + 1)) * (i + 1), precise: false }));
    }
    return [];
  }, [runEvents, run.test_plan, duration]);

  const seekTo = (time: number) => {
    if (videoRef.current) {
      videoRef.current.currentTime = time;
      videoRef.current.play();
    }
  };

  return (
    <div className="space-y-3">
      {/* Failure callout */}
      {hasFailure && (
        <div className="flex items-center justify-between p-3 rounded-[var(--radius-sm)] bg-[var(--danger-bg)] border border-[var(--danger)]/15">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-[var(--danger)]" />
            <span className="text-[12px] font-medium text-[var(--danger)]">
              {run.run_result!.failures[0]?.test}: {run.run_result!.failures[0]?.message?.slice(0, 80)}
            </span>
          </div>
          {failureAt > 0 && (
            <button
              onClick={() => seekTo(failureAt)}
              className="inline-flex items-center gap-1 px-2.5 py-1 rounded-[var(--radius-sm)] bg-[var(--danger)] text-white text-[10px] font-semibold hover:opacity-90 transition-opacity shrink-0"
            >
              Jump to failure ({failureAt.toFixed(1)}s)
            </button>
          )}
        </div>
      )}

      {/* Player */}
      <div className="rounded-[var(--radius)] border border-[var(--border)] overflow-hidden bg-black relative">
        <video
          ref={videoRef}
          src={videoSrc}
          controls
          className="w-full aspect-video"
          preload="metadata"
          onTimeUpdate={(e) => setCurrentTime((e.target as HTMLVideoElement).currentTime)}
        >
          Your browser does not support video playback.
        </video>
      </div>

      {/* Timeline scrubber with markers */}
      {duration > 0 && (
        <div className="relative h-6 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] border border-[var(--border)] overflow-hidden">
          {/* Progress */}
          <div className="absolute inset-y-0 left-0 bg-[var(--accent)]/10" style={{ width: `${(currentTime / duration) * 100}%` }} />
          {/* Failure marker */}
          {failureAt > 0 && (
            <button
              onClick={() => seekTo(failureAt)}
              className="absolute top-0 bottom-0 w-1 bg-[var(--danger)] hover:w-1.5 transition-all cursor-pointer z-10"
              style={{ left: `${(failureAt / duration) * 100}%` }}
              title={`Failure at ${failureAt.toFixed(1)}s`}
            />
          )}
          {/* Step markers */}
          {stepMarkers.map((m, i) => (
            <button
              key={i}
              onClick={() => seekTo(m.time)}
              className="absolute top-1 bottom-1 w-0.5 bg-[var(--accent)]/40 hover:bg-[var(--accent)] transition-colors cursor-pointer"
              style={{ left: `${(m.time / duration) * 100}%` }}
              title={`Step: ${m.label} (~${m.time.toFixed(0)}s)`}
            />
          ))}
          {/* Current position */}
          <div className="absolute top-0 bottom-0 w-0.5 bg-[var(--text-primary)]" style={{ left: `${(currentTime / duration) * 100}%` }} />
        </div>
      )}

      {/* Step chips */}
      {stepMarkers.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {stepMarkers.map((m, i) => {
            const isActive = currentTime >= m.time - 1 && currentTime < (stepMarkers[i + 1]?.time || duration);
            return (
              <button
                key={i}
                onClick={() => seekTo(m.time)}
                className={`px-2 py-1 rounded-[var(--radius-sm)] text-[10px] font-medium border transition-all ${
                  isActive
                    ? "bg-[var(--accent-bg)] border-[var(--accent)]/30 text-[var(--accent)]"
                    : "bg-[var(--bg-subtle)] border-[var(--border)] text-[var(--text-muted)] hover:text-[var(--text-secondary)]"
                }`}
              >
                {m.label.length > 25 ? m.label.slice(0, 25) + "…" : m.label}
              </button>
            );
          })}
        </div>
      )}

      {/* Controls bar */}
      <div className="flex items-center gap-4 text-[11px] text-[var(--text-muted)]">
        {duration > 0 && <span>{currentTime.toFixed(1)}s / {duration.toFixed(1)}s</span>}
        {stepMarkers.length > 0 && <span className="text-[var(--text-muted)]">{stepMarkers.length} steps ({stepMarkers[0]?.precise ? "precise" : "approximate"})</span>}
        <a href={videoSrc} download className="ml-auto text-[var(--accent)] hover:underline">Download video</a>
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
