"use client";

export default function SettingsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Settings</h1>
      <div className="space-y-6 max-w-lg">
        <section className="border rounded-lg p-4">
          <h2 className="font-semibold mb-3">API Configuration</h2>
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-zinc-500">API URL</dt>
              <dd className="font-mono">
                {process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}
              </dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-zinc-500">LLM Provider</dt>
              <dd>Anthropic (Claude)</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-zinc-500">Max Fix Attempts</dt>
              <dd>3</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-zinc-500">Timeout</dt>
              <dd>300s</dd>
            </div>
          </dl>
        </section>

        <section className="border rounded-lg p-4">
          <h2 className="font-semibold mb-3">Steel Browser</h2>
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-zinc-500">Steel API</dt>
              <dd className="font-mono">http://localhost:3000</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-zinc-500">Max Sessions</dt>
              <dd>10</dd>
            </div>
          </dl>
        </section>

        <section className="border rounded-lg p-4">
          <h2 className="font-semibold mb-3">Advanced Agent</h2>
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-zinc-500">LangGraph Sidecar</dt>
              <dd className="font-mono">http://localhost:8000</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-zinc-500">Status</dt>
              <dd className="text-zinc-400">Not connected</dd>
            </div>
          </dl>
        </section>
      </div>
    </div>
  );
}
