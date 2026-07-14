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
  return (
    <div
      aria-label="live trade broadcast"
      aria-live="polite"
      className="font-mono text-xs"
    >
      {feed.length === 0 ? (
        <div className="text-ink-dim">等待第一笔交易…</div>
      ) : (
        <ul
          className="divide-y divide-hairline/60"
          role="log"
          aria-relevant="additions"
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
      )}
    </div>
  );
}
