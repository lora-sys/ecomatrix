"use client";

import { useEffect } from "react";
import clsx from "clsx";
import { useStore } from "../hooks/store";

const INTENT_COLOR: Record<string, string> = {
  OFFER: "text-accent-gold",
  REQUEST: "text-accent-cyan",
  SOCIAL: "text-accent-violet",
  META: "text-ink-dim",
};

export function SocialFeed() {
  const social = useStore((s) => s.social);
  const setSocial = useStore((s) => s.setSocial);

  // One-time hydrate from /api/proxy/feeds on mount; thereafter WS keeps it fresh.
  useEffect(() => {
    let stopped = false;
    (async () => {
      try {
        const r = await fetch("/api/proxy/feeds?limit=50", { cache: "no-store" });
        if (!r.ok) return;
        const body = await r.json();
        if (stopped) return;
        const items = (body.feeds as Array<{
          ID: number; AgentID: number; Content: string;
          IntentType: "OFFER" | "REQUEST" | "SOCIAL" | "META";
          CreatedAt: string;
        }>).map((p) => ({
          post_id: p.ID,
          agent_id: String(p.AgentID),
          content: p.Content,
          intent_type: p.IntentType,
          ts: new Date(p.CreatedAt).getTime(),
        }));
        setSocial(items);
      } catch {
        /* ignore */
      }
    })();
    return () => {
      stopped = true;
    };
  }, [setSocial]);

  return (
    <div
      aria-label="agent social feed"
      aria-live="polite"
      className="font-mono text-xs"
    >
      {social.length === 0 ? (
        <div className="text-ink-dim">等待 Agent 们的第一条广播…</div>
      ) : (
        <ul
          className="divide-y divide-hairline/60"
          role="log"
          aria-relevant="additions"
        >
          {social.slice(0, 12).map((p) => (
            <li key={p.post_id} className="flex items-baseline justify-between gap-3 py-1.5">
              <span className="flex-1 truncate">
                <span className={clsx("mr-2", INTENT_COLOR[p.intent_type])}>
                  [{p.intent_type}]
                </span>
                <span className="text-ink">{p.agent_id}</span>
                <span className="mx-2 text-ink-dim">·</span>
                <span className="text-ink-muted">{p.content}</span>
              </span>
              <time className="shrink-0 text-ink-dim">
                {new Date(p.ts).toLocaleTimeString("zh-CN", { hour12: false })}
              </time>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
