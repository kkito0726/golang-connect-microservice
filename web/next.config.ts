import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      { source: "/api/user/:path*", destination: "http://localhost:8080/:path*" },
      { source: "/api/product/:path*", destination: "http://localhost:8081/:path*" },
      { source: "/api/order/:path*", destination: "http://localhost:8082/:path*" },
      { source: "/api/payment/:path*", destination: "http://localhost:8083/:path*" },
    ];
  },
};

export default nextConfig;
