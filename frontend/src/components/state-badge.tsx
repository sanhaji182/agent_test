import { cn } from "@/lib/utils";

const stateColors: Record<string, string> = {
  idle: "bg-zinc-100 text-zinc-700",
  analyzing: "bg-blue-100 text-blue-700",
  plan_generated: "bg-indigo-100 text-indigo-700",
  writing_tests: "bg-purple-100 text-purple-700",
  running: "bg-yellow-100 text-yellow-700",
  fixing: "bg-orange-100 text-orange-700",
  done: "bg-green-100 text-green-700",
  failed: "bg-red-100 text-red-700",
};

export function StateBadge({ state }: { state: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2 py-0.5 rounded text-xs font-medium",
        stateColors[state] || "bg-zinc-100 text-zinc-700"
      )}
    >
      {state}
    </span>
  );
}
