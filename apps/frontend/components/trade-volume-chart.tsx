"use client";

import { useEffect, useState } from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { fetchMetricsHistory } from "../lib/api";
import type { MetricsHistorySample } from "../lib/types";

// Trade volume per 1-second sample. The data is the last N samples from
// /v1/metrics/history, so the bar chart is a fine-grained time series.
export function TradeVolumeChart() {
  const [samples, setSamples] = useState<MetricsHistorySample[]>([]);

  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const h = await fetchMetricsHistory();
        if (!stopped) setSamples(h.samples);
      } catch {}
    };
    tick();
    const id = setInterval(tick, 3000);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, []);

  if (samples.length < 2) {
    return (
      <div className="font-mono text-xs text-ink-dim">交易采样中…</div>
    );
  }
  return (
    <div className="h-24">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={samples.map((s) => ({
            t: new Date(s.at).toLocaleTimeString("zh-CN", { hour12: false }),
            trades: s.trade_count,
          }))}
          margin={{ top: 4, right: 8, bottom: 4, left: 0 }}
        >
          <CartesianGrid stroke="#1F2A44" strokeDasharray="3 3" vertical={false} />
          <XAxis
            dataKey="t"
            tick={{ fill: "#5A6685", fontSize: 9, fontFamily: "var(--font-mono)" }}
            tickLine={false}
            axisLine={false}
            interval="preserveStartEnd"
            minTickGap={32}
          />
          <YAxis
            allowDecimals={false}
            tick={{ fill: "#5A6685", fontSize: 9, fontFamily: "var(--font-mono)" }}
            tickLine={false}
            axisLine={false}
            width={28}
          />
          <Tooltip
            contentStyle={{
              background: "#0E1422",
              border: "1px solid #1F2A44",
              fontFamily: "var(--font-mono)",
              fontSize: 11,
            }}
            labelStyle={{ color: "#E6EDF7" }}
            itemStyle={{ color: "#22D3EE" }}
            cursor={{ fill: "rgba(34,211,238,0.06)" }}
            formatter={(v: number) => [`${v}`, "trades"]}
          />
          <Bar dataKey="trades" fill="#22D3EE" isAnimationActive={false} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
