"use client";

import { motion } from "framer-motion";
import type { ReactNode } from "react";

/**
 * ISS-FRONTEND: The dramatic page header. Layered type, kicker,
 * gradient headline, secondary sub-headline, and a small parallelogram
 * mark. Use as the page-level introduction on every page.
 */
export function Masthead({
  kicker,
  headline,
  subhead,
  glyph,
  rightSlot,
}: {
  kicker?: string;
  headline: ReactNode;
  subhead?: ReactNode;
  glyph?: string;
  rightSlot?: ReactNode;
}) {
  return (
    <header className="relative isolate overflow-hidden rounded-md border border-hairline/50 bg-gradient-to-br from-panel/60 via-panel/30 to-canvas/40 p-6 sm:p-10">
      {/* Hairline radial accent in the corner. */}
      <div
        aria-hidden
        className="pointer-events-none absolute -right-32 -top-32 h-[60vmin] w-[60vmin] rounded-full opacity-50"
        style={{
          background:
            "radial-gradient(circle at 50% 50%, rgba(34,211,238,0.30) 0%, rgba(34,211,238,0) 60%)",
        }}
      />

      <div className="relative grid gap-6 lg:grid-cols-[3fr_2fr] lg:items-end">
        <div>
          {kicker && (
            <motion.div
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: "easeOut" }}
              className="mb-3 flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.32em] text-accent-cyan"
            >
              <span className="inline-flex h-[6px] w-[6px] rotate-45 bg-accent-cyan ring-cyan-glow" />
              {kicker}
            </motion.div>
          )}
          <motion.h1
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, ease: "easeOut", delay: 0.05 }}
            className="font-display text-[44px] font-semibold leading-[1.05] tracking-[-0.04em] text-ink sm:text-[64px] lg:text-[88px]"
          >
            {headline}
          </motion.h1>
          {subhead && (
            <motion.p
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: "easeOut", delay: 0.15 }}
              className="mt-3 max-w-2xl text-sm text-ink-muted sm:text-base"
            >
              {subhead}
            </motion.p>
          )}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: "easeOut", delay: 0.2 }}
          className="flex flex-col items-start gap-3 lg:items-end"
        >
          {glyph && (
            <div
              className="font-mono text-[64px] font-light leading-none text-ink-dim/40 sm:text-[88px]"
              aria-hidden
            >
              {glyph}
            </div>
          )}
          {rightSlot}
        </motion.div>
      </div>
    </header>
  );
}
