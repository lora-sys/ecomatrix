import Link from "next/link";
import { fetchAgent, fetchTransactions, fetchAgents, fetchLongTermMemory, fetchAgentSupervisorRuns } from "../../../lib/api";
import { notFound } from "next/navigation";
import { GlowingCard } from "../../../components/glowing-card";
import { AIThoughtTrace } from "../../../components/ai-thought-trace";
import { TracingBeam } from "../../../components/tracing-beam";
import { LiveProvider } from "../../../components/live-provider";
import { AgentSupervisorHistory } from "../../../components/agent-supervisor-history";
import { Masthead } from "../../../components/masthead";
import { AmbientBackground } from "../../../components/ambient-bg";
import { BigMetric } from "../../../components/big-metric";

export const dynamic = "force-dynamic";

export default async function AgentDetail({
  params,
}: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let agent;
  try {
    agent = await fetchAgent(id);
  } catch {
    notFound();
  }
  const [txs, allAgents, ltm, supervisorRuns] = await Promise.all([
    fetchTransactions(100),
    fetchAgents(),
    fetchLongTermMemory(id).catch(() => ({ summary: "", facts: [] })),
    fetchAgentSupervisorRuns(id, 6).catch(() => []),
  ]);
  const idToString = new Map(allAgents.map((a: any) => [a.ID, a.StringID]));
  const supervisorHistory = (supervisorRuns ?? []).map((r) => ({
    id: r.id,
    goal: r.goal,
    status: r.status,
    error: r.error,
    warnings: r.warnings,
    subtasks: r.subtasks,
    worker_results: r.worker_results,
    final_summary: r.final_summary,
    tokens_used: r.tokens_used,
    tokens_budget: r.tokens_budget,
    started_at: r.started_at,
    finished_at: r.finished_at,
    duration_ms: r.duration_ms,
  }));
  const mine = txs.filter((t: any) => t.FromID === agent.ID || t.ToID === agent.ID).slice(0, 20);

  return (
    <LiveProvider>
      <div className="mx-auto max-w-[1440px] space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        <Masthead
          kicker={`DOSSIER · AGENT · ${agent.JobType.toUpperCase()}`}
          headline={
            <span className="grad-text-rainbow">{agent.StringID}</span>
          }
          subhead={
            <>
              Long-term memory · AI trace · recent trades · supervisor
              participation. Everything we know about this agent, captured
              in one dossier.
            </>
          }
          glyph="✦"
          rightSlot={
            <Link
              href="/"
              className="rounded-md border border-hairline/60 bg-panel/40 px-3 py-1.5 font-mono text-[11px] uppercase tracking-[0.22em] text-accent-cyan hover:ring-cyan-glow"
            >
              ← 返回大盘
            </Link>
          }
        />

        <section className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-12">
          <div className="lg:col-span-4">
            <BigMetric label="余额 · BALANCE" value={agent.Balance} tone="gold" unit="GOLD" />
          </div>
          <div className="lg:col-span-4">
            <BigMetric label="活力 · VITALITY" value={agent.Vitality} tone="cyan" />
          </div>
          <div className="lg:col-span-4">
            <BigMetric label="信用 · CREDIT" value={agent.CreditScore} tone="violet" />
          </div>
        </section>

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="lg:col-span-5">
            <GlowingCard label="AI 思考链路 · LLM Trace" tone="cyan">
              <AIThoughtTrace agentId={id} />
            </GlowingCard>
            <div className="mt-4">
              <GlowingCard label="长期记忆 · LTM" tone="violet">
                {ltm.summary ? (
                  <p className="mb-2 text-sm text-ink">{ltm.summary}</p>
                ) : (
                  <p className="mb-2 font-mono text-xs text-ink-dim">
                    该 Agent 暂无长期记忆
                  </p>
                )}
                {ltm.facts.length > 0 ? (
                  <ul className="space-y-1 font-mono text-xs text-ink-muted">
                    {ltm.facts.slice(-8).reverse().map((f: any, i: number) => (
                      <li key={i} className="truncate">· {f}</li>
                    ))}
                  </ul>
                ) : null}
              </GlowingCard>
            </div>
          </div>
          <div className="lg:col-span-7">
            <GlowingCard label="近期交易 · 流水" tone="gold">
              {mine.length === 0 ? (
                <div className="font-mono text-xs text-ink-dim">
                  该 Agent 暂无近期交易
                </div>
              ) : (
                <TracingBeam
                  items={mine.map((t: any) => ({
                    id: String(t.ID),
                    tone: t.Status === "SETTLED" ? "cyan" : "rose",
                    node: (
                      <span className="font-mono text-xs">
                        {idToString.get(t.FromID) ?? t.FromID}
                        {" → "}
                        {idToString.get(t.ToID) ?? t.ToID}
                        {"  "}
                        <span className="text-accent-gold">{t.Amount} GOLD</span>
                        {"  "}
                        <span className={t.Status === "SETTLED" ? "text-accent-emerald" : "text-accent-rose"}>
                          {t.Status}
                        </span>
                      </span>
                    ),
                  }))}
                />
              )}
            </GlowingCard>
          </div>
        </section>

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="lg:col-span-12">
            <GlowingCard
              label="近期 Supervisor 运行 · 调度历史"
              tone="violet"
            >
              <AgentSupervisorHistory
                agentId={agent.StringID}
                initialRuns={supervisorHistory}
              />
            </GlowingCard>
          </div>
        </section>
      </div>
    </LiveProvider>
  );
}
