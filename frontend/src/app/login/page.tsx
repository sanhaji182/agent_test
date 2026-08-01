"use client";

import { useState } from "react";
import { login } from "@/lib/api";
import { KeyRound, Loader2, AlertTriangle, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

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
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg-page)] px-4">
      <div className="w-full max-w-[380px]">
        {/* Logo and Title */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-2 mb-4">
            <div className="w-10 h-10 rounded-lg bg-blue-600 flex items-center justify-center">
              <Zap className="w-5 h-5 text-white" />
            </div>
            <h1 className="text-2xl font-semibold text-[var(--text-primary)]">GoTest Agent</h1>
          </div>
          <p className="text-[var(--text-muted)]">Sign in to your workspace</p>
        </div>

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 text-red-700 text-sm border border-red-200">
              <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <Input
            id="api-key"
            type="password"
            label="Access Token"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="Enter your access token"
            autoFocus
            disabled={loading}
            leftIcon={<KeyRound className="w-4 h-4" />}
          />

          <Button
            type="submit"
            disabled={loading || !apiKey.trim()}
            isLoading={loading}
            className="w-full"
          >
            {loading ? "Signing in..." : "Sign in"}
          </Button>
        </form>

        <p className="mt-6 text-xs text-[var(--text-muted)] text-center leading-relaxed">
          Your access token is sent to the backend and exchanged for a secure session cookie.
          It is never stored in the browser.
        </p>
      </div>
    </div>
  );
}
