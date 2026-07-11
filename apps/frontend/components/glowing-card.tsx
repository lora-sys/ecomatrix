"use client";

import { ReactNode } from "react";
import { motion } from "framer-motion";
import clsx from "clsx";

interface Props {
  children: ReactNode;
  tone?: "cyan" | "gold" | "violet" | "rose";
  label?: string;
  className?: string;
}

const RING: Record<NonNullable<Props["tone"]>, string> = {
  cyan: "ring-cyan-glow",
  gold: "ring-gold-glow",
  violet: "ring-violet-glow",
  rose: "ring-rose-glow",
};

export function GlowingCard({ children, tone = "cyan", label, className }: Props) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.18, ease: "easeOut" }}
      className={clsx(
        "rounded-md border border-hairline bg-panel p-4",
        RING[tone],
        className,
      )}
    >
      {label ? (
        <div className="mb-2 text-[10px] uppercase tracking-[0.18em] text-ink-muted font-mono">
          {label}
        </div>
      ) : null}
      {children}
    </motion.div>
  );
}
