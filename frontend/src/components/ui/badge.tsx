import { cn } from "@/lib/utils";

const variants: Record<string, string> = {
  idle: "bg-[var(--bg-secondary)] text-[var(--text-secondary)] border-[var(--border)]",
  analyzing: "bg-[var(--info-bg)] text-[var(--info)] border-[var(--info)]/20",
  plan_generated: "bg-purple-500/10 text-purple-400 border-purple-500/20",
  writing_tests: "bg-violet-500/10 text-violet-400 border-violet-500/20",
  running: "bg-[var(--warning-bg)] text-[var(--warning)] border-[var(--warning)]/20 animate-pulse",
  fixing: "bg-orange-500/10 text-orange-400 border-orange-500/20",
  done: "bg-[var(--success-bg)] text-[var(--success)] border-[var(--success)]/20",
  failed: "bg-[var(--danger-bg)] text-[var(--danger)] border-[var(--danger)]/20",
};

export function StatusBadge({ state }: { state: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-semibold uppercase tracking-wide border",
        variants[state] || variants.idle
      )}
    >
      <span
        className={cn(
          "w-1.5 h-1.5 rounded-full",
          state === "done" && "bg-[var(--success)]",
          state === "failed" && "bg-[var(--danger)]",
          state === "running" && "bg-[var(--warning)]",
          !["done", "failed", "running"].includes(state) && "bg-current"
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
    low: "bg-[var(--bg-secondary)] text-[var(--text-secondary)]",
  };
  return (
    <span className={cn("px-2 py-0.5 rounded text-[10px] font-semibold uppercase", colors[priority] || colors.low)}>
      {priority}
    </span>
  );
}
