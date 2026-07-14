"use client";

import { motion } from "framer-motion";
import type { ReactNode } from "react";

/**
 * ISS-FRONTEND: KPI tile replacement. Big numeral, hairline label,
 * gradient frame, hover wave on the accent strip. Adopts the four
 * tones (cyan, gold, violet, rose) the rest of the dashboard uses.
 */
export function BigMetric({
  label,
  value,
  unit,
  tone = "cyan",
  delta,
  format = (n) => Math.round(n).toLocaleString("en-US"),
  hint,
}: {
  label: string;
  value: number;
  unit?: string;
  tone?: "cyan" | "gold" | "violet" | "rose";
  delta?: number;
  format?: (n: number) => string;
  hint?: ReactNode;
}) {
  const accents: Record<string, string> = {
    cyan: "from-accent-cyan/30 via-accent-cyan/0",
    gold: "from-accent-gold/30 via-accent-gold/0",
    violet: "from-accent-violet/30 via-accent-violet/0",
    rose: "from-accent-rose/30 via-accent-rose/0",
  };
  const ring: Record<string, string> = {
    cyan: "ring-cyan-glow",
    gold: "ring-gold-glow",
    violet: "ring-violet-glow",
    rose: "ring-rose-glow",
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-10% 0px" }}
      transition={{ duration: 0.55, ease: "easeOut" }}
      className={
        "ecomatrix-frame group relative isolate overflow-hidden rounded-md border border-hairline/40 bg-gradient-to-br from-panel/60 to-canvas/40 p-5 " +
        ring[tone]
      }
    >
      <div
        aria-hidden
        className={
          "pointer-events-none absolute -left-1/4 top-0 h-full w-1/2 bg-gradient-to-r to-transparent " +
          accents[tone]
        }
      />
      <div className="flex items-baseline justify-between font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
        <span>{label}</span>
        {typeof delta === "number" && (
          <span className={delta >= 0 ? "text-accent-emerald" : "text-accent-rose"}>
            {delta >= 0 ? "▲" : "▼"} {Math.abs(delta).toFixed(1)}
          </span>
        )}
      </div>
      <div className="mt-3 flex items-baseline gap-2">
        <motion.div
          key={String(value)}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: "easeOut" }}
          className={
            "ecomatrix-numeral font-display text-[44px] font-medium leading-none tracking-[-0.03em] sm:text-[56px] " +
            `text-ink`
          }
        >
          {format(value)}
        </motion.div>
        {unit && (
          <span className="font-mono text-[11px] uppercase tracking-[0.18em] text-ink-muted">
            {unit}
          </span>
        )}
      </div>
      {hint && <div className="mt-3 font-mono text-[11px] text-ink-muted">{hint}</div>}
      {/* Animated bottom strip. */}
      <div className="absolute inset-x-0 bottom-0 h-[2px] overflow-hidden">
        <div
          className={
            "h-full w-1/3 origin-left animate-[ecomatrix-wave_4s_ease-in-out_infinite] " +
            (tone === "cyan"
              ? "bg-accent-cyan"
              : tone === "gold"
                ? "bg-accent-gold"
                : tone === "violet"
                  ? "bg-accent-violet"
                  : "bg-accent-rose")
          }
        />
      </div>
    </motion.div>
  );
}
