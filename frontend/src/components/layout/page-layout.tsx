"use client";

import { Sidebar } from "@/components/sidebar";
import { useState } from "react";

interface PageLayoutProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
  action?: React.ReactNode;
}

export function PageLayout({ children, title, description, action }: PageLayoutProps) {
  return (
    <div className="ml-60 min-h-screen bg-[var(--bg-page)]">
      {/* Header */}
      {(title || action) && (
        <div className="sticky top-0 z-30 bg-[var(--bg-page)]/80 backdrop-blur-md border-b border-[var(--border-default)]">
          <div className="max-w-7xl mx-auto px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                {title && (
                  <h1 className="text-xl font-semibold text-[var(--text-primary)]">{title}</h1>
                )}
                {description && (
                  <p className="text-sm text-[var(--text-muted)] mt-1">{description}</p>
                )}
              </div>
              {action && <div>{action}</div>}
            </div>
          </div>
        </div>
      )}
      
      {/* Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {children}
      </main>
    </div>
  );
}
