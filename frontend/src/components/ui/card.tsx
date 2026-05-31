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
        "rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg-card)] p-5 shadow-[var(--shadow-xs)]",
        hover && "transition-all duration-150 hover:shadow-[var(--shadow-md)] hover:border-[var(--accent-light)]",
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
  const textColor = {
    default: "text-[var(--text-primary)]",
    success: "text-[var(--success)]",
    danger: "text-[var(--danger)]",
    warning: "text-[var(--warning)]",
    info: "text-[var(--info)]",
  };
  const iconBg = {
    default: "bg-[var(--bg-subtle)] text-[var(--text-secondary)]",
    success: "bg-[var(--success-bg)] text-[var(--success)]",
    danger: "bg-[var(--danger-bg)] text-[var(--danger)]",
    warning: "bg-[var(--warning-bg)] text-[var(--warning)]",
    info: "bg-[var(--info-bg)] text-[var(--info)]",
  };

  return (
    <Card>
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">
            {label}
          </p>
          <p className={cn("text-[26px] font-bold leading-none tracking-tight", textColor[color])}>{value}</p>
          {trend && (
            <p className="text-[11px] text-[var(--text-secondary)] mt-1.5">{trend}</p>
          )}
        </div>
        {icon && (
          <div className={cn("p-2.5 rounded-[var(--radius-sm)]", iconBg[color])}>{icon}</div>
        )}
      </div>
    </Card>
  );
}
