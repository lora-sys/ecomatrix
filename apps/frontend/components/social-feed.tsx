"use client";

import { useEffect, useState } from "react";
import { fetchAgents } from "../lib/api";
import clsx from "clsx";

type FeedPost = {
  ID: number;
  AgentID: number;
  Content: string;
  IntentType: "OFFER" | "REQUEST" | "SOCIAL" | "META";
  CreatedAt: string;
};

const INTENT_COLOR: Record<FeedPost["IntentType"], string> = {
  OFFER: "text-accent-gold",
  REQUEST: "text-accent-cyan",
  SOCIAL: "text-accent-violet",
  META: "text-ink-dim",
};

export function SocialFeed() {
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [names, setNames] = useState<Record<number, string>>({});

  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const [r, agents] = await Promise.all([
          fetch("/api/proxy/feeds?limit=50", { cache: "no-store" }).then((r) => r.ok ? r.json() : { feeds: [] }),
          names && Object.keys(names).length > 0 ? Promise.resolve(null) : fetchAgents(),
        ]);
        if (stopped) return;
        setPosts((r as { feeds: FeedPost[] }).feeds || []);
        if (agents) {
          const m: Record<number, string> = {};
          for (const a of agents) m[a.ID] = a.StringID;
          setNames(m);
        }
      } catch {
        /* ignore */
      }
    };
    tick();
    const id = setInterval(tick, 4000);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, []);  // eslint-disable-line react-hooks/exhaustive-deps

  if (posts.length === 0) {
    return (
      <div className="font-mono text-xs text-ink-dim">
        等待 Agent 们的第一条广播…
      </div>
    );
  }
  return (
    <ul className="divide-y divide-hairline/60 font-mono text-xs">
      {posts.slice(0, 12).map((p) => (
        <li key={p.ID} className="flex items-baseline justify-between gap-3 py-1.5">
          <span className="flex-1 truncate">
            <span className={clsx("mr-2", INTENT_COLOR[p.IntentType])}>
              [{p.IntentType}]
            </span>
            <span className="text-ink">{names[p.AgentID] ?? `#${p.AgentID}`}</span>
            <span className="mx-2 text-ink-dim">·</span>
            <span className="text-ink-muted">{p.Content}</span>
          </span>
          <time className="shrink-0 text-ink-dim">
            {new Date(p.CreatedAt).toLocaleTimeString("zh-CN", { hour12: false })}
          </time>
        </li>
      ))}
    </ul>
  );
}
