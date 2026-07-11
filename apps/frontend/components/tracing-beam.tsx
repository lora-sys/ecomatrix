"use client";

import { ReactNode } from "react";
import clsx from "clsx";

interface Props {
  items: { id: string; node: ReactNode; tone?: "cyan" | "gold" | "violet" | "rose" }[];
  className?: string;
}

const DOT: Record<NonNullable<Props["items"][number]["tone"]>, string> = {
  cyan: "bg-accent-cyan",
  gold: "bg-accent-gold",
  violet: "bg-accent-violet",
  rose: "bg-accent-rose",
};

export function TracingBeam({ items, className }: Props) {
  return (
    <div className={clsx("relative", className)}>
      {/* The beam itself. */}
      <div className="absolute left-[7px] top-2 bottom-2 w-px bg-gradient-to-b from-accent-cyan/40 via-accent-violet/30 to-transparent" />
      <ul className="space-y-2">
        {items.map((it) => (
          <li key={it.id} className="relative pl-6">
            <span
              className={clsx(
                "absolute left-1 top-2 inline-block h-3 w-3 rounded-full",
                DOT[it.tone ?? "cyan"],
              )}
              style={{ boxShadow: "0 0 12px currentColor" }}
            />
            <div className="text-sm leading-relaxed text-ink">{it.node}</div>
          </li>
        ))}
      </ul>
    </div>
  );
}
