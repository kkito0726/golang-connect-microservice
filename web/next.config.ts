import type { NextConfig } from "next";

const userService = process.env.USER_SERVICE_URL || "http://localhost:8080";
const productService = process.env.PRODUCT_SERVICE_URL || "http://localhost:8081";
const orderService = process.env.ORDER_SERVICE_URL || "http://localhost:8082";
const paymentService = process.env.PAYMENT_SERVICE_URL || "http://localhost:8083";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      { source: "/api/user/:path*", destination: `${userService}/:path*` },
      { source: "/api/product/:path*", destination: `${productService}/:path*` },
      { source: "/api/order/:path*", destination: `${orderService}/:path*` },
      { source: "/api/payment/:path*", destination: `${paymentService}/:path*` },
    ];
  },
};

export default nextConfig;
