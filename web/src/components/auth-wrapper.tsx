"use client";

import { useState, useEffect } from "react";
import { getToken, setToken, clearToken } from "@/lib/auth";
import { userApi } from "@/lib/api";
import Sidebar from "@/components/sidebar";
import { inputClass, btnPrimary } from "@/components/modal";

export default function AuthWrapper({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState("");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setTokenState(getToken());
    setMounted(true);
  }, []);

  // Avoid hydration mismatch by rendering nothing until mounted on client
  if (!mounted) return null;

  if (!token) {
    return (
      <LoginScreen
        onLogin={(t) => {
          setToken(t);
          setTokenState(t);
        }}
      />
    );
  }

  return (
    <>
      <Sidebar
        onLogout={() => {
          clearToken();
          setTokenState("");
        }}
      />
      <main className="ml-56 min-h-screen">
        <div className="px-8 py-6">{children}</div>
      </main>
    </>
  );
}

function LoginScreen({ onLogin }: { onLogin: (token: string) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await userApi.login({ email, password });
      onLogin(res.accessToken);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="w-full max-w-sm">
        <div className="flex items-center gap-2.5 justify-center mb-8">
          <div className="w-8 h-8 bg-accent rounded flex items-center justify-center">
            <svg width="16" height="16" viewBox="0 0 14 14" fill="none">
              <rect x="1" y="1" width="5" height="5" rx="1" fill="#0a0a0c" />
              <rect x="8" y="1" width="5" height="5" rx="1" fill="#0a0a0c" />
              <rect x="1" y="8" width="5" height="5" rx="1" fill="#0a0a0c" />
              <rect x="8" y="8" width="5" height="5" rx="1" fill="#0a0a0c" opacity="0.4" />
            </svg>
          </div>
          <span className="text-base font-semibold">Inventory</span>
        </div>

        <div className="bg-surface-1 border border-border rounded-xl p-6">
          <h1 className="text-sm font-semibold mb-1">Sign in</h1>
          <p className="text-xs text-text-muted mb-5">Enter your credentials to continue</p>

          <form onSubmit={handleSubmit} className="space-y-3">
            {error && (
              <div className="text-xs text-danger bg-red-500/10 border border-red-500/20 rounded px-3 py-2">
                {error}
              </div>
            )}
            <div>
              <label className="block text-[11px] font-mono uppercase tracking-wider text-text-muted mb-1">
                Email
              </label>
              <input
                className={inputClass}
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="user@example.com"
                required
                autoFocus
              />
            </div>
            <div>
              <label className="block text-[11px] font-mono uppercase tracking-wider text-text-muted mb-1">
                Password
              </label>
              <input
                className={inputClass}
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className={`${btnPrimary} w-full justify-center mt-2`}
            >
              {loading ? "Signing in..." : "Sign in"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
