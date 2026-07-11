"use client";

import { useMemo } from "react";
import { Agent } from "../lib/types";

interface Props {
  agents: Agent[];
}

// Simple horizontal bar visualization, sorted by balance.
// Uses the grad-wealth color scale.
export function WealthChart({ agents }: Props) {
  const sorted = useMemo(
    () => [...agents].sort((a, b) => b.Balance - a.Balance).slice(0, 12),
    [agents],
  );
  const max = Math.max(1, ...sorted.map((a) => a.Balance));
  return (
    <div className="space-y-1.5">
      {sorted.map((a) => {
        const w = (a.Balance / max) * 100;
        const tone = a.JobType === "merchant" ? "violet"
                   : a.JobType === "hacker"   ? "rose"
                   : a.JobType === "mediator" ? "cyan"
                   : "gold";
        return (
          <div key={a.ID} className="flex items-center gap-3">
            <div className="w-32 truncate font-mono text-xs text-ink-muted">
              {a.StringID}
            </div>
            <div className="relative h-3 flex-1 overflow-hidden rounded-sm bg-panel-2">
              <div
                className={`absolute inset-y-0 left-0 bg-accent-${tone}`}
                style={{ width: `${w}%`, opacity: 0.85 }}
              />
            </div>
            <div className="w-16 text-right font-mono text-xs text-ink">
              {a.Balance}
            </div>
          </div>
        );
      })}
    </div>
  );
}
