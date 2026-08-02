"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  getLLMProfiles,
  createLLMProfile,
  updateLLMProfile,
  deleteLLMProfile,
  activateLLMProfile,
  testLLMProfile,
  getAIProviders,
  getAIModels,
  type LLMProfile,
  type AIProvider,
} from "@/lib/api";
import {
  Plus, RefreshCw, Loader2, CheckCircle2, XCircle, Trash2,
  PencilLine, Zap, Star,
} from "lucide-react";

const emptyForm = {
  name: "",
  provider: "openai-compatible",
  base_url: "",
  api_key: "",
  model: "",
  temperature: "0.2",
  max_tokens: "4096",
  is_active: false,
};

export function ProfilesManager() {
  const [profiles, setProfiles] = useState<LLMProfile[]>([]);
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [loading, setLoading] = useState(true);

  // form state
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ ...emptyForm });
  const [models, setModels] = useState<string[]>([]);
  const [fetchingModels, setFetchingModels] = useState(false);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");

  // per-row action state
  const [testingId, setTestingId] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<Record<string, { success: boolean; error?: string }>>({});
  const [actionMsg, setActionMsg] = useState("");

  const load = async () => {
    try {
      const [profs, provs] = await Promise.all([
        getLLMProfiles().catch(() => []),
        getAIProviders().catch(() => []),
      ]);
      setProfiles(profs);
      setProviders(provs);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const set = (k: keyof typeof form, v: string | boolean) =>
    setForm((prev) => ({ ...prev, [k]: v }));

  const openAdd = () => {
    setEditingId(null);
    setForm({ ...emptyForm });
    setModels([]);
    setFormError("");
    setShowForm(true);
  };

  const openEdit = (p: LLMProfile) => {
    setEditingId(p.id);
    setForm({
      name: p.name,
      provider: p.provider,
      base_url: p.base_url,
      api_key: "", // masked; kosong = tidak diubah
      model: p.model,
      temperature: p.temperature || "0.2",
      max_tokens: p.max_tokens || "4096",
      is_active: p.is_active,
    });
    setModels([]);
    setFormError("");
    setShowForm(true);
  };

  const handleFetchModels = async () => {
    setFetchingModels(true);
    setFormError("");
    try {
      const res = await getAIModels({ provider: form.provider, api_key: form.api_key, base_url: form.base_url });
      if (res.error) {
        setFormError(res.error);
        setModels([]);
      } else {
        setModels(res.models || []);
      }
    } catch (e) {
      setFormError(e instanceof Error ? e.message : "gagal fetch models");
    } finally {
      setFetchingModels(false);
    }
  };

  const handleSave = async () => {
    if (!form.name.trim()) {
      setFormError("Nama profil wajib diisi.");
      return;
    }
    setSaving(true);
    setFormError("");
    try {
      const payload: Partial<LLMProfile> = {
        name: form.name,
        provider: form.provider,
        base_url: form.base_url,
        model: form.model,
        temperature: form.temperature,
        max_tokens: form.max_tokens,
        is_active: form.is_active,
      };
      // Hanya kirim api_key jika diisi (edit: kosong = pertahankan yang lama).
      if (form.api_key) payload.api_key = form.api_key;

      if (editingId) {
        await updateLLMProfile(editingId, payload);
      } else {
        await createLLMProfile(payload);
      }
      setShowForm(false);
      await load();
      flash(editingId ? "Profil diperbarui." : "Profil dibuat.");
    } catch (e) {
      setFormError(e instanceof Error ? e.message : "gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  const handleActivate = async (id: string) => {
    await activateLLMProfile(id);
    await load();
    flash("Profil diaktifkan — akan dipakai untuk run test.");
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      const res = await testLLMProfile(id);
      setTestResults((prev) => ({ ...prev, [id]: res }));
    } catch (e) {
      setTestResults((prev) => ({ ...prev, [id]: { success: false, error: e instanceof Error ? e.message : "error" } }));
    } finally {
      setTestingId(null);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Hapus profil ini?")) return;
    await deleteLLMProfile(id);
    await load();
    flash("Profil dihapus.");
  };

  const flash = (msg: string) => {
    setActionMsg(msg);
    setTimeout(() => setActionMsg(""), 2500);
  };

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-sm text-[var(--text-muted)] py-6 justify-center">
        <Loader2 className="w-4 h-4 animate-spin" /> Loading profiles…
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">Provider Profiles</h3>
          <p className="text-xs text-[var(--text-muted)] mt-0.5">
            Simpan beberapa provider; aktifkan salah satu untuk dipakai run test.
          </p>
        </div>
        <Button size="sm" onClick={openAdd}>
          <Plus className="w-3.5 h-3.5" />
          <span className="ml-1">Add Profile</span>
        </Button>
      </div>

      {actionMsg && (
        <div className="flex items-center gap-2 text-sm text-green-700 bg-green-50 border border-green-200 rounded-lg px-3 py-2">
          <CheckCircle2 className="w-4 h-4" /> {actionMsg}
        </div>
      )}

      {/* List */}
      {profiles.length === 0 && !showForm && (
        <div className="text-sm text-[var(--text-muted)] border border-dashed border-[var(--border-default)] rounded-lg px-4 py-8 text-center">
          Belum ada profil. Klik <span className="font-medium">Add Profile</span> untuk membuat.
        </div>
      )}

      <div className="space-y-2">
        {profiles.map((p) => {
          const tr = testResults[p.id];
          return (
            <div
              key={p.id}
              className={`border rounded-lg px-4 py-3 ${
                p.is_active ? "border-blue-300 bg-blue-50/50" : "border-[var(--border-default)] bg-white"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-[var(--text-primary)] truncate">{p.name}</span>
                    {p.is_active && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold bg-blue-100 text-blue-700">
                        <Star className="w-3 h-3" /> ACTIVE
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-[var(--text-muted)] mt-0.5 truncate">
                    {p.provider} · {p.model || "(no model)"}
                    {p.base_url ? ` · ${p.base_url}` : ""}
                  </p>
                  {tr && (
                    <p className={`text-xs mt-1 flex items-center gap-1 ${tr.success ? "text-green-600" : "text-red-600"}`}>
                      {tr.success ? <CheckCircle2 className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
                      {tr.success ? "Koneksi OK" : `Gagal: ${tr.error}`}
                    </p>
                  )}
                </div>

                <div className="flex items-center gap-1 shrink-0">
                  {!p.is_active && (
                    <Button size="sm" variant="secondary" onClick={() => handleActivate(p.id)}>
                      <Zap className="w-3.5 h-3.5" />
                      <span className="ml-1">Activate</span>
                    </Button>
                  )}
                  <Button size="sm" variant="ghost" onClick={() => handleTest(p.id)} disabled={testingId === p.id}>
                    {testingId === p.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
                    <span className="ml-1">Test</span>
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => openEdit(p)}>
                    <PencilLine className="w-3.5 h-3.5" />
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => handleDelete(p.id)} className="text-red-600 hover:text-red-700 hover:bg-red-50">
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Form */}
      {showForm && (
        <div className="border border-[var(--border-default)] rounded-lg p-4 bg-white space-y-4">
          <h4 className="text-sm font-semibold text-[var(--text-primary)]">
            {editingId ? "Edit Profile" : "New Profile"}
          </h4>

          <div>
            <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              Profile Name <span className="text-red-500">*</span>
            </label>
            <Input placeholder="cth: Production OpenAI, DeepSeek Murah, Local Ollama" value={form.name} onChange={(e) => set("name", e.target.value)} />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Provider</label>
              <select
                className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                value={form.provider}
                onChange={(e) => { set("provider", e.target.value); setModels([]); }}
              >
                {providers.length === 0 && <option value={form.provider}>{form.provider}</option>}
                {providers.map((pv) => (
                  <option key={pv.id} value={pv.id}>{pv.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Base URL</label>
              <Input placeholder="https://your-proxy.example.com/v1" value={form.base_url} onChange={(e) => set("base_url", e.target.value)} />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              API Key {editingId && <span className="text-xs text-[var(--text-muted)] font-normal">(kosongkan untuk mempertahankan)</span>}
            </label>
            <Input type="password" placeholder="sk-..." value={form.api_key} onChange={(e) => set("api_key", e.target.value)} />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="block text-sm font-medium text-[var(--text-primary)]">Model</label>
              <Button variant="secondary" size="sm" onClick={handleFetchModels} disabled={fetchingModels}>
                {fetchingModels ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
                <span className="ml-1">{fetchingModels ? "Fetching…" : "Fetch Models"}</span>
              </Button>
            </div>
            <Input placeholder="cth: gpt-5.5, claude-sonnet-4.6, atau ketik manual" value={form.model} onChange={(e) => set("model", e.target.value)} list="profile-models" />
            <datalist id="profile-models">
              {(models.length > 0 ? models : providers.find((pv) => pv.id === form.provider)?.models || []).map((m) => (
                <option key={m} value={m} />
              ))}
            </datalist>
            {models.length > 0 && <p className="text-xs text-green-600 mt-1">{models.length} model tersedia.</p>}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Temperature</label>
              <Input type="number" min="0" max="1" step="0.1" value={form.temperature} onChange={(e) => set("temperature", e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Max Tokens</label>
              <Input type="number" value={form.max_tokens} onChange={(e) => set("max_tokens", e.target.value)} />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm text-[var(--text-primary)] cursor-pointer">
            <input type="checkbox" checked={form.is_active} onChange={(e) => set("is_active", e.target.checked)} className="w-4 h-4 rounded border-[var(--border-default)]" />
            Aktifkan profil ini setelah disimpan (profil lain akan dinonaktifkan)
          </label>

          {formError && (
            <div className="flex items-center gap-2 text-sm text-red-700 bg-red-50 border border-red-200 rounded-lg px-3 py-2">
              <XCircle className="w-4 h-4" /> {formError}
            </div>
          )}

          <div className="flex items-center gap-3 pt-1">
            <Button onClick={handleSave} disabled={saving}>
              {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle2 className="w-4 h-4" />}
              <span className="ml-1">{saving ? "Saving…" : "Save Profile"}</span>
            </Button>
            <Button variant="secondary" onClick={() => setShowForm(false)}>Cancel</Button>
          </div>
        </div>
      )}
    </div>
  );
}
