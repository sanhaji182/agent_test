"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section } from "@/components/ui/section";
import {
  listUsers,
  createUser,
  updateUser,
  deleteUser,
  type UserEntry,
} from "@/lib/api";
import {
  Plus, Loader2, CheckCircle2, XCircle, Trash2, PencilLine,
  ShieldCheck, ShieldOff,
} from "lucide-react";

const ROLES = ["admin", "reviewer", "viewer", "api_client"];

export default function UsersPage() {
  const [users, setUsers] = useState<UserEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ email: "", password: "", name: "", role: "viewer" });
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");
  const [actionMsg, setActionMsg] = useState("");

  const load = async () => {
    try {
      setUsers(await listUsers());
    } catch {
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const openAdd = () => {
    setEditingId(null);
    setForm({ email: "", password: "", name: "", role: "viewer" });
    setFormError("");
    setShowForm(true);
  };

  const openEdit = (u: UserEntry) => {
    setEditingId(u.id);
    setForm({ email: u.email, password: "", name: u.name, role: u.role });
    setFormError("");
    setShowForm(true);
  };

  const handleSave = async () => {
    setSaving(true);
    setFormError("");
    try {
      if (editingId) {
        const data: { name?: string; role?: string; new_password?: string } = {
          name: form.name,
          role: form.role,
        };
        if (form.password) data.new_password = form.password;
        await updateUser(editingId, data);
      } else {
        if (!form.email.trim() || !form.password) {
          setFormError("Email dan password wajib diisi.");
          setSaving(false);
          return;
        }
        await createUser({
          email: form.email.trim(),
          password: form.password,
          name: form.name,
          role: form.role,
        });
      }
      setShowForm(false);
      await load();
      flash(editingId ? "User diperbarui." : "User dibuat.");
    } catch (e) {
      setFormError(e instanceof Error ? e.message : "gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  const handleToggleActive = async (u: UserEntry) => {
    await updateUser(u.id, { is_active: !u.is_active });
    await load();
    flash(u.is_active ? "User dinonaktifkan." : "User diaktifkan.");
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Hapus user ini?")) return;
    await deleteUser(id);
    await load();
    flash("User dihapus.");
  };

  const flash = (msg: string) => {
    setActionMsg(msg);
    setTimeout(() => setActionMsg(""), 2500);
  };

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-sm text-[var(--text-muted)] py-8 justify-center">
        <Loader2 className="w-4 h-4 animate-spin" /> Loading users…
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">User Management</h1>
        <p className="text-sm text-[var(--text-muted)]">
          Kelola user yang bisa login ke aplikasi (email + password).
        </p>
      </div>

      <Section title="Users">
        <div className="flex items-center justify-between mb-4">
          <span className="text-sm text-[var(--text-muted)]">{users.length} user</span>
          <Button size="sm" onClick={openAdd}>
            <Plus className="w-3.5 h-3.5" />
            <span className="ml-1">Add User</span>
          </Button>
        </div>

        {actionMsg && (
          <div className="flex items-center gap-2 text-sm text-green-700 bg-green-50 border border-green-200 rounded-lg px-3 py-2 mb-4">
            <CheckCircle2 className="w-4 h-4" /> {actionMsg}
          </div>
        )}

        <div className="overflow-x-auto border border-[var(--border-default)] rounded-lg">
          <table className="w-full text-left text-sm">
            <thead className="bg-[var(--bg-subtle)]">
              <tr>
                <th className="px-4 py-2.5 font-medium text-[var(--text-muted)]">Email</th>
                <th className="px-4 py-2.5 font-medium text-[var(--text-muted)]">Name</th>
                <th className="px-4 py-2.5 font-medium text-[var(--text-muted)]">Role</th>
                <th className="px-4 py-2.5 font-medium text-[var(--text-muted)]">Status</th>
                <th className="px-4 py-2.5 font-medium text-[var(--text-muted)] text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-t border-[var(--border-default)]">
                  <td className="px-4 py-3 font-medium text-[var(--text-primary)]">{u.email}</td>
                  <td className="px-4 py-3 text-[var(--text-secondary)]">{u.name || "-"}</td>
                  <td className="px-4 py-3"><RoleBadge role={u.role} /></td>
                  <td className="px-4 py-3">
                    {u.is_active ? (
                      <span className="inline-flex items-center gap-1 text-green-600 text-xs">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-[var(--text-muted)] text-xs">
                        <XCircle className="w-3.5 h-3.5" /> Inactive
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => openEdit(u)}>
                        <PencilLine className="w-3.5 h-3.5" />
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => handleToggleActive(u)}>
                        {u.is_active ? <ShieldOff className="w-3.5 h-3.5" /> : <ShieldCheck className="w-3.5 h-3.5" />}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(u.id)}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Section>

      {/* Form modal */}
      {showForm && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
          onClick={() => setShowForm(false)}
        >
          <div
            className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl p-6 w-full max-w-md space-y-4"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-lg font-semibold text-[var(--text-primary)]">
              {editingId ? "Edit User" : "Add User"}
            </h2>

            <Input
              id="u-email"
              type="email"
              label="Email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="user@example.com"
              disabled={!!editingId}
            />
            <Input
              id="u-name"
              label="Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Nama lengkap"
            />
            <Input
              id="u-password"
              type="password"
              label={editingId ? "New Password (kosongkan jika tidak diubah)" : "Password"}
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="minimal 6 karakter"
            />

            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Role</label>
              <select
                className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)]"
                value={form.role}
                onChange={(e) => setForm({ ...form, role: e.target.value })}
              >
                {ROLES.map((r) => (
                  <option key={r} value={r}>{r}</option>
                ))}
              </select>
            </div>

            {formError && (
              <div className="flex items-center gap-2 text-sm text-red-700 bg-red-50 border border-red-200 rounded-lg px-3 py-2">
                <XCircle className="w-4 h-4" /> {formError}
              </div>
            )}

            <div className="flex items-center gap-3 pt-1">
              <Button onClick={handleSave} disabled={saving}>
                {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle2 className="w-4 h-4" />}
                <span className="ml-1">{saving ? "Saving…" : "Save"}</span>
              </Button>
              <Button variant="secondary" onClick={() => setShowForm(false)}>Cancel</Button>
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
    reviewer: "bg-amber-100 text-amber-700 border-amber-200",
    viewer: "bg-blue-100 text-blue-700 border-blue-200",
    api_client: "bg-gray-100 text-gray-700 border-gray-200",
  };
  return (
    <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium border ${colors[role] || colors.viewer}`}>
      {role}
    </span>
  );
}
