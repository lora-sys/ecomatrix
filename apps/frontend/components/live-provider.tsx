"use client";

import { ReactNode, useEffect } from "react";
import { useEconomatrixStream } from "../hooks/use-economatrix-stream";
import { useStore } from "../hooks/store";
import { fetchMetrics } from "../lib/api";
import type { MetricsSnapshot } from "../lib/types";

// Bridges the WS stream + periodic metrics polls into the store.
// Accepts an optional SSR-fetched initialMetrics so the first paint shows
// real numbers instead of zeros.
export function LiveProvider({ children, initialMetrics }: {
  children: ReactNode;
  initialMetrics?: MetricsSnapshot | null;
}) {
  useEconomatrixStream();
  const setMetrics = useStore((s) => s.setMetrics);
  const setFetchError = useStore((s) => s.setFetchError);
  const connected = useStore((s) => s.connected);
  const metrics = useStore((s) => s.metrics);

  // Seed the store from the SSR snapshot on first mount.
  useEffect(() => {
    if (initialMetrics && metrics.agent_count === 0) {
      setMetrics({
        agent_count: initialMetrics.agent_count,
        total_gold: initialMetrics.total_gold,
        jobs_breakdown: initialMetrics.jobs_breakdown,
        recent_qps: initialMetrics.recent_qps,
        ws_connections: initialMetrics.ws_connections,
        last_trade_at: initialMetrics.last_trade_at ?? null,
      });
    }
    if (initialMetrics) {
      setFetchError(null); // clear any stale error
    }
    // run only once
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      if (stopped) return;
      try {
        const m = await fetchMetrics();
        setMetrics({
          agent_count: m.agent_count,
          total_gold: m.total_gold,
          jobs_breakdown: m.jobs_breakdown,
          recent_qps: m.recent_qps,
          ws_connections: m.ws_connections,
          last_trade_at: m.last_trade_at ?? null,
        });
        setFetchError(null);
      } catch (e) {
        setFetchError(String(e));
      }
    };
    tick();
    const id = setInterval(tick, connected ? 3000 : 1500);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, [connected, setMetrics, setFetchError]);

  return <>{children}</>;
}
