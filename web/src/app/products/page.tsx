"use client";

import { useEffect, useState, useCallback } from "react";
import { productApi, formatCents, formatDate } from "@/lib/api";
import type { Product } from "@/lib/api";
import Modal, { FormField, inputClass, selectClass, btnPrimary, btnSecondary } from "@/components/modal";

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [showStock, setShowStock] = useState<Product | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      const res = await productApi.list(1, 50);
      setProducts(res.products || []);
      setTotal(res.totalCount || 0);
    } catch {
      setError("Failed to load products");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold">Products</h1>
          <p className="text-xs text-text-muted mt-0.5">{total} items in catalog</p>
        </div>
        <button onClick={() => setShowCreate(true)} className={btnPrimary}>+ New Product</button>
      </div>

      {error && <p className="text-danger text-sm mb-4">{error}</p>}

      <div className="bg-surface-1 border border-border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-[11px] uppercase tracking-wider text-text-muted font-mono">
              <th className="text-left px-4 py-2.5">Product</th>
              <th className="text-left px-4 py-2.5">SKU</th>
              <th className="text-left px-4 py-2.5">Category</th>
              <th className="text-right px-4 py-2.5">Price</th>
              <th className="text-right px-4 py-2.5">Stock</th>
              <th className="text-left px-4 py-2.5">Created</th>
              <th className="text-right px-4 py-2.5">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {loading ? (
              [...Array(3)].map((_, i) => (
                <tr key={i}><td colSpan={7} className="px-4 py-3"><div className="h-4 bg-surface-2 rounded animate-pulse" /></td></tr>
              ))
            ) : products.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-text-muted">No products found</td></tr>
            ) : (
              products.map((p) => (
                <tr key={p.id} className="hover:bg-surface-2/50 transition">
                  <td className="px-4 py-3">
                    <p className="font-medium">{p.name}</p>
                    <p className="text-[11px] text-text-muted mt-0.5 truncate max-w-48">{p.description}</p>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-text-secondary">{p.sku}</td>
                  <td className="px-4 py-3">
                    {p.category && <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface-3 text-text-secondary border border-border">{p.category}</span>}
                  </td>
                  <td className="px-4 py-3 text-right font-mono">{formatCents(p.priceCents)}</td>
                  <td className="px-4 py-3 text-right">
                    <span className={`font-mono font-bold ${p.stockQuantity <= 5 ? (p.stockQuantity === 0 ? "text-danger" : "text-accent") : "text-text-primary"}`}>
                      {p.stockQuantity}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-text-muted">{formatDate(p.createdAt)}</td>
                  <td className="px-4 py-3 text-right">
                    <button onClick={() => setShowStock(p)} className="text-[11px] font-mono text-accent hover:text-accent-dim transition">
                      Stock
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <CreateProductModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={load} />
      <StockModal product={showStock} onClose={() => setShowStock(null)} onUpdated={load} />
    </div>
  );
}

function CreateProductModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [form, setForm] = useState({ sku: "", name: "", description: "", priceCents: "", stockQuantity: "", category: "" });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const set = (key: string, value: string) => setForm((prev) => ({ ...prev, [key]: value }));

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await productApi.create({
        sku: form.sku,
        name: form.name,
        description: form.description,
        priceCents: Math.round(parseFloat(form.priceCents) * 100),
        stockQuantity: parseInt(form.stockQuantity, 10),
        category: form.category,
      });
      setForm({ sku: "", name: "", description: "", priceCents: "", stockQuantity: "", category: "" });
      onCreated();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create product");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Product">
      <form onSubmit={handleSubmit}>
        {error && <p className="text-danger text-xs mb-3">{error}</p>}
        <div className="grid grid-cols-2 gap-3">
          <FormField label="SKU">
            <input className={inputClass} value={form.sku} onChange={(e) => set("sku", e.target.value)} placeholder="LAPTOP-001" required />
          </FormField>
          <FormField label="Category">
            <input className={inputClass} value={form.category} onChange={(e) => set("category", e.target.value)} placeholder="electronics" />
          </FormField>
        </div>
        <FormField label="Name">
          <input className={inputClass} value={form.name} onChange={(e) => set("name", e.target.value)} placeholder="MacBook Pro" required />
        </FormField>
        <FormField label="Description">
          <input className={inputClass} value={form.description} onChange={(e) => set("description", e.target.value)} placeholder="Apple laptop" />
        </FormField>
        <div className="grid grid-cols-2 gap-3">
          <FormField label="Price (JPY)">
            <input className={inputClass} type="number" step="1" value={form.priceCents} onChange={(e) => set("priceCents", e.target.value)} placeholder="1999" required />
          </FormField>
          <FormField label="Initial Stock">
            <input className={inputClass} type="number" value={form.stockQuantity} onChange={(e) => set("stockQuantity", e.target.value)} placeholder="10" required />
          </FormField>
        </div>
        <div className="flex justify-end gap-2 mt-4">
          <button type="button" onClick={onClose} className={btnSecondary}>Cancel</button>
          <button type="submit" disabled={submitting} className={btnPrimary}>{submitting ? "Creating..." : "Create"}</button>
        </div>
      </form>
    </Modal>
  );
}

function StockModal({ product, onClose, onUpdated }: { product: Product | null; onClose: () => void; onUpdated: () => void }) {
  const [delta, setDelta] = useState("");
  const [reason, setReason] = useState("STOCK_CHANGE_REASON_RESTOCK");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!product) return;
    setSubmitting(true);
    setError("");
    try {
      await productApi.updateStock({
        productId: product.id,
        delta: parseInt(delta, 10),
        reason,
      });
      setDelta("");
      onUpdated();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update stock");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={!!product} onClose={onClose} title={`Stock: ${product?.name || ""}`}>
      <div className="flex items-center gap-4 mb-4 p-3 bg-surface-1 rounded border border-border">
        <div>
          <p className="text-[11px] text-text-muted font-mono uppercase">Current Stock</p>
          <p className="text-2xl font-mono font-bold text-accent">{product?.stockQuantity ?? 0}</p>
        </div>
        <div>
          <p className="text-[11px] text-text-muted font-mono uppercase">SKU</p>
          <p className="text-sm font-mono text-text-secondary">{product?.sku}</p>
        </div>
      </div>
      <form onSubmit={handleSubmit}>
        {error && <p className="text-danger text-xs mb-3">{error}</p>}
        <div className="grid grid-cols-2 gap-3">
          <FormField label="Quantity (+/-)">
            <input className={inputClass} type="number" value={delta} onChange={(e) => setDelta(e.target.value)} placeholder="+5 or -3" required />
          </FormField>
          <FormField label="Reason">
            <select className={selectClass} value={reason} onChange={(e) => setReason(e.target.value)}>
              <option value="STOCK_CHANGE_REASON_RESTOCK">Restock</option>
              <option value="STOCK_CHANGE_REASON_ADJUSTMENT">Adjustment</option>
              <option value="STOCK_CHANGE_REASON_RETURN">Return</option>
            </select>
          </FormField>
        </div>
        <div className="flex justify-end gap-2 mt-4">
          <button type="button" onClick={onClose} className={btnSecondary}>Cancel</button>
          <button type="submit" disabled={submitting} className={btnPrimary}>{submitting ? "Updating..." : "Update Stock"}</button>
        </div>
      </form>
    </Modal>
  );
}
