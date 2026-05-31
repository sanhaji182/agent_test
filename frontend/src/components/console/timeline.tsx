import { PHASES } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Check, Loader2, Circle, X } from "lucide-react";

const LABELS: Record<string, string> = {
  analyzing: "Analyze",
  plan_generated: "Plan",
  writing_tests: "Write",
  running: "Run",
  fixing: "Fix",
  done: "Done",
};

// Timeline fase eksekusi. Status fase diturunkan dari state run saat ini.
export function ExecutionTimeline({ state }: { state: string }) {
  const failed = state === "failed";
  const currentIdx = PHASES.indexOf(state as (typeof PHASES)[number]);

  return (
    <div className="flex items-center gap-1 overflow-x-auto">
      {PHASES.map((phase, i) => {
        const done = currentIdx > i || state === "done";
        const active = state === phase && state !== "done";
        const isLast = i === PHASES.length - 1;

        return (
          <div key={phase} className="flex items-center">
            <div className="flex flex-col items-center gap-1.5 min-w-[58px]">
              <div
                className={cn(
                  "w-7 h-7 rounded-full flex items-center justify-center border-2",
                  done && "bg-[var(--success)] border-[var(--success)] text-white",
                  active && "border-[var(--warning)] text-[var(--warning)]",
                  failed && active && "border-[var(--danger)] text-[var(--danger)]",
                  !done && !active && "border-[var(--border-strong)] text-[var(--text-muted)]"
                )}
              >
                {done ? (
                  <Check className="w-3.5 h-3.5" />
                ) : active && failed ? (
                  <X className="w-3.5 h-3.5" />
                ) : active ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <Circle className="w-2 h-2 fill-current" />
                )}
              </div>
              <span
                className={cn(
                  "text-[10px] font-medium",
                  done || active ? "text-[var(--text-primary)]" : "text-[var(--text-muted)]"
                )}
              >
                {LABELS[phase]}
              </span>
            </div>
            {!isLast && (
              <div
                className={cn(
                  "h-0.5 w-6 -mt-5",
                  currentIdx > i || state === "done"
                    ? "bg-[var(--success)]"
                    : "bg-[var(--border-strong)]"
                )}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}
