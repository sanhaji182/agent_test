import { cn } from "@/lib/utils";

export function Badge({ children, variant = "default", size = "sm" }: { 
  children: React.ReactNode; 
  variant?: "default" | "success" | "warning" | "danger" | "info";
  size?: "sm" | "md";
}) {
  const colors: Record<string, string> = {
    default: "bg-gray-100 text-gray-700 border-gray-200",
    success: "bg-green-100 text-green-700 border-green-200",
    warning: "bg-yellow-100 text-yellow-700 border-yellow-200",
    danger: "bg-red-100 text-red-700 border-red-200",
    info: "bg-blue-100 text-blue-700 border-blue-200",
  };
  const sizes: Record<string, string> = {
    sm: "px-1.5 py-0.5 text-[10px]",
    md: "px-2 py-1 text-xs",
  };
  return (
    <span className={cn("inline-flex items-center font-medium rounded border", colors[variant], sizes[size])}>
      {children}
    </span>
  );
}

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

export function StatusBadge({ state, size = "md" }: { state: string; size?: "sm" | "md" }) {
  const dotColor: Record<string, string> = {
    done: "bg-[var(--success)]",
    failed: "bg-[var(--danger)]",
    running: "bg-[var(--warning)] animate-pulse",
    completed: "bg-[var(--success)]",
    recording: "bg-[var(--info)] animate-pulse",
    aborted: "bg-[var(--text-muted)]",
    generating: "bg-[var(--accent)] animate-pulse",
  };
  
  const isSm = size === "sm";
  
  return (
    <span className={cn(
      "inline-flex items-center gap-1.5 font-semibold capitalize border rounded-[var(--radius-sm)]",
      variants[state] || variants.idle,
      isSm ? "px-1.5 py-0.5 text-[9px]" : "px-2 py-[3px] text-[10px]"
    )}>
      <span className={cn("rounded-full", dotColor[state] || "bg-current", isSm ? "w-[3.5px] h-[3.5px]" : "w-[5px] h-[5px]")} />
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
