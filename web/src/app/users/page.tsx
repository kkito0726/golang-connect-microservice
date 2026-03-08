"use client";

import { useEffect, useState, useCallback } from "react";
import { userApi, formatDate } from "@/lib/api";
import type { User } from "@/lib/api";
import Modal, { FormField, inputClass, selectClass, btnPrimary, btnSecondary } from "@/components/modal";

const ROLE_LABEL: Record<string, { label: string; cls: string }> = {
  ROLE_CUSTOMER: { label: "Customer", cls: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  ROLE_ADMIN: { label: "Admin", cls: "bg-amber-500/10 text-amber-400 border-amber-500/20" },
};

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await userApi.list(1, 50);
      setUsers(res.users || []);
      setTotal(res.totalCount || 0);
    } catch { /* */ } finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete user "${name}"?`)) return;
    try {
      await userApi.delete(id);
      load();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to delete");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Users</h1>
          <p className="text-xs text-text-muted mt-0.5">{total} registered users</p>
        </div>
        <button onClick={() => setShowCreate(true)} className={btnPrimary}>+ New User</button>
      </div>

      <div className="bg-surface-1 border border-border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-[11px] uppercase tracking-wider text-text-muted font-mono">
              <th className="text-left px-4 py-2.5">User</th>
              <th className="text-left px-4 py-2.5">Email</th>
              <th className="text-left px-4 py-2.5">Role</th>
              <th className="text-left px-4 py-2.5">ID</th>
              <th className="text-left px-4 py-2.5">Created</th>
              <th className="text-right px-4 py-2.5">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {loading ? (
              [...Array(3)].map((_, i) => (
                <tr key={i}><td colSpan={6} className="px-4 py-3"><div className="h-4 bg-surface-2 rounded animate-pulse" /></td></tr>
              ))
            ) : users.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-text-muted">No users found</td></tr>
            ) : (
              users.map((u) => {
                const role = ROLE_LABEL[u.role] || { label: u.role, cls: "bg-surface-3 text-text-muted border-border" };
                return (
                  <tr key={u.id} className="hover:bg-surface-2/50 transition">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2.5">
                        <div className="w-7 h-7 rounded-full bg-surface-3 border border-border flex items-center justify-center text-[11px] font-mono text-text-secondary">
                          {u.name.charAt(0).toUpperCase()}
                        </div>
                        <span className="font-medium">{u.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-text-secondary">{u.email}</td>
                    <td className="px-4 py-3">
                      <span className={`text-[10px] font-mono font-medium px-2 py-0.5 rounded border ${role.cls}`}>{role.label}</span>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-text-muted">{u.id.slice(0, 8)}...</td>
                    <td className="px-4 py-3 text-xs text-text-muted">{formatDate(u.createdAt)}</td>
                    <td className="px-4 py-3 text-right">
                      <button onClick={() => handleDelete(u.id, u.name)} className="text-[11px] font-mono text-danger hover:text-red-300 transition">
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <CreateUserModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={load} />
    </div>
  );
}

function CreateUserModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [form, setForm] = useState({ name: "", email: "", password: "", role: "ROLE_CUSTOMER" });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const set = (key: string, value: string) => setForm((prev) => ({ ...prev, [key]: value }));

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await userApi.create(form);
      setForm({ name: "", email: "", password: "", role: "ROLE_CUSTOMER" });
      onCreated();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create user");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New User">
      <form onSubmit={handleSubmit}>
        {error && <p className="text-danger text-xs mb-3">{error}</p>}
        <FormField label="Name">
          <input className={inputClass} value={form.name} onChange={(e) => set("name", e.target.value)} placeholder="Tanaka Taro" required />
        </FormField>
        <FormField label="Email">
          <input className={inputClass} type="email" value={form.email} onChange={(e) => set("email", e.target.value)} placeholder="tanaka@example.com" required />
        </FormField>
        <FormField label="Password">
          <input className={inputClass} type="password" value={form.password} onChange={(e) => set("password", e.target.value)} placeholder="********" required />
        </FormField>
        <FormField label="Role">
          <select className={selectClass} value={form.role} onChange={(e) => set("role", e.target.value)}>
            <option value="ROLE_CUSTOMER">Customer</option>
            <option value="ROLE_ADMIN">Admin</option>
          </select>
        </FormField>
        <div className="flex justify-end gap-2 mt-4">
          <button type="button" onClick={onClose} className={btnSecondary}>Cancel</button>
          <button type="submit" disabled={submitting} className={btnPrimary}>{submitting ? "Creating..." : "Create"}</button>
        </div>
      </form>
    </Modal>
  );
}
