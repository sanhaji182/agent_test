import { cn } from "@/lib/utils";

export function Card({
  children,
  className,
  hover = false,
}: {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}) {
  return (
    <div
      className={cn(
        "rounded-xl border border-[var(--border)] bg-[var(--bg-card)] p-5",
        hover && "transition-colors hover:bg-[var(--bg-card-hover)] hover:border-[var(--border-subtle)]",
        className
      )}
    >
      {children}
    </div>
  );
}

export function StatCard({
  label,
  value,
  icon,
  trend,
  color = "default",
}: {
  label: string;
  value: string | number;
  icon?: React.ReactNode;
  trend?: string;
  color?: "default" | "success" | "danger" | "warning" | "info";
}) {
  const colors = {
    default: "text-[var(--text-primary)]",
    success: "text-[var(--success)]",
    danger: "text-[var(--danger)]",
    warning: "text-[var(--warning)]",
    info: "text-[var(--info)]",
  };

  return (
    <Card>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-[var(--text-muted)] mb-1">
            {label}
          </p>
          <p className={cn("text-2xl font-bold", colors[color])}>{value}</p>
          {trend && (
            <p className="text-xs text-[var(--text-secondary)] mt-1">{trend}</p>
          )}
        </div>
        {icon && (
          <div className="p-2 rounded-lg bg-[var(--bg-secondary)] text-[var(--text-secondary)]">
            {icon}
          </div>
        )}
      </div>
    </Card>
  );
}
