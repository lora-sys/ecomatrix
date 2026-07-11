"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Agent, MetricsSnapshot } from "../lib/types";
import { fetchMetrics } from "../lib/api";

interface Row {
  label: string;
  balance: number;
  job: string;
}

const JOB_COLOR: Record<string, string> = {
  merchant: "#7C5CFF", // violet
  miner: "#F5C044",    // gold
  hacker: "#F43F5E",   // rose
  mediator: "#22D3EE", // cyan
};

export function WealthChart({ agents }: { agents: Agent[] }) {
  const [snapshot, setSnapshot] = useState<MetricsSnapshot | null>(null);

  // Poll /v1/metrics to get the global total_gold for the chart footer.
  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const m = await fetchMetrics();
        if (!stopped) setSnapshot(m);
      } catch {
        /* ignore */
      }
    };
    tick();
    const id = setInterval(tick, 3000);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, []);

  // We don't have a time series yet; show the current sorted agent balances.
  // Future: replace with a ring buffer of snapshots for a true trend line.
  const rows = useMemo<Row[]>(() => {
    return [...agents]
      .sort((a, b) => b.Balance - a.Balance)
      .slice(0, 12)
      .map((a) => ({ label: a.StringID, balance: a.Balance, job: a.JobType }));
  }, [agents]);

  if (rows.length === 0) {
    return (
      <div className="font-mono text-xs text-ink-dim">无 Agent 数据</div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="h-44">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={rows} margin={{ top: 4, right: 8, bottom: 4, left: 0 }}>
            <defs>
              <linearGradient id="wealth-fill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#22D3EE" stopOpacity={0.55} />
                <stop offset="50%" stopColor="#F5C044" stopOpacity={0.35} />
                <stop offset="100%" stopColor="#F43F5E" stopOpacity={0.05} />
              </linearGradient>
            </defs>
            <CartesianGrid stroke="#1F2A44" strokeDasharray="3 3" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fill: "#8A95B2", fontSize: 10, fontFamily: "var(--font-mono)" }}
              tickLine={false}
              axisLine={{ stroke: "#1F2A44" }}
              interval={0}
              angle={-30}
              textAnchor="end"
              height={50}
            />
            <YAxis
              tick={{ fill: "#8A95B2", fontSize: 10, fontFamily: "var(--font-mono)" }}
              tickLine={false}
              axisLine={false}
              width={36}
            />
            <Tooltip
              contentStyle={{
                background: "#0E1422",
                border: "1px solid #1F2A44",
                fontFamily: "var(--font-mono)",
                fontSize: 11,
              }}
              labelStyle={{ color: "#E6EDF7" }}
              itemStyle={{ color: "#E6EDF7" }}
              cursor={{ stroke: "#7C5CFF", strokeWidth: 1, strokeDasharray: "3 3" }}
              formatter={(value: number) => [`${value} GOLD`, "BALANCE"]}
            />
            <Area
              type="monotone"
              dataKey="balance"
              stroke="#22D3EE"
              strokeWidth={2}
              fill="url(#wealth-fill)"
              dot={(props: { cx?: number; cy?: number; index?: number }) => {
                const { cx, cy, index } = props;
                if (cx == null || cy == null || index == null) return <g />;
                const color = JOB_COLOR[rows[index]?.job] ?? "#22D3EE";
                return <circle key={index} cx={cx} cy={cy} r={3.5} fill={color} />;
              }}
              activeDot={{ r: 5, stroke: "#E6EDF7", strokeWidth: 1 }}
            >
              {rows.map((row, i) => (
                <Cell key={i} fill={JOB_COLOR[row.job] ?? "#22D3EE"} />
              ))}
            </Area>
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="flex items-baseline justify-between border-t border-hairline pt-2 font-mono text-xs">
        <span className="text-ink-dim">TOP 12 GOLD</span>
        <span className="text-accent-gold">
          {snapshot ? snapshot.total_gold.toLocaleString() : "—"} GOLD
        </span>
      </div>
    </div>
  );
}
