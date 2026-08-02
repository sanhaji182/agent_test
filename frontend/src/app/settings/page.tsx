"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section } from "@/components/ui/section";
import { ProfilesManager } from "@/components/settings/profiles-manager";
import {
  getAIProviders,
  getAIModels,
  testAIProvider,
  getSettings,
  saveSettings,
  type AIProvider,
} from "@/lib/api";
import {
  Settings, Terminal, Globe, FileText, Save, CheckCircle2,
  AlertTriangle, RefreshCw, Loader2, XCircle,
} from "lucide-react";

// Provider yang memakai base URL custom (OpenAI-compatible endpoint).
const CUSTOM_BASE_PROVIDERS = new Set([
  "custom", "openai-compatible", "local", "ollama", "openrouter", "huggingface",
]);

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<"general" | "ai" | "integrations">("ai");

  // --- AI provider state ---
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [provider, setProvider] = useState("anthropic");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [originalApiKey, setOriginalApiKey] = useState(""); // masked value from server
  const [model, setModel] = useState("");
  const [models, setModels] = useState<string[]>([]);
  const [temperature, setTemperature] = useState("0.2");
  const [maxTokens, setMaxTokens] = useState("4096");

  // --- async state ---
  const [loadingSettings, setLoadingSettings] = useState(true);
  const [fetchingModels, setFetchingModels] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; error?: string } | null>(null);
  const [saveState, setSaveState] = useState<"idle" | "saving" | "saved" | "error">("idle");

  // Load providers + persisted settings on mount.
  useEffect(() => {
    (async () => {
      try {
        const [provs, settings] = await Promise.all([
          getAIProviders().catch(() => [] as AIProvider[]),
          getSettings().catch(() => ({} as Record<string, string>)),
        ]);
        setProviders(provs);
        if (settings.llm_provider) setProvider(settings.llm_provider);
        if (settings.llm_base_url) setBaseUrl(settings.llm_base_url);
        if (settings.llm_api_key) {
          setApiKey(settings.llm_api_key);
          setOriginalApiKey(settings.llm_api_key);
        }
        if (settings.llm_model) setModel(settings.llm_model);
        if (settings.llm_temperature) setTemperature(settings.llm_temperature);
        if (settings.llm_max_tokens) setMaxTokens(settings.llm_max_tokens);
      } finally {
        setLoadingSettings(false);
      }
    })();
  }, []);

  const isCustomBase = CUSTOM_BASE_PROVIDERS.has(provider.toLowerCase());

  const handleFetchModels = async () => {
    setFetchingModels(true);
    setTestResult(null);
    try {
      const res = await getAIModels({ provider, api_key: apiKey, base_url: baseUrl });
      if (res.error) {
        setTestResult({ success: false, error: res.error });
        setModels([]);
      } else {
        setModels(res.models || []);
        setTestResult({ success: true });
      }
    } catch (e) {
      setTestResult({ success: false, error: e instanceof Error ? e.message : "failed to fetch models" });
    } finally {
      setFetchingModels(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const res = await testAIProvider({ provider, model, api_key: apiKey, base_url: baseUrl });
      setTestResult(res);
    } catch (e) {
      setTestResult({ success: false, error: e instanceof Error ? e.message : "connection failed" });
    } finally {
      setTesting(false);
    }
  };

  const handleSave = async () => {
    setSaveState("saving");
    try {
      const payload: Record<string, string> = {
        llm_provider: provider,
        llm_model: model,
        llm_base_url: baseUrl,
        llm_temperature: temperature,
        llm_max_tokens: maxTokens,
      };
      // Only send the API key if the user actually changed it (server returns
      // a masked value; sending it back would overwrite the real key).
      if (apiKey && apiKey !== originalApiKey) {
        payload.llm_api_key = apiKey;
      }
      await saveSettings(payload);
      setOriginalApiKey(apiKey);
      setSaveState("saved");
      setTimeout(() => setSaveState("idle"), 2500);
    } catch {
      setSaveState("error");
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Settings</h1>
        <p className="text-sm text-[var(--text-muted)]">Configure your testing environment and preferences</p>
      </div>

      {/* Tabs */}
      <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
        {[
          { id: "ai", label: "AI Provider", icon: Terminal },
          { id: "general", label: "General", icon: Settings },
          { id: "integrations", label: "Integrations", icon: Globe },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as typeof activeTab)}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === tab.id
                ? "bg-blue-600 text-white shadow-sm"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
            }`}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* ── AI Provider Tab ── */}
      {activeTab === "ai" && (
        <Section title="LLM Provider Configuration">
          {loadingSettings ? (
            <div className="flex items-center gap-2 text-sm text-[var(--text-muted)] py-8 justify-center">
              <Loader2 className="w-4 h-4 animate-spin" /> Loading settings…
            </div>
          ) : (
            <div className="space-y-5">
              {/* Provider */}
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Provider</label>
                <select
                  className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                  value={provider}
                  onChange={(e) => { setProvider(e.target.value); setModels([]); setTestResult(null); }}
                >
                  {providers.length === 0 && <option value={provider}>{provider}</option>}
                  {providers.map((p) => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </select>
                <p className="text-xs text-[var(--text-muted)] mt-1">
                  Pilih <span className="font-medium">Custom</span> untuk provider OpenAI-compatible apa pun (termasuk proxy self-hosted).
                </p>
              </div>

              {/* Base URL */}
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                  Base URL {isCustomBase && <span className="text-red-500">*</span>}
                </label>
                <Input
                  placeholder={isCustomBase ? "https://your-proxy.example.com/v1" : "https://api.openai.com/v1 (default)"}
                  value={baseUrl}
                  onChange={(e) => setBaseUrl(e.target.value)}
                  helperText={isCustomBase
                    ? "Endpoint OpenAI-compatible. Harus berakhir sebelum /chat/completions (cth: .../v1)."
                    : "Kosongkan untuk memakai endpoint default provider."}
                />
              </div>

              {/* API Key */}
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                  API Key {provider !== "local" && provider !== "ollama" && <span className="text-red-500">*</span>}
                </label>
                <Input
                  type="password"
                  placeholder="sk-..."
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  helperText="Disimpan terenkripsi. Nilai yang tersimpan ditampilkan tersamar (masked)."
                />
              </div>

              {/* Model + Fetch */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-sm font-medium text-[var(--text-primary)]">Model</label>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={handleFetchModels}
                    disabled={fetchingModels || (isCustomBase && !baseUrl)}
                  >
                    {fetchingModels ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
                    <span className="ml-1">{fetchingModels ? "Fetching…" : "Fetch Models"}</span>
                  </Button>
                </div>
                <Input
                  placeholder="cth: gpt-5.5, claude-sonnet-4.6, atau ketik manual"
                  value={model}
                  onChange={(e) => setModel(e.target.value)}
                  list="llm-models"
                />
                <datalist id="llm-models">
                  {(models.length > 0
                    ? models
                    : providers.find((p) => p.id === provider)?.models || []
                  ).map((m) => (
                    <option key={m} value={m} />
                  ))}
                </datalist>
                {models.length > 0 && (
                  <p className="text-xs text-green-600 mt-1">{models.length} model tersedia dari provider.</p>
                )}
              </div>

              {/* Advanced: temperature + max tokens */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Temperature (0–1)</label>
                  <Input type="number" min="0" max="1" step="0.1" value={temperature} onChange={(e) => setTemperature(e.target.value)} />
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Max Tokens</label>
                  <Input type="number" value={maxTokens} onChange={(e) => setMaxTokens(e.target.value)} />
                </div>
              </div>

              {/* Test result */}
              {testResult && (
                <div className={`flex items-start gap-2 p-3 rounded-lg border text-sm ${
                  testResult.success
                    ? "bg-green-50 border-green-200 text-green-800"
                    : "bg-red-50 border-red-200 text-red-800"
                }`}>
                  {testResult.success ? <CheckCircle2 className="w-4 h-4 shrink-0 mt-0.5" /> : <XCircle className="w-4 h-4 shrink-0 mt-0.5" />}
                  <span>{testResult.success ? "Koneksi berhasil." : `Gagal: ${testResult.error}`}</span>
                </div>
              )}

              {/* Actions */}
              <div className="flex flex-wrap items-center gap-3 pt-2">
                <Button variant="secondary" onClick={handleTest} disabled={testing || !model}>
                  {testing ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
                  <span className="ml-1">{testing ? "Testing…" : "Test Connection"}</span>
                </Button>
                <Button onClick={handleSave} disabled={saveState === "saving"}>
                  {saveState === "saving" ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                  <span className="ml-1">{saveState === "saving" ? "Saving…" : "Save Configuration"}</span>
                </Button>
                {saveState === "saved" && (
                  <span className="flex items-center gap-1 text-sm text-green-600">
                    <CheckCircle2 className="w-4 h-4" /> Tersimpan
                  </span>
                )}
                {saveState === "error" && (
                  <span className="flex items-center gap-1 text-sm text-red-600">
                    <AlertTriangle className="w-4 h-4" /> Gagal menyimpan
                  </span>
                )}
              </div>
            </div>
          )}
        </Section>
      )}

      {/* ── Provider Profiles (multi-provider) ── */}
      {activeTab === "ai" && (
        <Section title="Provider Profiles">
          <ProfilesManager />
        </Section>
      )}

      {/* ── General Tab ── */}
      {activeTab === "general" && (
        <Section title="General Settings">
          <div className="rounded-lg bg-blue-50 border border-blue-200 p-4">
            <p className="text-sm text-blue-800">
              Konfigurasi umum (nama aplikasi, timezone, bahasa) tersedia pada versi mendatang.
              Fokus saat ini adalah konfigurasi AI provider.
            </p>
          </div>
        </Section>
      )}

      {/* ── Integrations Tab ── */}
      {activeTab === "integrations" && (
        <Section title="External Integrations">
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-white border border-[var(--border-default)] rounded-lg">
              <div className="flex items-center gap-3">
                <FileText className="w-5 h-5 text-[var(--text-muted)]" />
                <div>
                  <p className="text-sm font-medium text-[var(--text-primary)]">GitHub Webhooks</p>
                  <p className="text-xs text-[var(--text-muted)] mt-0.5">Auto-detect test drift on code changes</p>
                </div>
              </div>
            </div>
            <p className="text-xs text-[var(--text-muted)]">
              Integrasi Slack/GitHub dikonfigurasi melalui environment variables dan schedule webhook.
            </p>
          </div>
        </Section>
      )}
    </div>
  );
}
