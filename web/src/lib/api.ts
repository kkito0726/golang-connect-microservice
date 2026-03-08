const ENDPOINTS = {
  user: "/api/user",
  product: "/api/product",
  order: "/api/order",
  payment: "/api/payment",
} as const;

async function rpc<T>(
  service: keyof typeof ENDPOINTS,
  servicePath: string,
  method: string,
  body: Record<string, unknown> = {}
): Promise<T> {
  const url = `${ENDPOINTS[service]}/${servicePath}/${method}`;
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(err.message || res.statusText);
  }
  return res.json();
}

// User Service
export const userApi = {
  create: (data: { email: string; name: string; password: string; role: string }) =>
    rpc<{ user: User }>("user", "user.v1.UserService", "CreateUser", data),
  get: (id: string) =>
    rpc<{ user: User }>("user", "user.v1.UserService", "GetUser", { id }),
  list: (page = 1, pageSize = 20) =>
    rpc<{ users: User[]; totalCount: number }>("user", "user.v1.UserService", "ListUsers", { page, pageSize }),
  update: (data: { id: string; name: string; email: string }) =>
    rpc<{ user: User }>("user", "user.v1.UserService", "UpdateUser", data),
  delete: (id: string) =>
    rpc<object>("user", "user.v1.UserService", "DeleteUser", { id }),
};

// Product Service
export const productApi = {
  create: (data: { sku: string; name: string; description: string; priceCents: number; stockQuantity: number; category: string }) =>
    rpc<{ product: Product }>("product", "product.v1.ProductService", "CreateProduct", data),
  get: (id: string) =>
    rpc<{ product: Product }>("product", "product.v1.ProductService", "GetProduct", { id }),
  list: (page = 1, pageSize = 20, category = "") =>
    rpc<{ products: Product[]; totalCount: number }>("product", "product.v1.ProductService", "ListProducts", { page, pageSize, category }),
  update: (data: { id: string; name: string; description: string; priceCents: number; category: string }) =>
    rpc<{ product: Product }>("product", "product.v1.ProductService", "UpdateProduct", data),
  delete: (id: string) =>
    rpc<object>("product", "product.v1.ProductService", "DeleteProduct", { id }),
  updateStock: (data: { productId: string; delta: number; reason: string }) =>
    rpc<{ product: Product; movement: StockMovement }>("product", "product.v1.ProductService", "UpdateStock", data),
  getStockLevel: (productId: string) =>
    rpc<{ productId: string; stockQuantity: number; recentMovements: StockMovement[] }>("product", "product.v1.ProductService", "GetStockLevel", { productId }),
};

// Order Service
export const orderApi = {
  create: (data: { userId: string; items: { productId: string; quantity: number }[] }) =>
    rpc<{ order: Order }>("order", "order.v1.OrderService", "CreateOrder", data),
  get: (id: string) =>
    rpc<{ order: Order }>("order", "order.v1.OrderService", "GetOrder", { id }),
  list: (page = 1, pageSize = 20, userId = "", status = "") =>
    rpc<{ orders: Order[]; totalCount: number }>("order", "order.v1.OrderService", "ListOrders", {
      page, pageSize,
      ...(userId && { userId }),
      ...(status && { status }),
    }),
  cancel: (id: string) =>
    rpc<{ order: Order }>("order", "order.v1.OrderService", "CancelOrder", { id }),
  updateStatus: (id: string, status: string) =>
    rpc<{ order: Order }>("order", "order.v1.OrderService", "UpdateOrderStatus", { id, status }),
};

// Payment Service
export const paymentApi = {
  create: (data: { orderId: string; userId: string; method: string }) =>
    rpc<{ payment: Payment }>("payment", "payment.v1.PaymentService", "CreatePayment", data),
  get: (id: string) =>
    rpc<{ payment: Payment }>("payment", "payment.v1.PaymentService", "GetPayment", { id }),
  list: (page = 1, pageSize = 20, orderId = "", userId = "") =>
    rpc<{ payments: Payment[]; totalCount: number }>("payment", "payment.v1.PaymentService", "ListPayments", {
      page, pageSize,
      ...(orderId && { orderId }),
      ...(userId && { userId }),
    }),
  refund: (id: string) =>
    rpc<{ payment: Payment }>("payment", "payment.v1.PaymentService", "RefundPayment", { id }),
};

// Types
export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  description: string;
  priceCents: string;
  stockQuantity: number;
  category: string;
  createdAt: string;
  updatedAt: string;
}

export interface StockMovement {
  id: string;
  productId: string;
  delta: number;
  reason: string;
  referenceId: string;
  createdAt: string;
}

export interface OrderItem {
  id: string;
  productId: string;
  productName: string;
  quantity: number;
  unitPriceCents: string;
}

export interface Order {
  id: string;
  userId: string;
  status: string;
  items: OrderItem[];
  totalCents: string;
  createdAt: string;
  updatedAt: string;
}

export interface Payment {
  id: string;
  orderId: string;
  userId: string;
  amountCents: string;
  status: string;
  method: string;
  createdAt: string;
  updatedAt: string;
}

export function formatCents(cents: string | number): string {
  const num = typeof cents === "string" ? parseInt(cents, 10) : cents;
  return new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" }).format(num / 100);
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleString("ja-JP", {
    year: "numeric", month: "2-digit", day: "2-digit",
    hour: "2-digit", minute: "2-digit",
  });
}
