"use client";

import Link from "next/link";
import { useStore } from "../hooks/store";
import { KPITile } from "../components/kpi-tile";
import { GlowingCard } from "../components/glowing-card";
import { WealthChart } from "../components/wealth-chart";
import { WealthHistory } from "../components/wealth-history";
import { TradeVolumeChart } from "../components/trade-volume-chart";
import { TradeFeed } from "../components/trade-feed";
import { SocialFeed } from "../components/social-feed";
import { SupervisorLog } from "../components/supervisor-log";
import { ThreeDCard } from "../components/three-d-card";
import { Agent, MetricsSnapshot } from "../lib/types";
import { ErrorBanner } from "../components/error-banner";
import { LiveProvider } from "../components/live-provider";
import { Masthead } from "../components/masthead";
import { BigMetric } from "../components/big-metric";
import { TickerRibbon } from "../components/ticker-ribbon";

import type { SupervisorRun } from "../lib/types";

export function DashboardClient({
  initialAgents,
  initialMetrics,
  initialSupervisorRuns = [],
}: {
  initialAgents: Agent[];
  initialMetrics: MetricsSnapshot | null;
  initialSupervisorRuns?: SupervisorRun[];
}) {
  const metrics = useStore((s) => s.metrics);
  const liveTotalGold = metrics.total_gold || initialMetrics?.total_gold || 0;
  const liveAgentCount = metrics.agent_count || initialMetrics?.agent_count || 0;
  const liveWs = metrics.ws_connections || initialMetrics?.ws_connections || 0;
  const liveQps = metrics.recent_qps || initialMetrics?.recent_qps || 0;

  return (
    <LiveProvider initialMetrics={initialMetrics}>
      <ErrorBanner />

      <div className="mx-auto max-w-[1440px] space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        {/* Layer 1 — masthead + ticker */}
        <div className="space-y-4">
          <Masthead
            kicker="ECOMATRIX · AGENT ECONOMY · OBSERVED LIVE"
            headline={
              <span className="grad-text-cyan-violet">{"上帝视角"}</span>
            }
            subhead={
              <>
                Every <span className="text-accent-cyan">trade</span>, every{" "}
                <span className="text-accent-violet">conversation</span>, every
                supervisor <span className="text-accent-gold">decision</span>{" "}
                — captured as it happens across the
                {" "}
                <span className="text-ink">multiverse of economic agents</span>.
              </>
            }
            glyph="◬"
            rightSlot={
              <div className="flex flex-col items-start gap-2 lg:items-end">
                <span className="font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
                  DASHBOARD · v0.3.0
                </span>
                <Link
                  href="/agents/agent_miner_01"
                  className="rounded-md border border-hairline/60 bg-panel/40 px-3 py-1.5 font-mono text-[11px] uppercase tracking-[0.22em] text-accent-cyan hover:ring-cyan-glow"
                >
                  ↗ 旁观首个 Agent
                </Link>
              </div>
            }
          />
          <TickerRibbon initialMetrics={initialMetrics} />
        </div>

        {/* Layer 2 — full-bleed big metrics, asymmetric placement. */}
        <section className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-12">
          <div className="lg:col-span-3">
            <BigMetric
              label="存活 Agent"
              value={liveAgentCount}
              tone="cyan"
              hint={`${initialAgents.length} 总数 · ${Object.keys(metrics.jobs_breakdown ?? {}).length} 工种`}
            />
          </div>
          <div className="lg:col-span-3">
            <BigMetric
              label="全网总资产"
              value={liveTotalGold}
              tone="gold"
              unit="GOLD"
              format={(n) => Math.round(n).toLocaleString("en-US")}
              hint={<span>escrowed across all agents</span>}
            />
          </div>
          <div className="lg:col-span-3">
            <BigMetric
              label="近 10S QPS"
              value={liveQps}
              tone="violet"
              format={(n) => n.toFixed(2)}
              delta={liveQps > 0 ? liveQps * 0.4 : 0}
              hint={<span>rolling transactions / sec</span>}
            />
          </div>
          <div className="lg:col-span-3">
            <BigMetric
              label="在线观测端"
              value={liveWs}
              tone="rose"
              hint={<span>spectators attached via WS</span>}
            />
          </div>
        </section>

        {/* Layer 3 — body: charts + feeds, intentionally asymmetric. */}
        <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="space-y-4 lg:col-span-5">
            <GlowingCard label="财富分布 · TOP 12" tone="gold">
              <WealthChart agents={initialAgents} />
            </GlowingCard>
            <GlowingCard label="全网 GOLD · 历史 2 分钟" tone="gold">
              <WealthHistory />
            </GlowingCard>
          </div>
          <div className="space-y-4 lg:col-span-3">
            <GlowingCard label="赛博交易广播" tone="cyan">
              <TradeFeed />
            </GlowingCard>
            <GlowingCard label="交易量 · 1 秒桶" tone="cyan">
              <TradeVolumeChart />
            </GlowingCard>
          </div>
          <div className="lg:col-span-4">
            <GlowingCard label="社交广场 · POST_FEED" tone="violet">
              <SocialFeed />
            </GlowingCard>
          </div>
          <div className="lg:col-span-3 lg:col-start-9">
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

        {/* Layer 4 — supervisor log panel. */}
        <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="lg:col-span-7">
            <GlowingCard label="Supervisor 任务日志" tone="violet">
              <SupervisorLog initialRuns={initialSupervisorRuns} />
            </GlowingCard>
          </div>
          <div className="lg:col-span-5">
            <GlowingCard label="工种分布 · JOB CARDS" tone="cyan">
              <div className="grid grid-cols-2 gap-3">
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
              </div>
            </GlowingCard>
          </div>
        </section>
      </div>
    </LiveProvider>
  );
}
