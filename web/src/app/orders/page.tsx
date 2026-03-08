"use client";

import { useEffect, useState, useCallback } from "react";
import { orderApi, productApi, userApi, formatCents, formatDate } from "@/lib/api";
import type { Order, Product, User } from "@/lib/api";
import Modal, { FormField, inputClass, selectClass, btnPrimary, btnSecondary } from "@/components/modal";

const STATUS_OPTIONS = [
  { value: "", label: "All" },
  { value: "ORDER_STATUS_PENDING", label: "Pending" },
  { value: "ORDER_STATUS_CONFIRMED", label: "Confirmed" },
  { value: "ORDER_STATUS_SHIPPED", label: "Shipped" },
  { value: "ORDER_STATUS_DELIVERED", label: "Delivered" },
  { value: "ORDER_STATUS_CANCELLED", label: "Cancelled" },
];

const STATUS_BADGE: Record<string, { label: string; cls: string }> = {
  ORDER_STATUS_PENDING: { label: "Pending", cls: "bg-amber-500/10 text-amber-400 border-amber-500/20" },
  ORDER_STATUS_CONFIRMED: { label: "Confirmed", cls: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  ORDER_STATUS_SHIPPED: { label: "Shipped", cls: "bg-purple-500/10 text-purple-400 border-purple-500/20" },
  ORDER_STATUS_DELIVERED: { label: "Delivered", cls: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" },
  ORDER_STATUS_CANCELLED: { label: "Cancelled", cls: "bg-red-500/10 text-red-400 border-red-500/20" },
};

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [detail, setDetail] = useState<Order | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await orderApi.list(1, 50, "", filter);
      setOrders(res.orders || []);
      setTotal(res.totalCount || 0);
    } catch { /* */ } finally { setLoading(false); }
  }, [filter]);

  useEffect(() => { load(); }, [load]);

  async function handleCancel(id: string) {
    if (!confirm("Cancel this order? Stock will be restored.")) return;
    try {
      await orderApi.cancel(id);
      load();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to cancel");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Orders</h1>
          <p className="text-xs text-text-muted mt-0.5">{total} total orders</p>
        </div>
        <div className="flex items-center gap-3">
          <select className={`${selectClass} w-40`} value={filter} onChange={(e) => setFilter(e.target.value)}>
            {STATUS_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
          <button onClick={() => setShowCreate(true)} className={btnPrimary}>+ New Order</button>
        </div>
      </div>

      <div className="bg-surface-1 border border-border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-[11px] uppercase tracking-wider text-text-muted font-mono">
              <th className="text-left px-4 py-2.5">Order ID</th>
              <th className="text-left px-4 py-2.5">User ID</th>
              <th className="text-left px-4 py-2.5">Status</th>
              <th className="text-right px-4 py-2.5">Items</th>
              <th className="text-right px-4 py-2.5">Total</th>
              <th className="text-left px-4 py-2.5">Created</th>
              <th className="text-right px-4 py-2.5">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {loading ? (
              [...Array(3)].map((_, i) => (
                <tr key={i}><td colSpan={7} className="px-4 py-3"><div className="h-4 bg-surface-2 rounded animate-pulse" /></td></tr>
              ))
            ) : orders.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-text-muted">No orders found</td></tr>
            ) : (
              orders.map((o) => {
                const badge = STATUS_BADGE[o.status] || { label: o.status, cls: "bg-surface-3 text-text-muted border-border" };
                return (
                  <tr key={o.id} className="hover:bg-surface-2/50 transition">
                    <td className="px-4 py-3">
                      <button onClick={() => setDetail(o)} className="font-mono text-xs text-accent hover:underline">{o.id.slice(0, 8)}...</button>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-text-secondary">{o.userId.slice(0, 8)}...</td>
                    <td className="px-4 py-3">
                      <span className={`text-[10px] font-mono font-medium px-2 py-0.5 rounded border ${badge.cls}`}>{badge.label}</span>
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-text-secondary">{(o.items || []).length}</td>
                    <td className="px-4 py-3 text-right font-mono font-medium">{formatCents(o.totalCents)}</td>
                    <td className="px-4 py-3 text-xs text-text-muted">{formatDate(o.createdAt)}</td>
                    <td className="px-4 py-3 text-right">
                      {o.status === "ORDER_STATUS_PENDING" && (
                        <button onClick={() => handleCancel(o.id)} className="text-[11px] font-mono text-danger hover:text-red-300 transition">
                          Cancel
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <CreateOrderModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={load} />
      <OrderDetailModal order={detail} onClose={() => setDetail(null)} />
    </div>
  );
}

function CreateOrderModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [users, setUsers] = useState<User[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [userId, setUserId] = useState("");
  const [items, setItems] = useState([{ productId: "", quantity: "1" }]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    userApi.list(1, 100).then((r) => setUsers(r.users || [])).catch(() => {});
    productApi.list(1, 100).then((r) => setProducts(r.products || [])).catch(() => {});
  }, [open]);

  function addItem() { setItems([...items, { productId: "", quantity: "1" }]); }
  function removeItem(i: number) { setItems(items.filter((_, idx) => idx !== i)); }
  function setItem(i: number, key: string, value: string) {
    setItems(items.map((item, idx) => idx === i ? { ...item, [key]: value } : item));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await orderApi.create({
        userId,
        items: items.map((it) => ({ productId: it.productId, quantity: parseInt(it.quantity, 10) })),
      });
      setUserId("");
      setItems([{ productId: "", quantity: "1" }]);
      onCreated();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create order");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Order">
      <form onSubmit={handleSubmit}>
        {error && <p className="text-danger text-xs mb-3">{error}</p>}
        <FormField label="User">
          <select className={selectClass} value={userId} onChange={(e) => setUserId(e.target.value)} required>
            <option value="">Select user...</option>
            {users.map((u) => <option key={u.id} value={u.id}>{u.name} ({u.email})</option>)}
          </select>
        </FormField>

        <div className="mb-3">
          <div className="flex items-center justify-between mb-1.5">
            <span className="text-[11px] uppercase tracking-wider text-text-muted font-mono">Items</span>
            <button type="button" onClick={addItem} className="text-[11px] font-mono text-accent hover:text-accent-dim">+ Add Item</button>
          </div>
          {items.map((item, i) => (
            <div key={i} className="flex gap-2 mb-2">
              <select className={`${selectClass} flex-1`} value={item.productId} onChange={(e) => setItem(i, "productId", e.target.value)} required>
                <option value="">Select product...</option>
                {products.map((p) => <option key={p.id} value={p.id}>{p.name} (stock: {p.stockQuantity})</option>)}
              </select>
              <input className={`${inputClass} w-20`} type="number" min="1" value={item.quantity} onChange={(e) => setItem(i, "quantity", e.target.value)} />
              {items.length > 1 && (
                <button type="button" onClick={() => removeItem(i)} className="text-text-muted hover:text-danger text-sm px-1">&times;</button>
              )}
            </div>
          ))}
        </div>

        <div className="flex justify-end gap-2 mt-4">
          <button type="button" onClick={onClose} className={btnSecondary}>Cancel</button>
          <button type="submit" disabled={submitting} className={btnPrimary}>{submitting ? "Creating..." : "Place Order"}</button>
        </div>
      </form>
    </Modal>
  );
}

function OrderDetailModal({ order, onClose }: { order: Order | null; onClose: () => void }) {
  if (!order) return null;
  const badge = STATUS_BADGE[order.status] || { label: order.status, cls: "bg-surface-3 text-text-muted border-border" };

  return (
    <Modal open={!!order} onClose={onClose} title={`Order ${order.id.slice(0, 8)}...`}>
      <div className="space-y-4">
        <div className="grid grid-cols-3 gap-3">
          <div className="p-3 bg-surface-1 rounded border border-border">
            <p className="text-[11px] text-text-muted font-mono uppercase mb-1">Status</p>
            <span className={`text-[10px] font-mono font-medium px-2 py-0.5 rounded border ${badge.cls}`}>{badge.label}</span>
          </div>
          <div className="p-3 bg-surface-1 rounded border border-border">
            <p className="text-[11px] text-text-muted font-mono uppercase mb-1">Total</p>
            <p className="text-lg font-mono font-bold text-accent">{formatCents(order.totalCents)}</p>
          </div>
          <div className="p-3 bg-surface-1 rounded border border-border">
            <p className="text-[11px] text-text-muted font-mono uppercase mb-1">Date</p>
            <p className="text-xs text-text-secondary">{formatDate(order.createdAt)}</p>
          </div>
        </div>
        <div>
          <p className="text-[11px] text-text-muted font-mono uppercase mb-2">Items</p>
          <div className="divide-y divide-border border border-border rounded">
            {(order.items || []).map((item) => (
              <div key={item.id} className="px-3 py-2 flex items-center justify-between">
                <div>
                  <p className="text-sm">{item.productName}</p>
                  <p className="text-[11px] text-text-muted font-mono">{item.productId.slice(0, 8)}...</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-mono">{formatCents(item.unitPriceCents)} x {item.quantity}</p>
                  <p className="text-xs font-mono text-accent">{formatCents(String(parseInt(item.unitPriceCents, 10) * item.quantity))}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </Modal>
  );
}
