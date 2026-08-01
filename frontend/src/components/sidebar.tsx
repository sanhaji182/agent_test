"use client";

import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";
import { logout, isAdmin } from "@/lib/api";
import { 
  LayoutDashboard, PlayCircle, Settings, Zap, Calendar, Bell, Shield, 
  ClipboardCheck, Layers, Download, BookOpen, KeyRound, FileText, Video,
  ChevronLeft, Home, AlertTriangle
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
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const [collapsed, setCollapsed] = useState(false);
  const [mounted, setMounted] = useState(false);
  
  useEffect(() => {
    setMounted(true);
  }, []);
  
  const adminItems = coreNav.filter(item => !item.adminOnly);
  const extraItems = coreNav.filter(item => item.adminOnly);
  const allItems = [...adminItems, ...extraItems].filter(item => {
    if (!mounted) return false;
    return isAdmin() ? true : !item.adminOnly;
  });

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
      "fixed left-0 top-0 z-40 h-screen bg-white border-r border-[var(--border-default)] transition-all duration-300 ease-in-out",
      collapsed ? "w-16" : "w-60"
    )}>
      {/* Header */}
      <div className="flex items-center justify-between h-14 px-4 border-b border-[var(--border-default)]">
        <Link href="/" className="flex items-center gap-3">
          <div className="w-7 h-7 rounded-md bg-blue-600 flex items-center justify-center flex-shrink-0">
            <Zap className="w-4 h-4 text-white" />
          </div>
          {!collapsed && (
            <span className="font-semibold text-sm text-gray-900">GoTest Agent</span>
          )}
        </Link>
        
        {/* Collapse Toggle */}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-1.5 rounded-md hover:bg-gray-100 text-gray-500"
        >
          <ChevronLeft className={`w-4 h-4 transition-transform ${collapsed ? "rotate-180" : ""}`} />
        </button>
      </div>
      
      {/* Navigation */}
      <nav className="py-3 overflow-y-auto scrollbar-thin">
        <div className="px-2 space-y-0.5">
          {allItems.map((item) => {
            const active = isActive(item.href);
            const Icon = item.icon;
            
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                  active
                    ? "bg-blue-50 text-blue-700"
                    : "text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                )}
              >
                <Icon className={cn("w-4 h-4 flex-shrink-0", collapsed && "mx-auto")} />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            );
          })}
        </div>
        
        {/* Bottom Actions */}
        <div className="mt-4 pt-4 border-t border-gray-200 px-2">
          <button
            onClick={handleLogout}
            className={cn(
              "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 w-full",
              collapsed && "justify-center px-2"
            )}
          >
            <KeyRound className={cn("w-4 h-4 flex-shrink-0", collapsed && "mx-auto")} />
            {!collapsed && <span>Sign out</span>}
          </button>
        </div>
      </nav>
      
      {/* Collapsed Tooltip */}
      {collapsed && (
        <div className="absolute left-16 top-2 bg-gray-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity whitespace-nowrap z-50">
          Expand
        </div>
      )}
    </aside>
  );
}
