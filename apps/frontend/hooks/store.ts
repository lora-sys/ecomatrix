"use client";

import { create } from "zustand";
import { StreamEvent, SupervisorRun } from "../lib/types";

type TradeFeedKind = "settled" | "rejected" | "replay" | "other";
type SocialIntent = "OFFER" | "REQUEST" | "SOCIAL" | "META";

interface TradeFeedItem {
  id: string;
  ts: number;
  text: string;
  kind: TradeFeedKind;
}

interface SocialFeedItem {
  post_id: number;
  agent_id: string;
  content: string;
  intent_type: SocialIntent;
  ts: number;
}

interface StoreState {
  connected: boolean;
  fetchError: string | null;
  metrics: {
    agent_count: number;
    total_gold: number;
    jobs_breakdown: Record<string, number>;
    recent_qps: number;
    ws_connections: number;
    last_trade_at: string | null;
  };
  feed: TradeFeedItem[];
  social: SocialFeedItem[];
  supervisorRuns: SupervisorRun[];
  supervisorLatest: SupervisorRun | null;
  supervisorHydrated: boolean;
  setConnected: (v: boolean) => void;
  setFetchError: (msg: string | null) => void;
  setMetrics: (m: Partial<StoreState["metrics"]>) => void;
  setSocial: (items: SocialFeedItem[]) => void;
  pushSocial: (item: SocialFeedItem) => void;
  setSupervisorRuns: (runs: SupervisorRun[]) => void;
  setSupervisorLatest: (run: SupervisorRun | null) => void;
  setSupervisorHydrated: (v: boolean) => void;
  applyEvent: (ev: StreamEvent) => void;
}

export const useStore = create<StoreState>((set) => ({
  connected: false,
  fetchError: null,
  metrics: {
    agent_count: 0,
    total_gold: 0,
    jobs_breakdown: {},
    recent_qps: 0,
    ws_connections: 0,
    last_trade_at: null,
  },
  feed: [],
  social: [],
  supervisorRuns: [],
  supervisorLatest: null,
  supervisorHydrated: false,
  setConnected: (v) => set({ connected: v }),
  setFetchError: (msg) => set({ fetchError: msg }),
  setMetrics: (m) =>
    set((s) => ({ metrics: { ...s.metrics, ...m } })),
  setSocial: (items) =>
    set((s) => ({
      // Keep the most recent 50; if WS events arrived during the fetch, they're
      // already at the head of `s.social`, so dedupe by post_id and merge.
      social: mergeSocial(s.social, items),
    })),
  pushSocial: (item) =>
    set((s) => ({ social: dedupeSocial([item, ...s.social]).slice(0, 50) })),
  setSupervisorRuns: (runs) =>
    set((s) => {
      const byId = new Map<number, SupervisorRun>();
      for (const r of s.supervisorRuns) if (r.id) byId.set(r.id, r);
      for (const r of runs) if (r.id) byId.set(r.id, r);
      const merged = Array.from(byId.values())
        .sort((a, b) => (b.id || 0) - (a.id || 0))
        .slice(0, 20);
      return { supervisorRuns: merged, supervisorHydrated: true };
    }),
  setSupervisorLatest: (run) =>
    set({ supervisorLatest: run, supervisorHydrated: true }),
  setSupervisorHydrated: (v) => set({ supervisorHydrated: v }),
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
      if (ev.type === "feed.posted") {
        const item: SocialFeedItem = {
          post_id: Number(ev.post_id ?? 0),
          agent_id: String(ev.agent_id ?? ""),
          content: String(ev.content ?? ""),
          intent_type: (ev.intent_type as SocialIntent) ?? "SOCIAL",
          ts: Date.now(),
        };
        return {
          social: dedupeSocial([item, ...s.social]).slice(0, 50),
        };
      }
      if (ev.type === "supervisor.run.finished" || ev.type === "supervisor.run.started") {
        const run = (ev as { run: SupervisorRun }).run;
        if (!run) return {};
        return {
          supervisorLatest: run,
          supervisorRuns: [
            run,
            ...s.supervisorRuns.filter((r) => r.id !== run.id),
          ].slice(0, 20),
          supervisorHydrated: true,
        };
      }
      return {};
    }),
}));

function dedupeSocial(items: SocialFeedItem[]): SocialFeedItem[] {
  const seen = new Set<number>();
  const out: SocialFeedItem[] = [];
  for (const it of items) {
    if (seen.has(it.post_id)) continue;
    seen.add(it.post_id);
    out.push(it);
  }
  return out;
}

function mergeSocial(
  live: SocialFeedItem[],
  fetched: SocialFeedItem[],
): SocialFeedItem[] {
  // Live items (from WS) take precedence over fetched ones with the same id.
  const liveIds = new Set(live.map((x) => x.post_id));
  const filteredFetched = fetched.filter((x) => !liveIds.has(x.post_id));
  return dedupeSocial([...live, ...filteredFetched]).slice(0, 50);
}
