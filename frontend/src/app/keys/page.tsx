"use client";

import { useEffect, useState } from "react";
import {
  createAPIKey,
  deleteAPIKey,
  listAPIKeys,
  revokeAPIKey,
  listAuditLog,
  isAdmin,
  type APIKeyEntry,
  type AuditEntry,
} from "@/lib/api";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { KeyRound, Plus, Trash2, Clock } from "lucide-react";
import { cn } from "@/lib/utils";

export default function KeysPage() {
  const [keys, setKeys] = useState<APIKeyEntry[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<"keys" | "audit">("keys");

  const [showCreate, setShowCreate] = useState(false);
  const [newLabel, setNewLabel] = useState("");
  const [newRole, setNewRole] = useState("reviewer");
  const [creating, setCreating] = useState(false);
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  const refresh = async () => {
    try {
      const [keyData, auditData] = await Promise.all([
        listAPIKeys().catch(() => []),
        listAuditLog().catch(() => []),
      ]);
      setKeys(keyData);
      setAudit(auditData);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  const handleCreate = async () => {
    if (!newLabel.trim()) return;
    setCreating(true);
    try {
      const entry = await createAPIKey(newLabel.trim(), newRole);
      setCreatedKey(entry.key || null);
      setNewLabel("");
      setKeys((prev) => [entry, ...prev]);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const handleRevoke = async (id: string, active: boolean) => {
    try {
      await revokeAPIKey(id, active);
      setKeys((prev) => prev.map((k) => (k.id === id ? { ...k, active } : k)));
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this API key permanently?")) return;
    try {
      await deleteAPIKey(id);
      setKeys((prev) => prev.filter((k) => k.id !== id));
    } catch (e) {
      setError((e as Error).message);
    }
  };

  if (loading) return <LoadingSkeleton rows={8} />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">API Keys & Audit</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">
          Manage API keys with role-based access and view the audit trail.
        </p>
      </div>

      {error && (
        <div className="rounded-lg border border-[var(--danger)]/20 bg-[var(--danger-bg)] p-3 text-xs text-[var(--danger)]">
          {error}
          <button onClick={() => setError(null)} className="ml-3 underline">Dismiss</button>
        </div>
      )}

      <div className="flex items-center gap-1 border-b border-[var(--border)]">
        <button
          onClick={() => setTab("keys")}
          className={cn(
            "px-4 py-2.5 text-[13px] font-medium border-b-2 -mb-px transition-colors",
            tab === "keys" ? "border-[var(--accent)] text-[var(--accent)]" : "border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          )}
        >
          <KeyRound className="w-3.5 h-3.5 inline mr-1.5" />
          Keys ({keys.length})
        </button>
        <button
          onClick={() => setTab("audit")}
          className={cn(
            "px-4 py-2.5 text-[13px] font-medium border-b-2 -mb-px transition-colors",
            tab === "audit" ? "border-[var(--accent)] text-[var(--accent)]" : "border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          )}
        >
          <Clock className="w-3.5 h-3.5 inline mr-1.5" />
          Audit Log ({audit.length})
        </button>
      </div>

      {tab === "keys" && (
        <div className="space-y-4">
          {isAdmin() && (
          <>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-sm font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
          >
            <Plus className="w-4 h-4" /> Create Key
          </button>

          {showCreate && isAdmin() && (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] p-5 space-y-4">
              <h2 className="text-sm font-bold">Create New API Key</h2>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <label className="block text-[11px] font-semibold text-[var(--text-secondary)] mb-1.5 uppercase">Label</label>
                  <input value={newLabel} onChange={(e) => setNewLabel(e.target.value)} placeholder="e.g. Tim QA" className="input" />
                </div>
                <div>
                  <label className="block text-[11px] font-semibold text-[var(--text-secondary)] mb-1.5 uppercase">Role</label>
                  <select value={newRole} onChange={(e) => setNewRole(e.target.value)} className="input">
                    <option value="admin">Admin</option>
                    <option value="reviewer">Reviewer</option>
                    <option value="viewer">Viewer</option>
                    <option value="api_client">API Client</option>
                  </select>
                </div>
                <div className="flex items-end">
                  <button onClick={handleCreate} disabled={creating || !newLabel.trim()} className="w-full px-4 py-2 rounded-lg bg-[var(--success)] text-white text-sm font-semibold hover:bg-[#047857] disabled:opacity-50 transition-colors">
                    {creating ? "Creating..." : "Create"}
                  </button>
                </div>
              </div>
              {createdKey && (
                <div className="rounded-lg border border-[var(--success)]/30 bg-[var(--success-bg)] p-4">
                  <p className="text-[11px] font-semibold text-[var(--success)] uppercase mb-2">Key created — copy it now</p>
                  <pre className="text-[13px] font-mono text-[var(--text-primary)] bg-[var(--bg-subtle)] p-3 rounded cursor-pointer select-all" onClick={() => navigator.clipboard.writeText(createdKey)}>
                    {createdKey}
                  </pre>
                </div>
              )}
            </div>
          )}
          </>
        )}

          {keys.length === 0 ? (
            <EmptyState icon={<KeyRound className="w-6 h-6" />} title="No API keys" description="Create your first API key." />
          ) : (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] overflow-hidden">
              <div className="grid grid-cols-[1fr_100px_80px_80px_80px] gap-3 px-4 py-2.5 border-b border-[var(--border)] bg-[var(--bg-subtle)] text-[11px] font-semibold text-[var(--text-muted)] uppercase tracking-wider">
                <span>Label</span><span>Role</span><span>Status</span><span>Created</span><span />
              </div>
              <div className="divide-y divide-[var(--border)]">
                {keys.map((entry) => (
                  <div key={entry.id} className="grid grid-cols-[1fr_100px_80px_80px_80px] gap-3 px-4 py-3 items-center hover:bg-[var(--bg-hover)] transition-colors">
                    <div className="min-w-0">
                      <p className="text-sm font-medium truncate">{entry.label}</p>
                      <p className="text-[11px] text-[var(--text-muted)] font-mono truncate">{entry.id.slice(0, 12)}</p>
                    </div>
                    <span className={cn("px-2 py-0.5 rounded text-[10px] font-semibold", entry.role === "admin" ? "bg-[var(--danger-bg)] text-[var(--danger)]" : entry.role === "reviewer" ? "bg-[var(--warning-bg)] text-[var(--warning)]" : "bg-[var(--info-bg)] text-[var(--info)]")}>{entry.role}</span>
                    <span className={cn("text-[11px] font-medium", entry.active ? "text-[var(--success)]" : "text-[var(--danger)]")}>{entry.active ? "Active" : "Disabled"}</span>
                    <span className="text-[11px] text-[var(--text-muted)] tabular-nums">{new Date(entry.created_at).toLocaleDateString()}</span>
                    {isAdmin() && (
                    <div className="flex items-center gap-1">
                      <button onClick={() => handleRevoke(entry.id, !entry.active)} className="px-2 py-1 rounded text-[10px] font-semibold border border-[var(--border)] hover:bg-[var(--bg-hover)] transition-colors">{entry.active ? "Revoke" : "Enable"}</button>
                      <button onClick={() => handleDelete(entry.id)} className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--danger)] hover:bg-[var(--danger-bg)] transition-colors"><Trash2 className="w-3.5 h-3.5" /></button>
                    </div>
                  )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {tab === "audit" && (
        audit.length === 0 ? (
          <EmptyState icon={<Clock className="w-6 h-6" />} title="No audit entries" description="Actions will appear here." />
        ) : (
          <div className="space-y-1">
            {audit.slice(0, 50).map((entry) => (
              <div key={entry.id} className="flex items-center gap-3 px-3 py-2 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)] text-[12px]">
                <span className={cn("px-2 py-0.5 rounded text-[10px] font-semibold shrink-0", entry.action === "create" ? "bg-[var(--success-bg)] text-[var(--success)]" : entry.action === "delete" ? "bg-[var(--danger-bg)] text-[var(--danger)]" : entry.action === "approve" ? "bg-[var(--info-bg)] text-[var(--info)]" : entry.action === "reject" ? "bg-[var(--warning-bg)] text-[var(--warning)]" : "bg-[var(--bg-subtle)] text-[var(--text-muted)]")}>{entry.action}</span>
                <span className="text-[var(--text-secondary)] truncate flex-1"><span className="font-medium">{entry.actor_id}</span> <span className="text-[var(--text-muted)]">({entry.actor_role})</span> — {entry.detail || `${entry.resource}/${entry.resource_id}`}</span>
                <span className="text-[11px] text-[var(--text-muted)] tabular-nums shrink-0">{new Date(entry.created_at).toLocaleString()}</span>
              </div>
            ))}
          </div>
        )
      )}
    </div>
  );
}
