"use client";

import { create } from "zustand";
import { StreamEvent } from "../lib/types";

interface FeedItem {
  id: string;
  ts: number;
  text: string;
  kind: "settled" | "rejected" | "replay" | "other";
}

interface StoreState {
  connected: boolean;
  metrics: {
    agent_count: number;
    total_gold: number;
    jobs_breakdown: Record<string, number>;
    recent_qps: number;
    ws_connections: number;
    last_trade_at: string | null;
  };
  feed: FeedItem[];
  // targets updated on each WS event or poll; rendering layer damps toward these.
  setConnected: (v: boolean) => void;
  setMetrics: (m: Partial<StoreState["metrics"]>) => void;
  pushFeed: (item: FeedItem) => void;
  applyEvent: (ev: StreamEvent) => void;
  capFeed: (n: number) => void;
}

export const useStore = create<StoreState>((set) => ({
  connected: false,
  metrics: {
    agent_count: 0,
    total_gold: 0,
    jobs_breakdown: {},
    recent_qps: 0,
    ws_connections: 0,
    last_trade_at: null,
  },
  feed: [],
  setConnected: (v) => set({ connected: v }),
  setMetrics: (m) =>
    set((s) => ({ metrics: { ...s.metrics, ...m } })),
  pushFeed: (item) =>
    set((s) => ({ feed: [item, ...s.feed].slice(0, 50) })),
  applyEvent: (ev) =>
    set((s) => {
      const id = `${ev.type}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
      if (ev.type === "trade.settled") {
        return {
          feed: [
            { id, ts: Date.now(), kind: "settled" as const,
              text: `${ev.from} → ${ev.to}  ${ev.amount} GOLD` },
            ...s.feed,
          ].slice(0, 50),
        };
      }
      if (ev.type === "trade.rejected") {
        return {
          feed: [
            { id, ts: Date.now(), kind: "rejected" as const,
              text: `${ev.msg_id}  ${ev.code}` },
            ...s.feed,
          ].slice(0, 50),
        };
      }
      if (ev.type === "trade.idempotent_replay") {
        return {
          feed: [
            { id, ts: Date.now(), kind: "replay" as const,
              text: `replay ${ev.tx_id}` },
            ...s.feed,
          ].slice(0, 50),
        };
      }
      return {};
    }),
  capFeed: (n) => set((s) => ({ feed: s.feed.slice(0, n) })),
}));
