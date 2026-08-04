"use client";

import { useEffect, useState } from "react";
import {
  listAPIKeys,
  createAPIKey,
  revokeAPIKey,
  deleteAPIKey,
  type APIKeyEntry,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { KeyRound, Plus, Trash2, Copy, Shield, Ban, CheckCircle2, AlertTriangle } from "lucide-react";

export default function KeysPage() {
  const [keys, setKeys] = useState<APIKeyEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [selectedRole, setSelectedRole] = useState("viewer");
  const [creating, setCreating] = useState(false);
  const [createdKey, setCreatedKey] = useState<APIKeyEntry | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  const loadKeys = () => {
    return listAPIKeys()
      .then((k) => setKeys(k || []))
      .catch((e) => setError(e.message));
  };

  useEffect(() => {
    loadKeys().finally(() => setLoading(false));
  }, []);

  const handleCreateKey = async () => {
    if (!newKeyName.trim()) return;
    setCreating(true);
    setError(null);
    try {
      const entry = await createAPIKey(newKeyName.trim(), selectedRole);
      // Plain key is only returned on creation — show it once to the user.
      setCreatedKey(entry);
      setNewKeyName("");
      await loadKeys();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setCreating(false);
    }
  };

  const handleRevokeKey = async (key: APIKeyEntry) => {
    const action = key.active ? "revoke" : "re-enable";
    if (!confirm(`Are you sure you want to ${action} the key "${key.label}"?`)) return;
    setBusyId(key.id);
    setError(null);
    try {
      await revokeAPIKey(key.id, !key.active);
      await loadKeys();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const handleDeleteKey = async (key: APIKeyEntry) => {
    if (!confirm(`Are you sure you want to delete the key "${key.label}"? This action cannot be undone.`)) return;
    setBusyId(key.id);
    setError(null);
    try {
      await deleteAPIKey(key.id);
      await loadKeys();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    alert("Copied to clipboard!");
  };

  const closeModal = () => {
    setShowModal(false);
    setCreatedKey(null);
  };

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">API Keys & Audit</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Manage access tokens and view audit logs</p>
        </div>
        <Button onClick={() => setShowModal(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create API Key
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {/* API Keys List */}
      <Section title="Application API Keys">
        {keys.length === 0 ? (
          <EmptyState
            icon={<KeyRound className="w-8 h-8" />}
            title="No API keys created"
            description="Generate API keys for specific users or applications to control access to your testing platform."
            action={
              <Button onClick={() => setShowModal(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Create First Key
              </Button>
            }
          />
        ) : (
          <TableContainer>
            <table className="w-full text-left">
              <thead className="bg-gray-50 border-b border-[var(--border-default)]">
                <tr>
                  <Th>Label</Th>
                  <Th>Role</Th>
                  <Th>Status</Th>
                  <Th>Created</Th>
                  <Th>Created By</Th>
                  <Th align="right">Actions</Th>
                </tr>
              </thead>
              <tbody>
                {keys.map((key) => (
                  <Tr key={key.id} hover>
                    <Td className="font-medium">
                      <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-[var(--text-muted)]" />
                        <span>{key.label}</span>
                      </div>
                    </Td>
                    <Td>
                      <RoleBadge role={key.role} />
                    </Td>
                    <Td>
                      {key.active ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border bg-green-100 text-green-700 border-green-200">
                          <CheckCircle2 className="w-3 h-3" />
                          Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border bg-gray-100 text-gray-600 border-gray-200">
                          <Ban className="w-3 h-3" />
                          Revoked
                        </span>
                      )}
                    </Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                      {formatDate(key.created_at)}
                    </Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                      {key.created_by || "-"}
                    </Td>
                    <Td align="right" className="space-x-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={busyId === key.id}
                        title={key.active ? "Revoke key" : "Re-enable key"}
                        className={key.active ? "text-yellow-600 hover:text-yellow-700 hover:bg-yellow-50" : "text-green-600 hover:text-green-700 hover:bg-green-50"}
                        onClick={() => handleRevokeKey(key)}
                      >
                        {key.active ? <Ban className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={busyId === key.id}
                        title="Delete key"
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        onClick={() => handleDeleteKey(key)}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </TableContainer>
        )}
      </Section>

      {/* Security Info */}
      <div className="rounded-lg bg-blue-50 border border-blue-200 p-4">
        <div className="flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-blue-600 shrink-0 mt-0.5" />
          <div>
            <h3 className="text-sm font-semibold text-blue-800 mb-1">Security Best Practices</h3>
            <ul className="text-xs text-blue-700 space-y-1 list-disc list-inside">
              <li>Rotate keys regularly - at least every 90 days</li>
              <li>Avoid committing API keys to version control</li>
              <li>Use environment variables for key storage</li>
              <li>Apply principle of least privilege when assigning roles</li>
              <li>Monitor key usage via audit logs</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Create Key Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            {createdKey?.key ? (
              <>
                <h2 className="text-lg font-semibold mb-4">API Key Created</h2>
                <p className="text-sm text-[var(--text-muted)] mb-3">
                  Copy this key now. For security reasons, it will not be shown again.
                </p>
                <div className="flex items-center gap-2 p-3 rounded-lg bg-gray-50 border border-gray-200">
                  <code className="flex-1 text-xs font-mono break-all text-[var(--text-secondary)]">{createdKey.key}</code>
                  <Button variant="secondary" size="sm" onClick={() => copyToClipboard(createdKey.key!)}>
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
                <div className="flex justify-end pt-4 mt-4 border-t">
                  <Button onClick={closeModal}>Done</Button>
                </div>
              </>
            ) : (
              <>
                <h2 className="text-lg font-semibold mb-4">Create New API Key</h2>

                <div className="space-y-4">
                  <div>
                    <label htmlFor="key-name" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                      Key Name <span className="text-red-500">*</span>
                    </label>
                    <Input
                      id="key-name"
                      placeholder="e.g., Production Deployments"
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                      autoFocus
                    />
                  </div>

                  <div>
                    <label htmlFor="role-select" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                      Role <span className="text-red-500">*</span>
                    </label>
                    <select
                      id="role-select"
                      className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                      value={selectedRole}
                      onChange={(e) => setSelectedRole(e.target.value)}
                    >
                      <option value="admin">Admin - Full Access</option>
                      <option value="reviewer">Reviewer - View & Review</option>
                      <option value="viewer">Viewer - Read Only</option>
                      <option value="api_client">API Client - API Access</option>
                    </select>
                    <p className="text-xs text-[var(--text-muted)] mt-1">
                      The assigned role determines what actions this key can perform.
                    </p>
                  </div>

                  <div className="flex justify-end gap-3 pt-4 border-t">
                    <Button variant="secondary" onClick={closeModal}>Cancel</Button>
                    <Button onClick={handleCreateKey} disabled={!newKeyName.trim() || creating}>
                      {creating ? "Creating..." : "Create Key"}
                    </Button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function RoleBadge({ role }: { role: string }) {
  const colors: Record<string, string> = {
    admin: "bg-red-100 text-red-700 border-red-200",
    reviewer: "bg-yellow-100 text-yellow-700 border-yellow-200",
    viewer: "bg-gray-100 text-gray-700 border-gray-200",
    api_client: "bg-blue-100 text-blue-700 border-blue-200",
  };

  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${colors[role] || "bg-gray-100 text-gray-700 border-gray-200"}`}>
      {role.replace("_", " ").replace(/\b\w/g, l => l.toUpperCase())}
    </span>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}
