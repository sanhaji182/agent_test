import { cn } from "@/lib/utils";

export function EmptyState({
  icon,
  title,
  description,
  action,
}: {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-14 px-4">
      {icon && (
        <div className="w-12 h-12 rounded-[var(--radius)] bg-[var(--bg-subtle)] border border-[var(--border)] flex items-center justify-center mb-3 text-[var(--text-muted)]">
          {icon}
        </div>
      )}
      <h3 className="text-[13px] font-semibold text-[var(--text-primary)] mb-0.5">{title}</h3>
      {description && (
        <p className="text-[11px] text-[var(--text-muted)] max-w-[280px] text-center leading-relaxed">{description}</p>
      )}
      {action && <div className="mt-3">{action}</div>}
    </div>
  );
}

export function Section({
  title,
  children,
  action,
  className,
  noPadding,
}: {
  title: string;
  children: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
  noPadding?: boolean;
}) {
  return (
    <section className={cn("rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-xs)] overflow-hidden", className)}>
      <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border)]">
        <h2 className="text-[13px] font-semibold text-[var(--text-primary)]">{title}</h2>
        {action}
      </div>
      <div className={noPadding ? "" : "p-5"}>{children}</div>
    </section>
  );
}

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-2.5">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-9 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] animate-pulse" style={{ width: `${100 - i * 8}%` }} />
      ))}
    </div>
  );
}
