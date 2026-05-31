"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  PlayCircle,
  FolderOpen,
  Settings,
  Zap,
} from "lucide-react";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/runs", label: "Test Runs", icon: PlayCircle },
  { href: "/projects", label: "Projects", icon: FolderOpen },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-[240px] min-h-screen border-r border-[var(--border)] bg-[var(--bg-card)] flex flex-col">
      {/* Logo */}
      <div className="h-14 px-5 border-b border-[var(--border)] flex items-center">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-indigo-600 flex items-center justify-center shadow-sm">
            <Zap className="w-4 h-4 text-white" />
          </div>
          <div>
            <h1 className="text-sm font-bold text-[var(--text-primary)] leading-tight">GoTest Agent</h1>
            <p className="text-[10px] text-[var(--text-muted)]">AI Testing Platform</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-1">
        {nav.map((item) => {
          const active =
            item.href === "/"
              ? pathname === "/"
              : pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13px] font-medium transition-colors",
                active
                  ? "bg-[var(--accent-bg)] text-[var(--accent)]"
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
              )}
            >
              <item.icon className="h-[18px] w-[18px]" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="p-3 border-t border-[var(--border)]">
        <div className="px-3 py-2.5 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
          <p className="text-[10px] font-medium text-[var(--text-muted)] uppercase tracking-wider">Plan</p>
          <p className="text-xs font-semibold text-[var(--text-primary)]">Self-Hosted</p>
        </div>
      </div>
    </aside>
  );
}
