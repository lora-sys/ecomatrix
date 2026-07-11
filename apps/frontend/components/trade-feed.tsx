"use client";

import { useStore } from "../hooks/store";
import clsx from "clsx";

const KIND_COLOR: Record<string, string> = {
  settled: "text-accent-emerald",
  rejected: "text-accent-rose",
  replay: "text-accent-violet",
  other: "text-ink-muted",
};

export function TradeFeed() {
  const feed = useStore((s) => s.feed);
  if (feed.length === 0) {
    return (
      <div
        className="font-mono text-xs text-ink-dim"
        role="status"
        aria-live="polite"
      >
        等待第一笔交易…
      </div>
    );
  }
  return (
    <ul
      className="divide-y divide-hairline/60 font-mono text-xs"
      role="log"
      aria-live="polite"
      aria-relevant="additions"
      aria-label="live trade broadcast"
    >
      {feed.map((it) => (
        <li key={it.id} className="flex items-baseline justify-between py-1.5">
          <span className={clsx("truncate", KIND_COLOR[it.kind])}>{it.text}</span>
          <time className="ml-2 shrink-0 text-ink-dim">
            {new Date(it.ts).toLocaleTimeString("zh-CN", { hour12: false })}
          </time>
        </li>
      ))}
    </ul>
  );
}
