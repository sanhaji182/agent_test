"use client";

import { cn } from "@/lib/utils";

interface PageLayoutProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

/**
 * Standard page shell: a consistent page header (title + description + actions)
 * followed by the page body. Spacing/padding is provided by the AppShell.
 */
export function PageLayout({ children, title, description, action, className }: PageLayoutProps) {
  return (
    <div className={cn("space-y-6", className)}>
      {(title || action) && (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            {title && (
              <h1 className="text-2xl font-semibold tracking-tight text-[var(--text-primary)]">{title}</h1>
            )}
            {description && (
              <p className="text-sm text-[var(--text-muted)] mt-1.5">{description}</p>
            )}
          </div>
          {action && <div className="shrink-0">{action}</div>}
        </div>
      )}
      
      <div>{children}</div>
    </div>
  );
}
