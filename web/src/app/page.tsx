"use client";

import { useEffect, useState } from "react";
import { userApi, productApi, orderApi, paymentApi, formatCents, formatDate } from "@/lib/api";
import type { Order, Product } from "@/lib/api";

export default function Dashboard() {
  const [stats, setStats] = useState({ users: 0, products: 0, orders: 0, revenue: "0" });
  const [recentOrders, setRecentOrders] = useState<Order[]>([]);
  const [lowStock, setLowStock] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [users, products, orders, payments] = await Promise.all([
          userApi.list(1, 1).catch(() => ({ users: [], totalCount: 0 })),
          productApi.list(1, 100).catch(() => ({ products: [], totalCount: 0 })),
          orderApi.list(1, 5).catch(() => ({ orders: [], totalCount: 0 })),
          paymentApi.list(1, 100).catch(() => ({ payments: [], totalCount: 0 })),
        ]);

        const totalRevenue = (payments.payments || [])
          .filter((p) => p.status === "PAYMENT_STATUS_COMPLETED")
          .reduce((sum, p) => sum + parseInt(p.amountCents || "0", 10), 0);

        setStats({
          users: users.totalCount || 0,
          products: products.totalCount || 0,
          orders: orders.totalCount || 0,
          revenue: String(totalRevenue),
        });

        setRecentOrders(orders.orders || []);
        setLowStock((products.products || []).filter((p) => p.stockQuantity <= 5).slice(0, 5));
      } catch {
        // services might not be running
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const cards = [
    { label: "Users", value: stats.users, color: "text-info" },
    { label: "Products", value: stats.products, color: "text-accent" },
    { label: "Orders", value: stats.orders, color: "text-success" },
    { label: "Revenue", value: formatCents(stats.revenue), color: "text-accent" },
  ];

  if (loading) return <LoadingSkeleton />;

  return (
    <div>
      <h1 className="text-lg font-semibold mb-6">Dashboard</h1>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {cards.map(({ label, value, color }) => (
          <div key={label} className="bg-surface-1 border border-border rounded-lg p-4 hover:border-border-light transition">
            <p className="text-[11px] uppercase tracking-wider text-text-muted font-mono mb-2">{label}</p>
            <p className={`text-2xl font-bold font-mono ${color}`}>{value}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-3 gap-6">
        {/* Recent Orders */}
        <div className="col-span-2 bg-surface-1 border border-border rounded-lg">
          <div className="px-4 py-3 border-b border-border">
            <h2 className="text-sm font-semibold">Recent Orders</h2>
          </div>
          <div className="divide-y divide-border">
            {recentOrders.length === 0 ? (
              <p className="px-4 py-8 text-center text-text-muted text-sm">No orders yet</p>
            ) : (
              recentOrders.map((o) => (
                <div key={o.id} className="px-4 py-3 flex items-center justify-between hover:bg-surface-2/50 transition">
                  <div>
                    <p className="text-xs font-mono text-text-secondary">{o.id.slice(0, 8)}...</p>
                    <p className="text-[11px] text-text-muted mt-0.5">{formatDate(o.createdAt)}</p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-mono font-medium">{formatCents(o.totalCents)}</span>
                    <StatusBadge status={o.status} />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Low Stock Alert */}
        <div className="bg-surface-1 border border-border rounded-lg">
          <div className="px-4 py-3 border-b border-border">
            <h2 className="text-sm font-semibold">Low Stock Alert</h2>
          </div>
          <div className="divide-y divide-border">
            {lowStock.length === 0 ? (
              <p className="px-4 py-8 text-center text-text-muted text-sm">All products stocked</p>
            ) : (
              lowStock.map((p) => (
                <div key={p.id} className="px-4 py-3 flex items-center justify-between hover:bg-surface-2/50 transition">
                  <div>
                    <p className="text-sm">{p.name}</p>
                    <p className="text-[11px] text-text-muted font-mono">{p.sku}</p>
                  </div>
                  <span className={`text-sm font-mono font-bold ${p.stockQuantity === 0 ? "text-danger" : "text-accent"}`}>
                    {p.stockQuantity}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; cls: string }> = {
    ORDER_STATUS_PENDING: { label: "Pending", cls: "bg-amber-500/10 text-amber-400 border-amber-500/20" },
    ORDER_STATUS_CONFIRMED: { label: "Confirmed", cls: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
    ORDER_STATUS_SHIPPED: { label: "Shipped", cls: "bg-purple-500/10 text-purple-400 border-purple-500/20" },
    ORDER_STATUS_DELIVERED: { label: "Delivered", cls: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" },
    ORDER_STATUS_CANCELLED: { label: "Cancelled", cls: "bg-red-500/10 text-red-400 border-red-500/20" },
  };
  const { label, cls } = map[status] || { label: status, cls: "bg-surface-3 text-text-muted border-border" };
  return <span className={`text-[10px] font-mono font-medium px-2 py-0.5 rounded border ${cls}`}>{label}</span>;
}

function LoadingSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-6 w-32 bg-surface-2 rounded mb-6" />
      <div className="grid grid-cols-4 gap-4 mb-8">
        {[...Array(4)].map((_, i) => <div key={i} className="h-24 bg-surface-1 rounded-lg" />)}
      </div>
      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 h-64 bg-surface-1 rounded-lg" />
        <div className="h-64 bg-surface-1 rounded-lg" />
      </div>
    </div>
  );
}
