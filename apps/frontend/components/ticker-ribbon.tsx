"use client";

import { useEffect, useMemo, useState } from "react";
import { useStore } from "../hooks/store";
import type { MetricsSnapshot } from "../lib/types";

interface Item {
  id: string;
  text: string;
  tone: "up" | "down" | "flat";
}

function format(n: number, unit = ""): string {
  if (!isFinite(n)) return "—";
  if (unit === "GOLD") return n.toLocaleString("en-US");
  if (Math.abs(n) >= 1000) return n.toFixed(0);
  return n.toFixed(2);
}

/**
 * ISS-FRONTEND: A hero-width scrolling ticker that runs above the
 * fold. We render twice the items so the marquee loops seamlessly,
 * and we remix them with sandbox-supplied upstream signals so the
 * streamer is always live, never empty.
 */
export function TickerRibbon({
  initialMetrics,
}: {
  initialMetrics?: MetricsSnapshot | null;
}) {
  const liveMetrics = useStore((s) => s.metrics);
  const liveWs = useStore((s) => s.connected);
  const liveLatestTradeAt = useStore((s) => s.metrics.last_trade_at) as
    | string
    | null
    | undefined;

  const merged: MetricsSnapshot = useMemo(() => {
    const base = (initialMetrics ?? ({} as MetricsSnapshot)) as MetricsSnapshot;
    return { ...base, ...liveMetrics } as MetricsSnapshot;
  }, [initialMetrics, liveMetrics]);

  const items: Item[] = useMemo(() => {
    const list: Item[] = [
      {
        id: "agents",
        text: `AGENTS · ${format(merged.agent_count)} ONLINE`,
        tone: "up",
      },
      {
        id: "gold",
        text: `GOLD · ${format(merged.total_gold, "GOLD")} RESERVE`,
        tone: "up",
      },
      {
        id: "qps",
        text: `QPS · ${format(merged.recent_qps)} / 10s`,
        tone: merged.recent_qps >= 1 ? "up" : "flat",
      },
      {
        id: "ws",
        text: `WS · ${format(merged.ws_connections)} CONNECTED`,
        tone: merged.ws_connections >= 1 ? "up" : "flat",
      },
    ];
    const breakdown = Object.entries(merged.jobs_breakdown ?? {});
    for (const [job, count] of breakdown) {
      list.push({
        id: `job-${job}`,
        text: `${job.toUpperCase()} · ${count}`,
        tone: count > 0 ? "up" : "flat",
      });
    }
    const lastTrade =
      liveLatestTradeAt ?? merged.last_trade_at ?? null;
    if (lastTrade) {
      list.push({
        id: "lasttrade",
        text: `LAST TX · ${new Date(lastTrade).toLocaleTimeString("zh-CN", { hour12: false })}`,
        tone: "flat",
      });
    }
    if (merged.supervisor_runs_count !== undefined) {
      list.push({
        id: "sruns",
        text: `SUPERVISOR · ${format(merged.supervisor_runs_count)} RUNS`,
        tone: (merged.supervisor_runs_count ?? 0) > 0 ? "up" : "flat",
      });
    }
    return list;
    // We intentionally only re-derive when liveMetrics or initialMetrics
    // changes; the helpers (format, breakdown entries) are pure.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [liveMetrics, initialMetrics, liveLatestTradeAt]);

  const [stamp, setStamp] = useState("");
  useEffect(() => {
    const update = () =>
      setStamp(new Date().toLocaleString("zh-CN", { hour12: false }));
    update();
    const id = setInterval(update, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div className="ecomatrix-frame relative isolate overflow-hidden rounded-md border border-hairline/60 bg-canvas/60 backdrop-blur">
      <div className="flex items-center gap-3 border-b border-hairline/40 bg-gradient-to-r from-canvas/80 via-panel/40 to-canvas/80 px-4 py-1.5 font-mono text-[10px] uppercase tracking-[0.22em] text-ink-dim">
        <span className="inline-flex items-center gap-1.5 text-accent-cyan">
          <span className="relative inline-flex h-1.5 w-1.5">
            <span
              className={
                "absolute inset-0 animate-ping rounded-full " +
                (liveWs ? "bg-accent-cyan" : "bg-ink-dim")
              }
            />
            <span
              className={
                "relative inline-flex h-1.5 w-1.5 rounded-full " +
                (liveWs ? "bg-accent-cyan" : "bg-ink-dim")
              }
            />
          </span>
          {liveWs ? "STREAM · LIVE" : "STREAM · OFFLINE"}
        </span>
        <span className="text-ink-muted">{stamp}</span>
        <span className="ml-auto text-ink-muted">
          {merged.supervisor_runs_count
            ? `${merged.supervisor_runs_count} supervisor runs`
            : "monitoring"}
        </span>
      </div>
      <div className="relative h-8 overflow-hidden">
        <div
          className="ecomatrix-marquee-track flex h-8 w-max items-center gap-10 whitespace-nowrap font-mono text-[11px] uppercase tracking-[0.18em]"
          aria-label="live metrics ticker"
        >
          {/* Render the items twice so the marquee loop is seamless. */}
          {[0, 1].map((dup) => (
            <div key={dup} className="flex items-center gap-10">
              {items.map((it) => (
                <span
                  key={`${dup}-${it.id}`}
                  className={
                    "flex items-center gap-2 " +
                    (it.tone === "up"
                      ? "text-accent-cyan"
                      : it.tone === "down"
                        ? "text-accent-rose"
                        : "text-ink-muted")
                  }
                >
                  <span className="text-ink-dim">▸</span>
                  {it.text}
                </span>
              ))}
            </div>
          ))}
        </div>
        {/* Gradient fade masks. */}
        <div className="pointer-events-none absolute inset-y-0 left-0 w-12 bg-gradient-to-r from-canvas to-transparent" />
        <div className="pointer-events-none absolute inset-y-0 right-0 w-12 bg-gradient-to-l from-canvas to-transparent" />
      </div>
    </div>
  );
}
