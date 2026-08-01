"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { logout, isAdmin, getUserRole } from "@/lib/api";
import {
  	LayoutDashboard, PlayCircle, Settings, Zap,
  	Calendar, Bell, Tag, Shield, ClipboardCheck, Layers, Download, BookOpen, FileText,
  	LogOut, Video, KeyRound
  } from "lucide-react";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/tests", label: "Test Library", icon: FileText },
  { href: "/runs", label: "Runs", icon: PlayCircle },
  { href: "/recordings", label: "Recordings", icon: Video },
  { href: "/risk", label: "Risk", icon: Shield },
  { href: "/monitoring", label: "Monitoring", icon: Calendar },
  { href: "/releases", label: "Releases", icon: Tag },
  { href: "/suites", label: "Suites", icon: Layers },
  { href: "/reviews", label: "Reviews", icon: ClipboardCheck },
  { href: "/alerts", label: "Alerts", icon: Bell },
  { href: "/exports", label: "Exports", icon: Download },
  { href: "/docs", label: "Docs", icon: BookOpen },
  { href: "/settings", label: "Settings", icon: Settings },
  ...(typeof window !== "undefined" && isAdmin() ? [
    { href: "/keys", label: "Keys & Audit", icon: KeyRound },
  ] : []),
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = async () => {
    await logout();
    router.push("/login");
  };

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

      <div className="px-3 pt-4 pb-2">
        <Link
          href="/create"
          className="flex items-center justify-center gap-2 w-full py-2 bg-[var(--accent)] text-white text-[13px] font-semibold rounded-[var(--radius-sm)] hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
        >
          <PlayCircle className="w-4 h-4" />
          Create Test
        </Link>
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
        <button onClick={handleLogout} className="w-full flex items-center gap-2.5 px-2.5 py-[7px] rounded-[var(--radius-sm)] text-[12px] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] transition-all duration-100 text-left">
          <LogOut className="w-[15px] h-[15px]" />
          Sign out
        </button>
      </nav>

      <div className="px-2 py-2 border-t border-[var(--border)]">
        <div className="px-2.5 py-2 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] space-y-1.5">
          <div className="flex items-center gap-1.5">
            <span className={cn(
              "px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider",
              getUserRole() === "admin" ? "bg-[var(--danger-bg)] text-[var(--danger)]" :
              getUserRole() === "reviewer" ? "bg-[var(--warning-bg)] text-[var(--warning)]" :
              "bg-[var(--info-bg)] text-[var(--info)]"
            )}>
              {getUserRole()}
            </span>
          </div>
          <p className="text-[9px] font-semibold text-[var(--text-muted)] uppercase tracking-wider">Plan · Self-Hosted</p>
        </div>
      </div>
    </aside>
  );
}
