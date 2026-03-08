"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const NAV = [
  { href: "/", label: "Dashboard", icon: GridIcon },
  { href: "/products", label: "Products", icon: BoxIcon },
  { href: "/orders", label: "Orders", icon: ClipboardIcon },
  { href: "/payments", label: "Payments", icon: CreditCardIcon },
  { href: "/users", label: "Users", icon: UsersIcon },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed top-0 left-0 z-40 h-screen w-56 flex flex-col bg-surface-1 border-r border-border">
      <div className="flex items-center gap-2.5 px-5 h-14 border-b border-border">
        <div className="w-7 h-7 bg-accent rounded flex items-center justify-center">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <rect x="1" y="1" width="5" height="5" rx="1" fill="#0a0a0c" />
            <rect x="8" y="1" width="5" height="5" rx="1" fill="#0a0a0c" />
            <rect x="1" y="8" width="5" height="5" rx="1" fill="#0a0a0c" />
            <rect x="8" y="8" width="5" height="5" rx="1" fill="#0a0a0c" opacity="0.4" />
          </svg>
        </div>
        <span className="text-sm font-semibold tracking-tight text-text-primary">
          Inventory
        </span>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {NAV.map(({ href, label, icon: Icon }) => {
          const active = href === "/" ? pathname === "/" : pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-[13px] font-medium transition-all duration-150
                ${active
                  ? "bg-accent-glow text-accent border border-accent/20"
                  : "text-text-secondary hover:text-text-primary hover:bg-surface-2 border border-transparent"
                }`}
            >
              <Icon size={15} />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="px-4 py-3 border-t border-border">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-success animate-pulse" />
          <span className="text-[11px] text-text-muted font-mono">4 services online</span>
        </div>
      </div>
    </aside>
  );
}

function GridIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
      <rect x="2" y="2" width="5" height="5" rx="1" /><rect x="9" y="2" width="5" height="5" rx="1" />
      <rect x="2" y="9" width="5" height="5" rx="1" /><rect x="9" y="9" width="5" height="5" rx="1" />
    </svg>
  );
}

function BoxIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M2 5l6-3 6 3-6 3-6-3z" /><path d="M2 5v6l6 3V8" /><path d="M14 5v6l-6 3V8" />
    </svg>
  );
}

function ClipboardIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
      <rect x="3" y="2" width="10" height="12" rx="1.5" /><path d="M6 2V1.5a1 1 0 011-1h2a1 1 0 011 1V2" />
      <path d="M6 6h4M6 9h3" />
    </svg>
  );
}

function CreditCardIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
      <rect x="1" y="3" width="14" height="10" rx="2" /><path d="M1 7h14" /><path d="M4 10h3" />
    </svg>
  );
}

function UsersIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
      <circle cx="6" cy="5" r="2.5" /><path d="M1.5 14c0-3 2-4.5 4.5-4.5s4.5 1.5 4.5 4.5" />
      <circle cx="11" cy="5.5" r="2" /><path d="M12 9.5c1.5.5 2.5 1.5 2.5 3.5" />
    </svg>
  );
}
