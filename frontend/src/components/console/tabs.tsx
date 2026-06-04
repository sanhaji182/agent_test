"use client";

import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";

export interface Tab {
  id: string;
  label: string;
  count?: number;
  content: React.ReactNode;
}

// Tabbed panel sederhana untuk execution console
export function Tabs({ tabs, initial }: { tabs: Tab[]; initial?: string }) {
  const [active, setActive] = useState(initial || tabs[0]?.id);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (initial) setActive(initial);
  }, [initial]);

  const activeTab = tabs.find((t) => t.id === active);

  return (
    <div>
      <div className="flex items-center gap-1 border-b border-[var(--border)] px-1">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActive(tab.id)}
            className={cn(
              "px-3 py-2.5 text-[13px] font-medium border-b-2 -mb-px transition-colors",
              active === tab.id
                ? "border-[var(--accent)] text-[var(--accent)]"
                : "border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            )}
          >
            {tab.label}
            {tab.count !== undefined && (
              <span
                className={cn(
                  "ml-1.5 px-1.5 py-0.5 rounded text-[10px] font-semibold",
                  active === tab.id
                    ? "bg-[var(--accent-bg)] text-[var(--accent)]"
                    : "bg-[var(--bg-subtle)] text-[var(--text-muted)]"
                )}
              >
                {tab.count}
              </span>
            )}
          </button>
        ))}
      </div>
      <div className="pt-4">{activeTab?.content}</div>
    </div>
  );
}
