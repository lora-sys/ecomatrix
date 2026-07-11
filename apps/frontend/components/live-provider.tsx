"use client";

import { ReactNode, useEffect } from "react";
import { useEconomatrixStream } from "../hooks/use-economatrix-stream";
import { useStore } from "../hooks/store";
import { fetchMetrics } from "../lib/api";

// Bridges the WS stream + periodic metrics polls into the store.
export function LiveProvider({ children }: { children: ReactNode }) {
  useEconomatrixStream();
  const setMetrics = useStore((s) => s.setMetrics);
  const connected = useStore((s) => s.connected);

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
      } catch {
        /* ignore; tile will hold last value */
      }
    };
    tick();
    const id = setInterval(tick, connected ? 3000 : 1500);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, [connected, setMetrics]);

  return <>{children}</>;
}
