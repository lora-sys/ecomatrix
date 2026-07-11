"use client";

import Link from "next/link";
import { useStore } from "../hooks/store";
import { KPITile } from "../components/kpi-tile";
import { GlowingCard } from "../components/glowing-card";
import { WealthChart } from "../components/wealth-chart";
import { TradeFeed } from "../components/trade-feed";
import { ThreeDCard } from "../components/three-d-card";
import { Agent } from "../lib/types";
import { LiveProvider } from "../components/live-provider";

export function DashboardClient({ initialAgents }: { initialAgents: Agent[] }) {
  const metrics = useStore((s) => s.metrics);
  return (
    <LiveProvider>
      <div className="mx-auto max-w-[1440px] px-6 py-6">
        <header className="mb-6 flex items-end justify-between">
          <div>
            <h1 className="font-display text-2xl font-semibold tracking-tight text-ink">
              上帝视角
            </h1>
            <p className="mt-1 text-sm text-ink-muted">
              EcoMatrix 经济演化网络的实时观测
            </p>
          </div>
          <Link
            href="/agents/agent_miner_01"
            className="font-mono text-xs text-accent-cyan hover:underline"
          >
            查看 Agent →
          </Link>
        </header>

        {/* KPI row */}
        <section className="grid grid-cols-2 gap-3 md:grid-cols-4 mb-6">
          <KPITile label="存活 Agent" value={metrics.agent_count} tone="cyan" />
          <KPITile label="全网总资产" value={metrics.total_gold} tone="gold" unit="GOLD" />
          <KPITile
            label="近 10s QPS"
            value={metrics.recent_qps}
            tone="violet"
            format={(n) => n.toFixed(2)}
          />
          <KPITile label="在线观测端" value={metrics.ws_connections} tone="rose" />
        </section>

        {/* Body: wealth + feed + agents */}
        <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="lg:col-span-5">
            <GlowingCard label="财富分布 · TOP 12" tone="gold">
              <WealthChart agents={initialAgents} />
            </GlowingCard>
          </div>
          <div className="lg:col-span-4">
            <GlowingCard label="赛博交易广播" tone="cyan">
              <TradeFeed />
            </GlowingCard>
          </div>
          <div className="lg:col-span-3">
            <GlowingCard label="公民一览" tone="violet">
              <ul className="divide-y divide-hairline/60 text-sm">
                {initialAgents.slice(0, 8).map((a) => (
                  <li key={a.ID}>
                    <Link
                      href={`/agents/${a.StringID}`}
                      className="flex items-baseline justify-between py-1.5 hover:text-accent-cyan"
                    >
                      <span className="font-mono text-xs">{a.StringID}</span>
                      <span className="font-mono text-xs text-ink-muted">
                        {a.Balance}
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            </GlowingCard>
          </div>
        </section>

        {/* Optional: showcase 3D card with a job-color ring */}
        <section className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {(["miner", "merchant", "hacker", "mediator"] as const).map((j) => {
            const a = initialAgents.find((x) => x.JobType === j);
            if (!a) return null;
            return (
              <Link key={j} href={`/agents/${a.StringID}`}>
                <ThreeDCard
                  tone={j === "merchant" ? "gold" : "cyan"}
                  className="hover:cursor-pointer"
                >
                  <div className="font-mono text-[10px] uppercase tracking-[0.18em] text-ink-muted">
                    {j}
                  </div>
                  <div className="mt-1 font-display text-lg text-ink">{a.StringID}</div>
                  <div className="mt-2 grid grid-cols-3 gap-2 font-mono text-xs text-ink-muted">
                    <span>BAL {a.Balance}</span>
                    <span>VIT {a.Vitality}</span>
                    <span>CR  {a.CreditScore}</span>
                  </div>
                </ThreeDCard>
              </Link>
            );
          })}
        </section>
      </div>
    </LiveProvider>
  );
}
