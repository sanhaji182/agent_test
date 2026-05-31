import { cn } from "@/lib/utils";

const variants: Record<string, string> = {
  idle: "bg-[var(--bg-subtle)] text-[var(--text-secondary)] border-[var(--border)]",
  analyzing: "bg-[var(--info-bg)] text-[var(--info)] border-[var(--info)]/20",
  plan_generated: "bg-[var(--accent-bg)] text-[var(--accent)] border-[var(--accent)]/20",
  writing_tests: "bg-[var(--accent-bg)] text-[var(--accent)] border-[var(--accent)]/20",
  running: "bg-[var(--warning-bg)] text-[var(--warning)] border-[var(--warning)]/20",
  fixing: "bg-[var(--warning-bg)] text-[var(--warning)] border-[var(--warning)]/20",
  done: "bg-[var(--success-bg)] text-[var(--success)] border-[var(--success)]/20",
  failed: "bg-[var(--danger-bg)] text-[var(--danger)] border-[var(--danger)]/20",
};

export function StatusBadge({ state }: { state: string }) {
  const dotColor: Record<string, string> = {
    done: "bg-[var(--success)]",
    failed: "bg-[var(--danger)]",
    running: "bg-[var(--warning)]",
  };
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-semibold capitalize border",
        variants[state] || variants.idle
      )}
    >
      <span
        className={cn(
          "w-1.5 h-1.5 rounded-full",
          dotColor[state] || "bg-current",
          state === "running" && "animate-pulse"
        )}
      />
      {state.replace("_", " ")}
    </span>
  );
}

export function PriorityBadge({ priority }: { priority: string }) {
  const colors: Record<string, string> = {
    high: "bg-[var(--danger-bg)] text-[var(--danger)]",
    medium: "bg-[var(--warning-bg)] text-[var(--warning)]",
    low: "bg-[var(--bg-subtle)] text-[var(--text-secondary)]",
  };
  return (
    <span className={cn("px-2 py-0.5 rounded text-[10px] font-semibold uppercase", colors[priority] || colors.low)}>
      {priority}
    </span>
  );
}
