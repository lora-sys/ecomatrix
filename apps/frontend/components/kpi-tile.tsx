"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { GlowingCard } from "./glowing-card";
import { clamp, dampValue } from "../lib/damping";

interface Props {
  label: string;
  value: number;
  format?: (n: number) => string;
  tone?: "cyan" | "gold" | "violet" | "rose";
  unit?: string;
}

export function KPITile({ label, value, format, tone = "cyan", unit }: Props) {
  const [display, setDisplay] = useState(value);
  const [prev, setPrev] = useState(value);
  const [lastTick, setLastTick] = useState(() => Date.now());

  useEffect(() => {
    if (value === prev) return;
    let raf = 0;
    const step = () => {
      const now = Date.now();
      const dt = clamp(now - lastTick, 0, 100);
      setDisplay((d) => dampValue(d, value, dt, 220));
      if (Math.abs(display - value) > 0.5) {
        setLastTick(now);
        raf = requestAnimationFrame(step);
      } else {
        setDisplay(value);
        setPrev(value);
      }
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [value]);  // eslint-disable-line react-hooks/exhaustive-deps

  const text = format ? format(display) : Math.round(display).toLocaleString();

  return (
    <GlowingCard label={label} tone={tone}>
      <motion.div
        className="flex items-baseline gap-2 font-display"
        aria-live="polite"
        aria-label={`${label} ${text}${unit ?? ""}`}
      >
        <span className="text-3xl font-semibold tracking-tight text-ink">
          {text}
        </span>
        {unit ? <span className="text-xs text-ink-dim font-mono">{unit}</span> : null}
      </motion.div>
    </GlowingCard>
  );
}
