"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard, PlayCircle, FolderOpen, Settings, Zap,
  Calendar, Bell, Tag, Shield, ClipboardCheck, Layers,
} from "lucide-react";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/runs", label: "Test Runs", icon: PlayCircle },
  { href: "/monitoring", label: "Schedules", icon: Calendar },
  { href: "/risk", label: "Risk & Intel", icon: Shield },
  { href: "/reviews", label: "Reviews", icon: ClipboardCheck },
  { href: "/suites", label: "Suites", icon: Layers },
  { href: "/releases", label: "Releases", icon: Tag },
  { href: "/alerts", label: "Alerts", icon: Bell },
  { href: "/projects", label: "Projects", icon: FolderOpen },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-[220px] min-h-screen border-r border-[var(--border)] bg-[var(--bg-card)] flex flex-col shadow-[var(--shadow-xs)]">
      <div className="h-[52px] px-4 border-b border-[var(--border)] flex items-center">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-[var(--radius-sm)] bg-gradient-to-br from-indigo-500 to-indigo-600 flex items-center justify-center shadow-sm">
            <Zap className="w-3.5 h-3.5 text-white" />
          </div>
          <span className="text-[13px] font-bold text-[var(--text-primary)]">GoTest</span>
        </div>
      </div>

      <nav className="flex-1 px-2 py-2 space-y-0.5 overflow-y-auto">
        {nav.map((item) => {
          const active = item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-2.5 px-2.5 py-[7px] rounded-[var(--radius-sm)] text-[12px] font-medium transition-all duration-100",
                active
                  ? "bg-[var(--accent-bg)] text-[var(--accent)] shadow-[var(--shadow-xs)]"
                  : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
              )}
            >
              <item.icon className="w-[15px] h-[15px]" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="px-2 py-2 border-t border-[var(--border)]">
        <div className="px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)]">
          <p className="text-[9px] font-semibold text-[var(--text-muted)] uppercase tracking-wider">Plan</p>
          <p className="text-[11px] font-semibold text-[var(--text-primary)]">Self-Hosted</p>
        </div>
      </div>
    </aside>
  );
}
