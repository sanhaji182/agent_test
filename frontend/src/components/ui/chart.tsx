"use client";

// Lightweight CSS bar chart for trend visualization
export function TrendChart({
  data,
  label,
}: {
  data: { label: string; value: number; max?: number }[];
  label: string;
}) {
  if (data.length === 0) return null;
  const maxVal = Math.max(...data.map((d) => d.max || d.value), 1);

  return (
    <div>
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">{label}</p>
      <div className="flex items-end gap-1 h-16">
        {data.map((d, i) => {
          const height = Math.max((d.value / maxVal) * 100, 4);
          const color = d.value >= maxVal * 0.8 ? "bg-[var(--success)]" : d.value >= maxVal * 0.5 ? "bg-[var(--warning)]" : "bg-[var(--danger)]";
          return (
            <div key={i} className="flex-1 flex flex-col items-center gap-0.5" title={`${d.label}: ${(d.value * 100).toFixed(0)}%`}>
              <div className={`w-full rounded-sm ${color} transition-all`} style={{ height: `${height}%` }} />
              <span className="text-[8px] text-[var(--text-muted)] truncate w-full text-center">{d.label}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// Pass rate mini chart from trend data
export function PassRateChart({ trend }: { trend: { date: string; pass_rate: number }[] }) {
  if (trend.length === 0) return null;
  const data = trend.slice(-7).map((t) => ({
    label: t.date.slice(5), // MM-DD
    value: t.pass_rate,
    max: 1,
  }));
  return <TrendChart data={data} label="Pass Rate (7d)" />;
}
