"use client";

import { useEffect, useState, useCallback } from "react";
import { paymentApi, orderApi, formatCents, formatDate } from "@/lib/api";
import type { Payment, Order } from "@/lib/api";
import Modal, { FormField, selectClass, btnPrimary, btnSecondary } from "@/components/modal";

const PAYMENT_STATUS: Record<string, { label: string; cls: string }> = {
  PAYMENT_STATUS_PENDING: { label: "Pending", cls: "bg-amber-500/10 text-amber-400 border-amber-500/20" },
  PAYMENT_STATUS_COMPLETED: { label: "Completed", cls: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" },
  PAYMENT_STATUS_FAILED: { label: "Failed", cls: "bg-red-500/10 text-red-400 border-red-500/20" },
  PAYMENT_STATUS_REFUNDED: { label: "Refunded", cls: "bg-purple-500/10 text-purple-400 border-purple-500/20" },
};

const METHOD_LABEL: Record<string, string> = {
  PAYMENT_METHOD_CREDIT_CARD: "Credit Card",
  PAYMENT_METHOD_BANK_TRANSFER: "Bank Transfer",
  PAYMENT_METHOD_WALLET: "Wallet",
};

export default function PaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await paymentApi.list(1, 50);
      setPayments(res.payments || []);
      setTotal(res.totalCount || 0);
    } catch { /* */ } finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function handleRefund(id: string) {
    if (!confirm("Refund this payment? The order will be cancelled and stock restored.")) return;
    try {
      await paymentApi.refund(id);
      load();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to refund");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Payments</h1>
          <p className="text-xs text-text-muted mt-0.5">{total} total payments</p>
        </div>
        <button onClick={() => setShowCreate(true)} className={btnPrimary}>+ New Payment</button>
      </div>

      <div className="bg-surface-1 border border-border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-[11px] uppercase tracking-wider text-text-muted font-mono">
              <th className="text-left px-4 py-2.5">Payment ID</th>
              <th className="text-left px-4 py-2.5">Order ID</th>
              <th className="text-left px-4 py-2.5">Method</th>
              <th className="text-left px-4 py-2.5">Status</th>
              <th className="text-right px-4 py-2.5">Amount</th>
              <th className="text-left px-4 py-2.5">Created</th>
              <th className="text-right px-4 py-2.5">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {loading ? (
              [...Array(3)].map((_, i) => (
                <tr key={i}><td colSpan={7} className="px-4 py-3"><div className="h-4 bg-surface-2 rounded animate-pulse" /></td></tr>
              ))
            ) : payments.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-text-muted">No payments found</td></tr>
            ) : (
              payments.map((p) => {
                const badge = PAYMENT_STATUS[p.status] || { label: p.status, cls: "bg-surface-3 text-text-muted border-border" };
                return (
                  <tr key={p.id} className="hover:bg-surface-2/50 transition">
                    <td className="px-4 py-3 font-mono text-xs text-text-secondary">{p.id.slice(0, 8)}...</td>
                    <td className="px-4 py-3 font-mono text-xs text-text-secondary">{p.orderId.slice(0, 8)}...</td>
                    <td className="px-4 py-3 text-xs">{METHOD_LABEL[p.method] || p.method}</td>
                    <td className="px-4 py-3">
                      <span className={`text-[10px] font-mono font-medium px-2 py-0.5 rounded border ${badge.cls}`}>{badge.label}</span>
                    </td>
                    <td className="px-4 py-3 text-right font-mono font-medium">{formatCents(p.amountCents)}</td>
                    <td className="px-4 py-3 text-xs text-text-muted">{formatDate(p.createdAt)}</td>
                    <td className="px-4 py-3 text-right">
                      {p.status === "PAYMENT_STATUS_COMPLETED" && (
                        <button onClick={() => handleRefund(p.id)} className="text-[11px] font-mono text-danger hover:text-red-300 transition">
                          Refund
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

      <CreatePaymentModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={load} />
    </div>
  );
}

function CreatePaymentModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [orders, setOrders] = useState<Order[]>([]);
  const [orderId, setOrderId] = useState("");
  const [method, setMethod] = useState("PAYMENT_METHOD_CREDIT_CARD");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    orderApi.list(1, 100, "", "ORDER_STATUS_PENDING")
      .then((r) => setOrders(r.orders || []))
      .catch(() => {});
  }, [open]);

  const selectedOrder = orders.find((o) => o.id === orderId);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedOrder) return;
    setSubmitting(true);
    setError("");
    try {
      await paymentApi.create({
        orderId,
        userId: selectedOrder.userId,
        method,
      });
      setOrderId("");
      onCreated();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create payment");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Payment">
      <form onSubmit={handleSubmit}>
        {error && <p className="text-danger text-xs mb-3">{error}</p>}
        <FormField label="Pending Order">
          <select className={selectClass} value={orderId} onChange={(e) => setOrderId(e.target.value)} required>
            <option value="">Select order...</option>
            {orders.map((o) => (
              <option key={o.id} value={o.id}>{o.id.slice(0, 8)}... - {formatCents(o.totalCents)}</option>
            ))}
          </select>
        </FormField>
        {selectedOrder && (
          <div className="p-3 mb-3 bg-surface-1 rounded border border-border">
            <p className="text-[11px] text-text-muted font-mono uppercase mb-1">Amount</p>
            <p className="text-xl font-mono font-bold text-accent">{formatCents(selectedOrder.totalCents)}</p>
          </div>
        )}
        <FormField label="Payment Method">
          <select className={selectClass} value={method} onChange={(e) => setMethod(e.target.value)}>
            <option value="PAYMENT_METHOD_CREDIT_CARD">Credit Card</option>
            <option value="PAYMENT_METHOD_BANK_TRANSFER">Bank Transfer</option>
            <option value="PAYMENT_METHOD_WALLET">Wallet</option>
          </select>
        </FormField>
        <div className="flex justify-end gap-2 mt-4">
          <button type="button" onClick={onClose} className={btnSecondary}>Cancel</button>
          <button type="submit" disabled={submitting || !orderId} className={btnPrimary}>{submitting ? "Processing..." : "Pay"}</button>
        </div>
      </form>
    </Modal>
  );
}
