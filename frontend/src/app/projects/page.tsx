"use client";

import { EmptyState } from "@/components/ui/section";
import { FolderOpen } from "lucide-react";

export default function ProjectsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-[var(--text-primary)]">Projects</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage your registered projects</p>
      </div>
      <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
        <EmptyState
          icon={<FolderOpen className="w-6 h-6" />}
          title="No projects registered"
          description="Projects are auto-detected when you run tests via MCP or API."
        />
      </div>
    </div>
  );
}
