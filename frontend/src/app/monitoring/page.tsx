"use client";

import { useEffect, useState } from "react";
import { getSchedules, runScheduleNow, updateSchedule, type Schedule } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { StatusBadge } from "@/components/ui/badge";
import { Calendar, PlayCircle, Pause, Play } from "lucide-react";

export default function MonitoringPage() {
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getSchedules().then(setSchedules).catch(() => {}).finally(() => setLoading(false));
  }, []);

  const toggleEnabled = async (id: string, enabled: boolean) => {
    await updateSchedule(id, { enabled: !enabled } as Partial<Schedule>);
    setSchedules((prev) => prev.map((s) => (s.id === id ? { ...s, enabled: !enabled } : s)));
  };

  const triggerRun = async (id: string) => {
    await runScheduleNow(id);
    getSchedules().then(setSchedules);
  };

  if (loading) return <div className="space-y-6"><LoadingSkeleton rows={5} /></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Monitoring</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Recurring test schedules that run automatically. Set up once, get notified on failures.</p>
      </div>

      <Section title="Schedules" action={<span className="text-[11px] text-[var(--text-muted)]">{schedules.length} configured</span>}>
        {schedules.length === 0 ? (
          <EmptyState
            icon={<Calendar className="w-6 h-6" />}
            title="No schedules configured"
            description="Create a schedule via POST /api/v1/schedules to enable recurring test runs."
          />
        ) : (
          <div className="space-y-2">
            {schedules.map((sch) => (
              <div key={sch.id} className="flex items-center gap-4 p-4 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold">{sch.name || "Unnamed"}</span>
                    <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${sch.enabled ? "bg-[var(--success-bg)] text-[var(--success)]" : "bg-[var(--bg-subtle)] text-[var(--text-muted)]"}`}>
                      {sch.enabled ? "Active" : "Paused"}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-[var(--info-bg)] text-[var(--info)]">{sch.frequency}</span>
                    {sch.environment && (
                      <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-[var(--accent-bg)] text-[var(--accent)]">{sch.environment}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 mt-1 text-[11px] text-[var(--text-muted)]">
                    <span>Next: {new Date(sch.next_run_at).toLocaleString()}</span>
                    {sch.last_run_status && <StatusBadge state={sch.last_run_status} />}
                  </div>
                </div>
                <div className="flex items-center gap-1.5">
                  <button onClick={() => triggerRun(sch.id)} className="p-2 rounded-lg hover:bg-[var(--bg-hover)] text-[var(--accent)]" title="Run now">
                    <PlayCircle className="w-4 h-4" />
                  </button>
                  <button onClick={() => toggleEnabled(sch.id, sch.enabled)} className="p-2 rounded-lg hover:bg-[var(--bg-hover)] text-[var(--text-secondary)]" title={sch.enabled ? "Pause" : "Resume"}>
                    {sch.enabled ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}
