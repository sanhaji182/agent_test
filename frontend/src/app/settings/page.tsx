"use client";

import { Section } from "@/components/ui/section";

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-[var(--text-primary)]">Settings</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">System configuration</p>
      </div>

      <div className="grid gap-4 max-w-2xl">
        <Section title="API Configuration">
          <dl className="space-y-3 text-sm">
            <SettingRow label="API URL" value={process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"} />
            <SettingRow label="LLM Provider" value="Anthropic Claude" />
            <SettingRow label="Max Fix Attempts" value="3" />
            <SettingRow label="Timeout" value="300s" />
          </dl>
        </Section>

        <Section title="Steel Browser">
          <dl className="space-y-3 text-sm">
            <SettingRow label="Steel API" value="http://steel-browser:3000" />
            <SettingRow label="Max Sessions" value="10" />
          </dl>
        </Section>

        <Section title="LangGraph Sidecar">
          <dl className="space-y-3 text-sm">
            <SettingRow label="Sidecar URL" value="http://langgraph-sidecar:8000" />
            <SettingRow label="Mode" value="Available" />
          </dl>
        </Section>
      </div>
    </div>
  );
}

function SettingRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-1">
      <dt className="text-[var(--text-secondary)]">{label}</dt>
      <dd className="font-mono text-xs text-[var(--text-primary)] bg-[var(--bg-secondary)] px-2 py-1 rounded">{value}</dd>
    </div>
  );
}
