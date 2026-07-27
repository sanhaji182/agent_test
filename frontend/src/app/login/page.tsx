"use client";

import { useState } from "react";
import { login } from "@/lib/api";
import { KeyRound, Loader2, AlertTriangle } from "lucide-react";

export default function LoginPage() {
  const [apiKey, setApiKey] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!apiKey.trim()) return;
    setLoading(true);
    setError(null);
    try {
      const result = await login(apiKey.trim());
      if (result.status === "ok") {
        window.location.href = result.redirect || "/";
      } else {
        setError("Invalid API key. Please try again.");
      }
    } catch {
      setError("Connection failed. Check that the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg-page)]">
      <div className="w-full max-w-[380px] mx-4">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
            GoTest Agent
          </h1>
          <p className="mt-2 text-sm text-[var(--text-muted)]">
            Enter your API key to access the dashboard
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="bg-[var(--bg-card)] rounded-xl border border-[var(--border)] p-6 space-y-4 shadow-[var(--shadow-sm)]"
        >
          {error && (
            <div className="flex items-start gap-2 p-3 rounded-lg bg-[var(--danger-bg)] text-[var(--danger)] text-sm border border-[var(--danger)]/20">
              <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label
              htmlFor="api-key"
              className="block text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wider mb-1.5"
            >
              API Key
            </label>
            <div className="relative">
              <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
              <input
                id="api-key"
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter your API key..."
                autoFocus
                disabled={loading}
                className="w-full pl-9 pr-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg-input)] text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:border-[var(--accent)] disabled:opacity-50 transition-colors"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading || !apiKey.trim()}
            className="w-full flex items-center justify-center gap-2 py-2.5 rounded-lg bg-[var(--accent)] text-white text-sm font-semibold hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            {loading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Signing in...
              </>
            ) : (
              "Sign in"
            )}
          </button>
        </form>

        <p className="mt-4 text-center text-xs text-[var(--text-muted)]">
          Your API key is sent to the backend and exchanged for a secure session
          cookie. It is never stored in the browser.
        </p>
      </div>
    </div>
  );
}
