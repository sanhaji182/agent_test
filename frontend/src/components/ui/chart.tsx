"use client";

// Premium mini bar chart for trend visualization
export function TrendChart({
  data,
  label,
  height = 56,
}: {
  data: { label: string; value: number; max?: number }[];
  label: string;
  height?: number;
}) {
  if (data.length === 0) return null;
  const maxVal = Math.max(...data.map((d) => d.max || d.value), 0.01);

  return (
    <div>
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-3">{label}</p>
      <div className="flex items-end gap-[3px]" style={{ height }}>
        {data.map((d, i) => {
          const pct = Math.max((d.value / maxVal) * 100, 3);
          const color =
            d.value >= maxVal * 0.8 ? "bg-[var(--success)]" :
            d.value >= maxVal * 0.5 ? "bg-[var(--accent)]" :
            d.value >= maxVal * 0.25 ? "bg-[var(--warning)]" : "bg-[var(--danger)]";
          return (
            <div key={i} className="flex-1 flex flex-col items-center justify-end h-full group relative">
              <div
                className={`w-full rounded-t-[2px] ${color} opacity-80 group-hover:opacity-100 transition-opacity`}
                style={{ height: `${pct}%` }}
              />
              {/* Tooltip */}
              <div className="absolute -top-7 left-1/2 -translate-x-1/2 hidden group-hover:block px-1.5 py-0.5 rounded bg-[var(--text-primary)] text-[var(--bg-card)] text-[9px] font-medium whitespace-nowrap z-10">
                {(d.value * 100).toFixed(0)}%
              </div>
            </div>
          );
        })}
      </div>
      <div className="flex gap-[3px] mt-1">
        {data.map((d, i) => (
          <span key={i} className="flex-1 text-[8px] text-[var(--text-muted)] text-center truncate">{d.label}</span>
        ))}
      </div>
    </div>
  );
}

export function PassRateChart({ trend }: { trend: { date: string; pass_rate: number }[] }) {
  if (trend.length === 0) return null;
  const data = trend.slice(-10).map((t) => ({
    label: t.date.slice(5),
    value: t.pass_rate,
    max: 1,
  }));
  return <TrendChart data={data} label="Pass Rate Trend" height={64} />;
}
