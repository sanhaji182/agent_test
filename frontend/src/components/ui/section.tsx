import { cn } from "@/lib/utils";

export function EmptyState({
  icon,
  title,
  description,
}: {
  icon?: React.ReactNode;
  title: string;
  description?: string;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4">
      {icon && (
        <div className="w-14 h-14 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--border)] flex items-center justify-center mb-4 text-[var(--text-muted)]">
          {icon}
        </div>
      )}
      <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-1">{title}</h3>
      {description && (
        <p className="text-xs text-[var(--text-muted)] max-w-xs text-center">{description}</p>
      )}
    </div>
  );
}

export function Section({
  title,
  children,
  action,
  className,
}: {
  title: string;
  children: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("rounded-xl border border-[var(--border)] bg-[var(--bg-card)] overflow-hidden", className)}>
      <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border)]">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)]">
          {title}
        </h2>
        {action}
      </div>
      <div className="p-5">{children}</div>
    </section>
  );
}

export function LoadingSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-10 rounded-lg bg-[var(--bg-secondary)] animate-pulse" />
      ))}
    </div>
  );
}
