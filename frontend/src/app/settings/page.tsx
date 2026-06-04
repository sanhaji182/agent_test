"use client";

import { useState, useEffect, useCallback } from "react";
import { Section } from "@/components/ui/section";

import { getAIProviders, testAIProvider, type AIProvider } from "@/lib/api";

type Settings = Record<string, string>;

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings>({});
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);

  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; error?: string } | null>(null);

  const fetchSettings = useCallback(async () => {
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/v1/settings`);
      if (res.ok) {
        const data = await res.json();
        setSettings(data);
        setApiKeyInput(data.llm_api_key || "");
      }
      const aiProviders = await getAIProviders();
      setProviders(aiProviders);
    } catch {
      // Fallback if backend unreachable
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { 
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchSettings(); 
  }, [fetchSettings]);

  const updateSettings = async (updates: Partial<Settings>) => {
    setSaving(true);
    setSaved(false);
    const newSettings = { ...settings, ...updates } as Settings;
    setSettings(newSettings);
    try {
      await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/v1/settings`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch {
      // Handle error silently
    } finally {
      setSaving(false);
    }
  };

  const selectedProvider = providers.find((p) => p.id === (settings.llm_provider || "anthropic")) || providers[0] || { id: "anthropic", name: "Anthropic", models: [] };

  const handleTestConnection = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const result = await testAIProvider({
        provider: selectedProvider.id,
        model: settings.llm_model || selectedProvider.models[0] || "",
        api_key: apiKeyInput,
        base_url: settings.llm_base_url || "",
      });
      setTestResult(result);
    } catch (err: unknown) {
      setTestResult({ success: false, error: err instanceof Error ? err.message : String(err) });
    } finally {
      setTesting(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Settings</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-0.5">Loading configuration...</p>
        </div>
        <div className="space-y-4 max-w-3xl">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-32 rounded-[var(--radius)] bg-[var(--bg-subtle)] animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Settings</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-0.5">Configure your AI testing engine</p>
        </div>
        {saved && (
          <div className="flex items-center gap-1.5 text-xs font-medium text-[var(--success)] bg-[var(--success-bg)] px-3 py-1.5 rounded-full animate-in fade-in duration-300">
            <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>
            Settings saved
          </div>
        )}
      </div>

      <div className="grid gap-5 max-w-3xl">
        {/* ─── AI Provider Selection ─── */}
        <Section title="🤖 AI Provider">
          <div className="grid grid-cols-3 gap-3">
            {providers.map((provider) => {
              const isSelected = provider.id === (settings.llm_provider || "anthropic");
              return (
                <button
                  key={provider.id}
                  onClick={() => {
                    const defaultModel = provider.models[0];
                    updateSettings({ llm_provider: provider.id, llm_model: defaultModel });
                  }}
                  className={`relative flex flex-col items-center gap-2 p-4 rounded-[var(--radius)] border-2 transition-all duration-200 cursor-pointer group
                    ${isSelected
                      ? "border-[var(--accent)] bg-[var(--accent-bg)] shadow-[0_0_0_1px_var(--accent)]"
                      : "border-[var(--border)] bg-[var(--bg-card)] hover:border-[var(--border-strong)] hover:bg-[var(--bg-hover)]"
                    }`}
                >
                  <span className="text-2xl">{provider.id === 'anthropic' ? '🧠' : provider.id === 'openai' ? '✨' : '💎'}</span>
                  <span className={`text-sm font-semibold ${isSelected ? "text-[var(--accent)]" : "text-[var(--text-primary)]"}`}>
                    {provider.name}
                  </span>
                  {isSelected && (
                    <div className="absolute top-2 right-2 w-5 h-5 rounded-full bg-[var(--accent)] flex items-center justify-center">
                      <svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="white" strokeWidth={3}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        </Section>

        {/* ─── Model Selection ─── */}
        <Section title="🎯 Model Selection">
          <div className="mb-5 p-4 rounded-[var(--radius)] bg-[var(--accent-bg)] border border-[var(--accent)]/30 text-sm">
            <h4 className="font-semibold text-[var(--accent)] mb-1 flex items-center gap-2">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.9 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>
              AI Brain Recommendations
            </h4>
            <p className="text-[11px] text-[var(--text-muted)] mb-4">Updated June 2026 — Based on latest frontier model releases</p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-[13px]">
              
              {/* Powerful */}
              <div className="space-y-2 p-3.5 rounded-[var(--radius-sm)] bg-[var(--bg-card)] border border-purple-500/30 relative overflow-hidden">
                <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-purple-500 to-pink-500" />
                <div className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                  <span>🔥</span> Powerful <span className="text-[11px] font-normal text-[var(--text-muted)]">(Max Capability)</span>
                </div>
                <div className="space-y-1 text-[var(--text-secondary)] leading-relaxed">
                  <div className="flex items-center gap-1.5"><span className="text-[10px] px-1 py-0.5 rounded bg-purple-500/10 text-purple-400 font-medium shrink-0">BEST</span> <strong>Claude Opus 4.8</strong></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>GPT-5.5-Pro</strong> <span className="text-[11px] text-[var(--text-muted)]">— OpenAI</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>Gemini 3.5 Pro</strong> <span className="text-[11px] text-[var(--text-muted)]">— Google</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>Qwen 3.7-Max</strong> <span className="text-[11px] text-[var(--text-muted)]">— Alibaba</span></div>
                </div>
                <p className="text-[11px] text-[var(--text-muted)] pt-1 border-t border-[var(--border)]">Agentic planning, autonomous coding, complex multi-step reasoning</p>
              </div>

              {/* Sweetspot */}
              <div className="space-y-2 p-3.5 rounded-[var(--radius-sm)] bg-[var(--bg-card)] border border-[var(--accent)]/30 relative overflow-hidden">
                <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-[var(--accent)] to-cyan-400" />
                <div className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                  <span>⭐</span> Sweetspot <span className="text-[11px] font-normal text-[var(--text-muted)]">(Best Balance)</span>
                </div>
                <div className="space-y-1 text-[var(--text-secondary)] leading-relaxed">
                  <div className="flex items-center gap-1.5"><span className="text-[10px] px-1 py-0.5 rounded bg-[var(--accent)]/10 text-[var(--accent)] font-medium shrink-0">REC</span> <strong>Claude Sonnet 4.6</strong></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>GPT-5.4</strong> <span className="text-[11px] text-[var(--text-muted)]">— OpenAI</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>Gemini 3.1 Pro</strong> <span className="text-[11px] text-[var(--text-muted)]">— Google (1M ctx)</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>DeepSeek V4 Pro</strong> <span className="text-[11px] text-[var(--text-muted)]">— DeepSeek</span></div>
                </div>
                <p className="text-[11px] text-[var(--text-muted)] pt-1 border-t border-[var(--border)]">Best price-to-performance ratio for daily testing and code generation</p>
              </div>

              {/* Optimal */}
              <div className="space-y-2 p-3.5 rounded-[var(--radius-sm)] bg-[var(--bg-card)] border border-blue-500/30 relative overflow-hidden">
                <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-blue-500 to-sky-400" />
                <div className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                  <span>⚡</span> Optimal <span className="text-[11px] font-normal text-[var(--text-muted)]">(High Speed)</span>
                </div>
                <div className="space-y-1 text-[var(--text-secondary)] leading-relaxed">
                  <div className="flex items-center gap-1.5"><span className="text-[10px] px-1 py-0.5 rounded bg-blue-500/10 text-blue-400 font-medium shrink-0">FAST</span> <strong>Gemini 3.5 Flash</strong></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>Claude Haiku 4.5</strong> <span className="text-[11px] text-[var(--text-muted)]">— Anthropic</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>GPT-5.4-Mini</strong> <span className="text-[11px] text-[var(--text-muted)]">— OpenAI</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>DeepSeek V4 Flash</strong> <span className="text-[11px] text-[var(--text-muted)]">— DeepSeek</span></div>
                </div>
                <p className="text-[11px] text-[var(--text-muted)] pt-1 border-t border-[var(--border)]">Ultra-fast inference for routine test runs and quick evaluations</p>
              </div>

              {/* Cheapest */}
              <div className="space-y-2 p-3.5 rounded-[var(--radius-sm)] bg-[var(--bg-card)] border border-emerald-500/30 relative overflow-hidden">
                <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-emerald-500 to-green-400" />
                <div className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                  <span>💰</span> Cheapest <span className="text-[11px] font-normal text-[var(--text-muted)]">(Budget / Local)</span>
                </div>
                <div className="space-y-1 text-[var(--text-secondary)] leading-relaxed">
                  <div className="flex items-center gap-1.5"><span className="text-[10px] px-1 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-medium shrink-0">FREE</span> <strong>Llama 4 Maverick</strong> <span className="text-[11px] text-[var(--text-muted)]">— Meta (MoE)</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>GPT-5.4-Nano</strong> <span className="text-[11px] text-[var(--text-muted)]">— $0.20/1M in</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>Gemini 3.1 Flash-Lite</strong> <span className="text-[11px] text-[var(--text-muted)]">— Google</span></div>
                  <div className="flex items-center gap-1.5"><span className="w-1 h-1 rounded-full bg-[var(--text-muted)] shrink-0" /> <strong>DeepSeek R1</strong> <span className="text-[11px] text-[var(--text-muted)]">— Open-weights / Ollama</span></div>
                </div>
                <p className="text-[11px] text-[var(--text-muted)] pt-1 border-t border-[var(--border)]">Lowest cost APIs or self-hosted via Ollama / vLLM for full privacy</p>
              </div>

            </div>
          </div>

          <div className="space-y-2">
            {selectedProvider.models.length === 0 ? (
              <div className="space-y-3">
                <p className="text-xs text-[var(--text-muted)]">
                  Enter the exact model ID for your custom provider (e.g. <code>deepseek-chat</code>, <code>llama-3-8b</code>).
                </p>
                <input
                  type="text"
                  value={settings.llm_model || ""}
                  onChange={(e) => updateSettings({ llm_model: e.target.value })}
                  placeholder="e.g. llama3"
                  className="w-full px-3 py-2 text-sm rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
                />
              </div>
            ) : (
              selectedProvider.models.map((model) => {
                const isSelected = model === (settings.llm_model || selectedProvider.models[0]);
                return (
                  <button
                    key={model}
                    onClick={() => updateSettings({ llm_model: model })}
                    className={`w-full flex items-center justify-between px-4 py-3 rounded-[var(--radius-sm)] border transition-all duration-150 cursor-pointer
                      ${isSelected
                        ? "border-[var(--accent)] bg-[var(--accent-bg)]"
                        : "border-[var(--border)] bg-[var(--bg-card)] hover:bg-[var(--bg-hover)]"
                      }`}
                  >
                    <div className="flex items-center gap-3">
                      <div className={`w-3 h-3 rounded-full border-2 transition-all ${isSelected ? "border-[var(--accent)] bg-[var(--accent)]" : "border-[var(--border-strong)]"}`} />
                      <span className={`text-sm font-medium ${isSelected ? "text-[var(--accent)]" : "text-[var(--text-primary)]"}`}>{model}</span>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </Section>

        {/* ─── API Key ─── */}
        <Section title="🔑 API Key">
          <div className="space-y-3">
            <p className="text-xs text-[var(--text-muted)]">
              Enter your API key for <strong className="text-[var(--text-secondary)]">{selectedProvider.name}</strong>. Keys are stored securely in the database.
            </p>
            <div className="flex gap-2">
              <input
                type="password"
                value={apiKeyInput}
                onChange={(e) => setApiKeyInput(e.target.value)}
                placeholder={`sk-...`}
                className="flex-1 px-3 py-2 text-sm rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
              />
              <button
                onClick={handleTestConnection}
                disabled={testing || !apiKeyInput}
                className="px-4 py-2 text-sm font-medium rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] hover:bg-[var(--bg-hover)] transition-colors disabled:opacity-50 cursor-pointer whitespace-nowrap"
              >
                {testing ? "Testing..." : "Test Connection"}
              </button>
              <button
                onClick={() => updateSettings({ llm_api_key: apiKeyInput })}
                disabled={saving}
                className="px-4 py-2 text-sm font-medium rounded-[var(--radius-sm)] bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)] transition-colors disabled:opacity-50 cursor-pointer whitespace-nowrap"
              >
                {saving ? "Saving..." : "Save Key"}
              </button>
            </div>
            {testResult && (
              <div className={`text-sm p-3 rounded-[var(--radius-sm)] ${testResult.success ? 'bg-[var(--success-bg)] text-[var(--success)]' : 'bg-[var(--danger-bg)] text-[var(--danger)]'}`}>
                {testResult.success ? "Connection successful! AI Provider is ready." : `Connection failed: ${testResult.error}`}
              </div>
            )}
          </div>
        </Section>

        {/* ─── Provider Endpoint ─── */}
        <Section title="🌐 Provider Endpoint">
          <div className="space-y-3">
            <p className="text-xs text-[var(--text-muted)]">
              Optional for OpenAI-compatible providers and local runtimes such as Ollama/vLLM. Leave empty for the provider default.
            </p>
            <input
              type="text"
              value={settings.llm_base_url || ""}
              onChange={(e) => updateSettings({ llm_base_url: e.target.value })}
              placeholder="http://localhost:11434/v1"
              className="w-full px-3 py-2 text-sm rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
            />
          </div>
        </Section>

        {/* ─── Integrations (CI/CD) ─── */}
        <Section title="🔗 Integrations (CI/CD)">
          <div className="space-y-3">
            <p className="text-xs text-[var(--text-muted)]">
              Automatically trigger test suites when developers open a Pull Request. Add this Webhook URL to your GitHub repository settings.
            </p>
            <div className="p-3 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] font-mono text-sm flex items-center justify-between">
              <span className="text-[var(--text-primary)] select-all truncate mr-2">
                {typeof window !== 'undefined' ? `${window.location.protocol}//${window.location.host}` : 'http://localhost:8080'}/api/v1/webhooks/github
              </span>
              <button 
                onClick={(e) => {
                  navigator.clipboard.writeText(`${window.location.protocol}//${window.location.host}/api/v1/webhooks/github`);
                  const el = e.currentTarget;
                  el.innerText = 'Copied!';
                  setTimeout(() => el.innerText = 'Copy', 2000);
                }}
                className="text-xs bg-[var(--bg-card)] px-2 py-1 rounded border border-[var(--border)] hover:bg-[var(--bg-hover)] cursor-pointer"
              >
                Copy
              </button>
            </div>
          </div>
        </Section>

        {/* ─── Advanced Settings ─── */}
        <Section title="⚙️ Advanced">
          <div className="grid grid-cols-2 gap-4">
            <SettingField
              label="Temperature"
              description="Controls randomness (0 = deterministic, 1 = creative)"
              value={settings.llm_temperature || "0.2"}
              onChange={(v) => updateSettings({ llm_temperature: v })}
              type="number"
              min="0" max="1" step="0.1"
            />
            <SettingField
              label="Max Tokens"
              description="Maximum tokens for AI response"
              value={settings.llm_max_tokens || "4096"}
              onChange={(v) => updateSettings({ llm_max_tokens: v })}
              type="number"
              min="256" max="16384" step="256"
            />
            <SettingField
              label="Browser Timeout (s)"
              description="Max seconds for test execution"
              value={settings.browser_timeout || "300"}
              onChange={(v) => updateSettings({ browser_timeout: v })}
              type="number"
              min="30" max="1800" step="30"
            />
            <SettingField
              label="Max Fix Attempts"
              description="How many times AI retries failed tests"
              value={settings.max_fix_attempts || "3"}
              onChange={(v) => updateSettings({ max_fix_attempts: v })}
              type="number"
              min="0" max="10" step="1"
            />
          </div>
          <div className="mt-4 flex items-center justify-between px-1">
            <div>
              <span className="text-sm font-medium text-[var(--text-primary)]">Headless Browser</span>
              <p className="text-xs text-[var(--text-muted)]">Run browser without visible window</p>
            </div>
            <button
              onClick={() => updateSettings({ browser_headless: settings.browser_headless === "true" ? "false" : "true" })}
              className={`relative w-11 h-6 rounded-full transition-colors duration-200 cursor-pointer ${settings.browser_headless === "true" ? "bg-[var(--accent)]" : "bg-[var(--border-strong)]"}`}
            >
              <span className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-sm transition-transform duration-200 ${settings.browser_headless === "true" ? "translate-x-5" : ""}`} />
            </button>
          </div>
        </Section>

        {/* ─── Status ─── */}
        <Section title="📊 Current Configuration">
          <dl className="space-y-2 text-sm">
            <StatusRow label="Provider" value={selectedProvider.name} icon={selectedProvider.id === 'anthropic' ? '🧠' : selectedProvider.id === 'openai' ? '✨' : selectedProvider.id === 'google' ? '🌐' : selectedProvider.id === 'deepseek' ? '🐋' : selectedProvider.id === 'local' ? '🖥️' : '💎'} />
            <StatusRow label="Model" value={settings.llm_model || selectedProvider.models[0] || ""} />
            <StatusRow label="Base URL" value={settings.llm_base_url || "Provider default"} />
            <StatusRow label="API Key" value={settings.llm_api_key && settings.llm_api_key.length > 4 ? "Configured ✓" : "Not Set ✗"} status={settings.llm_api_key && settings.llm_api_key.length > 4 ? "ok" : "error"} />
            <StatusRow label="Temperature" value={settings.llm_temperature || "0.2"} />
            <StatusRow label="Max Tokens" value={settings.llm_max_tokens || "4096"} />
          </dl>
        </Section>
      </div>
    </div>
  );
}

function SettingField({
  label, description, value, onChange, type = "text", min, max, step,
}: {
  label: string; description: string; value: string;
  onChange: (v: string) => void; type?: string;
  min?: string; max?: string; step?: string;
}) {
  const [localValue, setLocalValue] = useState(value);
  useEffect(() => { 
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLocalValue(value); 
  }, [value]);

  return (
    <div className="space-y-1.5">
      <label className="text-sm font-medium text-[var(--text-primary)]">{label}</label>
      <p className="text-[11px] text-[var(--text-muted)] leading-tight">{description}</p>
      <input
        type={type}
        value={localValue}
        min={min} max={max} step={step}
        onChange={(e) => setLocalValue(e.target.value)}
        onBlur={() => onChange(localValue)}
        className="w-full px-3 py-2 text-sm rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
      />
    </div>
  );
}

function StatusRow({ label, value, icon, status }: { label: string; value: string; icon?: string; status?: "ok" | "error" }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-[var(--border)] last:border-0">
      <dt className="text-[var(--text-secondary)] flex items-center gap-1.5">
        {icon && <span>{icon}</span>}
        {label}
      </dt>
      <dd className={`font-mono text-xs px-2.5 py-1 rounded-full ${
        status === "ok" ? "bg-[var(--success-bg)] text-[var(--success)]" :
        status === "error" ? "bg-[var(--danger-bg)] text-[var(--danger)]" :
        "bg-[var(--bg-subtle)] text-[var(--text-primary)]"
      }`}>
        {value}
      </dd>
    </div>
  );
}
