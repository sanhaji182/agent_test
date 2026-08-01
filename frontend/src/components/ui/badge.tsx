import { cn } from "@/lib/utils";

const variants: Record<string, string> = {
  idle: "bg-[var(--bg-subtle)] text-[var(--text-muted)] border-[var(--border)]",
  analyzing: "bg-[var(--info-bg)] text-[var(--info)] border-[var(--info)]/15",
  plan_generated: "bg-[var(--accent-bg)] text-[var(--accent)] border-[var(--accent)]/15",
  writing_tests: "bg-[var(--accent-bg)] text-[var(--accent)] border-[var(--accent)]/15",
  running: "bg-[var(--warning-bg)] text-[var(--warning)] border-[var(--warning)]/15",
  fixing: "bg-[var(--warning-bg)] text-[var(--warning)] border-[var(--warning)]/15",
  done: "bg-[var(--success-bg)] text-[var(--success)] border-[var(--success)]/15",
  failed: "bg-[var(--danger-bg)] text-[var(--danger)] border-[var(--danger)]/15",
  recording: "bg-[var(--info-bg)] text-[var(--info)] border-[var(--info)]/15",
  completed: "bg-[var(--success-bg)] text-[var(--success)] border-[var(--success)]/15",
  aborted: "bg-[var(--bg-subtle)] text-[var(--text-muted)] border-[var(--border)]",
  generating: "bg-[var(--accent-bg)] text-[var(--accent)] border-[var(--accent)]/15",
};

export function StatusBadge({ state }: { state: string }) {
  const dotColor: Record<string, string> = {
    done: "bg-[var(--success)]",
    failed: "bg-[var(--danger)]",
    running: "bg-[var(--warning)] animate-pulse",
    completed: "bg-[var(--success)]",
    recording: "bg-[var(--info)] animate-pulse",
    aborted: "bg-[var(--text-muted)]",
    generating: "bg-[var(--accent)] animate-pulse",
  };
  return (
    <span className={cn("inline-flex items-center gap-1.5 px-2 py-[3px] rounded-[var(--radius-sm)] text-[10px] font-semibold capitalize border", variants[state] || variants.idle)}>
      <span className={cn("w-[5px] h-[5px] rounded-full", dotColor[state] || "bg-current")} />
      {state.replace("_", " ")}
    </span>
  );
}

export function PriorityBadge({ priority }: { priority: string }) {
  const colors: Record<string, string> = {
    high: "bg-[var(--danger-bg)] text-[var(--danger)]",
    medium: "bg-[var(--warning-bg)] text-[var(--warning)]",
    low: "bg-[var(--bg-subtle)] text-[var(--text-muted)]",
  };
  return (
    <span className={cn("px-1.5 py-[2px] rounded text-[9px] font-bold uppercase tracking-wide", colors[priority] || colors.low)}>
      {priority}
    </span>
  );
}
