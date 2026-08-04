"use client";

import { usePathname } from "next/navigation";
import { Sidebar } from "@/components/sidebar";
import { useSidebar } from "@/components/layout/sidebar-context";

export function AppShell({ children }: { children: React.ReactNode }) {
  const { collapsed } = useSidebar();
  const pathname = usePathname();
  const isLogin = pathname === "/login";

  if (isLogin) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-[var(--bg-page)]">
      <Sidebar />
      <div className={`${collapsed ? "pl-16" : "pl-60"} transition-all duration-300 ease-in-out`}>
        <div className="px-6 lg:px-10 py-6 lg:py-8">
          <div className="mx-auto w-full max-w-[1440px]">{children}</div>
        </div>
      </div>
    </div>
  );
}
