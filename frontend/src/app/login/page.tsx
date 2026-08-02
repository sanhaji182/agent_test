"use client";

import { useState } from "react";
import { login } from "@/lib/api";
import { KeyRound, AlertTriangle, Zap, Mail, Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function LoginPage() {
  const [mode, setMode] = useState<"password" | "apikey">("password");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const result =
        mode === "password"
          ? await login({ email: email.trim(), password })
          : await login({ apiKey: apiKey.trim() });
      if (result.status === "ok") {
        // Kembali ke halaman yang dituju (dibawa middleware via query param).
        const params = new URLSearchParams(window.location.search);
        const target = params.get("redirect") || result.redirect || "/";
        window.location.href = target;
      } else {
        setError(mode === "password" ? "Email atau password salah." : "API key tidak valid.");
      }
    } catch {
      setError("Koneksi gagal. Pastikan backend sedang berjalan.");
    } finally {
      setLoading(false);
    }
  };

  const canSubmit = mode === "password" ? email.trim() && password : apiKey.trim();

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
          <p className="text-[var(--text-muted)]">Masuk ke workspace kamu</p>
        </div>

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 text-red-700 text-sm border border-red-200">
              <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {mode === "password" ? (
            <>
              <Input
                id="email"
                type="email"
                label="Email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@gotest.local"
                autoFocus
                disabled={loading}
                leftIcon={<Mail className="w-4 h-4" />}
              />
              <Input
                id="password"
                type="password"
                label="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                disabled={loading}
                leftIcon={<Lock className="w-4 h-4" />}
              />
            </>
          ) : (
            <Input
              id="api-key"
              type="password"
              label="API Key"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="gta_..."
              autoFocus
              disabled={loading}
              leftIcon={<KeyRound className="w-4 h-4" />}
            />
          )}

          <Button type="submit" disabled={loading || !canSubmit} isLoading={loading} className="w-full">
            {loading ? "Memproses..." : "Masuk"}
          </Button>
        </form>

        {/* Switch mode */}
        <button
          type="button"
          onClick={() => {
            setMode(mode === "password" ? "apikey" : "password");
            setError(null);
          }}
          className="mt-4 w-full text-center text-xs text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
        >
          {mode === "password"
            ? "Login dengan API key (untuk CI / program)"
            : "Login dengan email + password"}
        </button>

        <p className="mt-6 text-xs text-[var(--text-muted)] text-center leading-relaxed">
          {mode === "password"
            ? "Default admin: admin@gotest.local (password = nilai API_KEY di .env saat first-run)."
            : "API key dipakai untuk akses programmatic / CI pipeline."}
        </p>
      </div>
    </div>
  );
}
