import Link from "next/link";
import { fetchAgent, fetchTransactions, fetchAgents, fetchLongTermMemory } from "../../../lib/api";
import { notFound } from "next/navigation";
import { GlowingCard } from "../../../components/glowing-card";
import { AIThoughtTrace } from "../../../components/ai-thought-trace";
import { TracingBeam } from "../../../components/tracing-beam";
import { LiveProvider } from "../../../components/live-provider";

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
  const [txs, allAgents, ltm] = await Promise.all([
    fetchTransactions(100),
    fetchAgents(),
    fetchLongTermMemory(id).catch(() => ({ summary: "", facts: [] })),
  ]);
  const idToString = new Map(allAgents.map((a: any) => [a.ID, a.StringID]));
  const mine = txs.filter((t: any) => t.FromID === agent.ID || t.ToID === agent.ID).slice(0, 20);

  return (
    <LiveProvider>
      <div className="mx-auto max-w-[1440px] px-6 py-6">
        <Link href="/" className="font-mono text-xs text-accent-cyan hover:underline">
          ← 返回大盘
        </Link>
        <header className="mt-3 mb-6">
          <h1 className="font-display text-2xl font-semibold tracking-tight text-ink">
            {agent.StringID}
          </h1>
          <p className="mt-1 font-mono text-xs uppercase tracking-[0.18em] text-ink-muted">
            {agent.JobType}
          </p>
        </header>

        <section className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <GlowingCard label="基础状态" tone="cyan">
            <dl className="grid grid-cols-3 gap-3 font-mono text-sm">
              <div>
                <dt className="text-[10px] text-ink-dim">BALANCE</dt>
                <dd className="mt-1 text-lg text-accent-gold">{agent.Balance}</dd>
              </div>
              <div>
                <dt className="text-[10px] text-ink-dim">VITALITY</dt>
                <dd className="mt-1 text-lg text-accent-cyan">{agent.Vitality}</dd>
              </div>
              <div>
                <dt className="text-[10px] text-ink-dim">CREDIT</dt>
                <dd className="mt-1 text-lg text-accent-violet">{agent.CreditScore}</dd>
              </div>
            </dl>
          </GlowingCard>
          <div className="lg:col-span-2">
            <GlowingCard label="AI 思考链路 · LLM Trace" tone="cyan">
              <AIThoughtTrace agentId={id} />
            </GlowingCard>
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
          <div className="lg:col-span-2 lg:col-start-2 lg:row-start-2">
            <GlowingCard label="近期交易" tone="gold">
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
      </div>
    </LiveProvider>
  );
}
