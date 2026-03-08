"use client";

import { useEffect, useRef } from "react";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export default function Modal({ open, onClose, title, children }: ModalProps) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    if (open) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [open]);

  if (!open) return null;

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      className="fixed inset-0 z-50 m-auto w-full max-w-lg rounded-lg bg-surface-2 border border-border text-text-primary backdrop:bg-black/60 backdrop:backdrop-blur-sm p-0"
    >
      <div className="flex items-center justify-between px-5 py-3.5 border-b border-border">
        <h2 className="text-sm font-semibold">{title}</h2>
        <button onClick={onClose} className="text-text-muted hover:text-text-primary text-lg leading-none">&times;</button>
      </div>
      <div className="p-5">{children}</div>
    </dialog>
  );
}

export function FormField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block mb-3">
      <span className="text-[11px] uppercase tracking-wider text-text-muted font-mono block mb-1.5">{label}</span>
      {children}
    </label>
  );
}

export const inputClass =
  "w-full px-3 py-2 bg-surface-1 border border-border rounded text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/20 transition";

export const selectClass =
  "w-full px-3 py-2 bg-surface-1 border border-border rounded text-sm text-text-primary focus:outline-none focus:border-accent/50 transition appearance-none";

export const btnPrimary =
  "px-4 py-2 bg-accent text-surface-0 text-sm font-semibold rounded hover:bg-accent-dim transition disabled:opacity-40 disabled:cursor-not-allowed";

export const btnSecondary =
  "px-4 py-2 bg-surface-3 text-text-secondary text-sm font-medium rounded hover:bg-border hover:text-text-primary transition";
