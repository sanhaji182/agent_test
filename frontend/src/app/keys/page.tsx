"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { KeyRound, Plus, Trash2, Copy, Shield, Clock, CheckCircle2, AlertTriangle } from "lucide-react";

export default function KeysPage() {
  const [keys, setKeys] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [selectedRole, setSelectedRole] = useState("viewer");

  useEffect(() => {
    // TODO: Fetch real API keys from database
    setTimeout(() => {
      setKeys([
        { 
          id: "key_1", 
          name: "Production Deployments", 
          role: "api_client", 
          prefix: "sk_prod_...",
          createdAt: "2024-01-15T10:30:00Z",
          lastUsed: "2026-07-31T14:20:00Z"
        },
        { 
          id: "key_2", 
          name: "CI/CD Pipeline", 
          role: "reviewer", 
          prefix: "sk_ci_...",
          createdAt: "2024-02-20T09:15:00Z",
          lastUsed: "2026-07-31T12:00:00Z"
        },
      ]);
      setLoading(false);
    }, 500);
  }, []);

  const handleCreateKey = () => {
    if (!newKeyName.trim()) return;
    
    const newKey = {
      id: `key_${Date.now()}`,
      name: newKeyName,
      role: selectedRole,
      prefix: "sk_new_...",
      createdAt: new Date().toISOString(),
      lastUsed: null,
    };
    
    setKeys(prev => [newKey, ...prev]);
    setNewKeyName("");
    setShowModal(false);
  };

  const handleDeleteKey = (id: string) => {
    if (confirm("Are you sure you want to delete this API key? This action cannot be undone.")) {
      setKeys(prev => prev.filter(key => key.id !== id));
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    alert("Copied to clipboard!");
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

      {/* Global API Key Display */}
      <Section title="Global API Key">
        <div className="flex items-center gap-3 p-4 rounded-lg bg-gray-50 border border-gray-200">
          <KeyRound className="w-6 h-6 text-[var(--accent)]" />
          <div className="flex-1 min-w-0">
            <code className="text-sm text-[var(--text-secondary)] truncate block">{apiKey}</code>
            <p className="text-xs text-[var(--text-muted)] mt-1">Environment-level authentication key</p>
          </div>
          <Button variant="secondary" size="sm" onClick={() => copyToClipboard(apiKey)}>
            <Copy className="w-4 h-4" />
          </Button>
        </div>
      </Section>

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
                  <Th>Name</Th>
                  <Th>Role</Th>
                  <Th>Prefix</Th>
                  <Th>Created</Th>
                  <Th>Last Used</Th>
                  <Th align="right">Actions</Th>
                </tr>
              </thead>
              <tbody>
                {keys.map((key) => (
                  <Tr key={key.id} hover>
                    <Td className="font-medium">
                      <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-[var(--text-muted)]" />
                        <span>{key.name}</span>
                      </div>
                    </Td>
                    <Td>
                      <RoleBadge role={key.role} />
                    </Td>
                    <Td className="font-mono text-xs text-[var(--text-muted)]">
                      {key.prefix}
                    </Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                      {formatDate(key.createdAt)}
                    </Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                      {key.lastUsed ? formatDate(key.lastUsed) : "Never"}
                    </Td>
                    <Td align="right" className="space-x-2">
                      <Button variant="ghost" size="sm" onClick={() => copyToClipboard(`${key.prefix}abc123`)}>
                        <Copy className="w-3.5 h-3.5" />
                      </Button>
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        onClick={() => handleDeleteKey(key.id)}
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
                  <option value="api_client">API Client - Read Only</option>
                </select>
                <p className="text-xs text-[var(--text-muted)] mt-1">
                  The assigned role determines what actions this key can perform.
                </p>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t">
                <Button variant="secondary" onClick={() => setShowModal(false)}>Cancel</Button>
                <Button onClick={handleCreateKey} disabled={!newKeyName.trim()}>
                  Create Key
                </Button>
              </div>
            </div>
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
