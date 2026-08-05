"use client";

import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useSidebar } from "@/components/layout/sidebar-context";
import { logout, isAdmin } from "@/lib/api";
import { 
  LayoutDashboard, PlayCircle, Settings, Zap, Calendar, Bell, Shield, 
  ClipboardCheck, Layers, Download, BookOpen, KeyRound, FileText, Video,
  ChevronLeft, ChevronRight, Home, Users
} from "lucide-react";

interface NavItem {
  href: string;
  label: string;
  icon: React.ElementType;
  adminOnly?: boolean;
}

const coreNav: NavItem[] = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/projects", label: "Projects", icon: Home },
  { href: "/tests", label: "Tests", icon: FileText },
  { href: "/runs", label: "Runs", icon: PlayCircle },
  { href: "/recordings", label: "Recordings", icon: Video },
  { href: "/risk", label: "Risk Analysis", icon: Shield },
  { href: "/monitoring", label: "Monitoring", icon: Calendar },
  { href: "/releases", label: "Releases", icon: Layers },
  { href: "/reviews", label: "Reviews", icon: ClipboardCheck },
  { href: "/alerts", label: "Alerts", icon: Bell },
  { href: "/exports", label: "Exports", icon: Download },
  { href: "/docs", label: "Documentation", icon: BookOpen },
  { href: "/settings", label: "Settings", icon: Settings },
  { href: "/users", label: "Users", icon: Users, adminOnly: true },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { collapsed, toggle } = useSidebar();
  const [isAdminUser] = useState<boolean>(() => isAdmin());
  
  const allItems = coreNav.filter(item => isAdminUser || !item.adminOnly);

  const handleLogout = async () => {
    await logout();
    router.push("/login");
  };
  
  const isActive = (href: string) => {
    if (href === "/") return pathname === "/";
    return pathname.startsWith(href);
  };
  
  return (
    <aside className={cn(
      "fixed left-0 top-0 z-40 h-screen bg-white border-r border-[var(--border-default)] flex flex-col transition-all duration-300 ease-in-out",
      collapsed ? "w-16" : "w-60"
    )}>
      {/* Header */}
      <div className="flex items-center justify-between h-16 px-4 border-b border-[var(--border-default)] shrink-0">
        <Link href="/" className={cn("flex items-center gap-3", collapsed && "justify-center w-full")}>
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center flex-shrink-0 shadow-sm">
            <Zap className="w-4 h-4 text-white" />
          </div>
          {!collapsed && (
            <span className="font-semibold text-[15px] text-[var(--text-primary)] tracking-tight">GoTest Agent</span>
          )}
        </Link>
        
	        {/* Collapse/Expand Toggle */}
	        <button
	          onClick={toggle}
	          className="p-1.5 rounded-md hover:bg-[var(--bg-hover)] text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
	          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
	        >
	          {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
	        </button>
      </div>
      
      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto scrollbar-thin py-4">
        <div className="px-3 space-y-0.5">
          {allItems.map((item) => {
            const active = isActive(item.href);
            const Icon = item.icon;
            
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "relative flex items-center gap-3 rounded-md text-sm font-medium transition-colors",
                  collapsed ? "justify-center px-2 py-2.5" : "px-3 py-2.5",
                  active
                    ? "bg-[var(--accent-light)] text-[var(--accent)]"
                    : "text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]"
                )}
              >
                {active && !collapsed && (
                  <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-[var(--accent)]" />
                )}
                <Icon className={cn("w-[18px] h-[18px] flex-shrink-0", collapsed && "mx-auto")} />
                {!collapsed && <span className="truncate">{item.label}</span>}
              </Link>
            );
          })}
        </div>
      </nav>
      
      {/* Bottom Actions */}
      <div className="p-3 border-t border-[var(--border-default)] shrink-0">
        <button
          onClick={handleLogout}
          className={cn(
            "flex items-center gap-3 rounded-md text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)] transition-colors w-full",
            collapsed ? "justify-center px-2 py-2.5" : "px-3 py-2.5"
          )}
        >
          <KeyRound className={cn("w-[18px] h-[18px] flex-shrink-0", collapsed && "mx-auto")} />
          {!collapsed && <span>Sign out</span>}
        </button>
      </div>
    </aside>
  );
}
