"use client";

import { useEffect, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { fetchMetricsHistory } from "../lib/api";
import type { MetricsHistorySample } from "../lib/types";

export function WealthHistory() {
  const [samples, setSamples] = useState<MetricsHistorySample[]>([]);

  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const h = await fetchMetricsHistory();
        if (!stopped) setSamples(h.samples);
      } catch {
        /* ignore; banner will show */
      }
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
      <div className="font-mono text-xs text-ink-dim">
        历史快照采集中…
      </div>
    );
  }

  return (
    <div className="h-32">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart
          data={samples.map((s) => ({
            t: new Date(s.at).toLocaleTimeString("zh-CN", {
              hour12: false,
            }),
            gold: s.total_gold,
            qps: s.recent_qps,
          }))}
          margin={{ top: 4, right: 8, bottom: 4, left: 0 }}
        >
          <defs>
            <linearGradient id="history-fill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#F5C044" stopOpacity={0.45} />
              <stop offset="100%" stopColor="#F5C044" stopOpacity={0.05} />
            </linearGradient>
          </defs>
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
            domain={["auto", "auto"]}
            tick={{ fill: "#5A6685", fontSize: 9, fontFamily: "var(--font-mono)" }}
            tickLine={false}
            axisLine={false}
            width={48}
          />
          <Tooltip
            contentStyle={{
              background: "#0E1422",
              border: "1px solid #1F2A44",
              fontFamily: "var(--font-mono)",
              fontSize: 11,
            }}
            labelStyle={{ color: "#E6EDF7" }}
            itemStyle={{ color: "#F5C044" }}
            cursor={{ stroke: "#7C5CFF", strokeWidth: 1, strokeDasharray: "3 3" }}
            formatter={(v: number) => [`${v.toLocaleString()} GOLD`, "TOTAL"]}
          />
          <Area
            type="monotone"
            dataKey="gold"
            stroke="#F5C044"
            strokeWidth={1.5}
            fill="url(#history-fill)"
            isAnimationActive={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
