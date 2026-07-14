import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchSupervisorRun } from "../../../lib/api";
import { GlowingCard } from "../../../components/glowing-card";
import { SupervisorRunDetail } from "../../../components/supervisor-run-detail";
import { Masthead } from "../../../components/masthead";
import type { SupervisorRun } from "../../../lib/types";

export const dynamic = "force-dynamic";

export default async function SupervisorRunPage({
  params,
}: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const numId = Number(id);
  if (!Number.isFinite(numId) || numId <= 0) {
    notFound();
  }
  let run: SupervisorRun;
  try {
    const payload = await fetchSupervisorRun(numId);
    run = {
      id: payload.id,
      goal: payload.goal,
      status: payload.status,
      error: payload.error,
      warnings: payload.warnings,
      subtasks: payload.subtasks,
      worker_results: payload.worker_results,
      final_summary: payload.final_summary,
      tokens_used: payload.tokens_used,
      tokens_budget: payload.tokens_budget,
      started_at: payload.started_at,
      finished_at: payload.finished_at,
      duration_ms: payload.duration_ms,
    };
  } catch {
    notFound();
  }

  const tokensPct =
    run.tokens_budget > 0
      ? Math.min(100, Math.round((run.tokens_used / run.tokens_budget) * 100))
      : 0;

  return (
    <div className="mx-auto max-w-[1440px] space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <Masthead
        kicker="TRANSMISSAL · SUPERVISOR RUN"
        headline={
          <span className="grad-text-cyan-violet">RUN #{run.id}</span>
        }
        subhead={<>{run.goal}</>}
        glyph="⚡"
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
        <div className="lg:col-span-3">
          <div className="ecomatrix-frame relative rounded-md border border-hairline/40 bg-gradient-to-br from-panel/60 to-canvas/40 p-4">
            <div className="font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
              SUBTASKS
            </div>
            <div className="ecomatrix-numeral mt-2 font-display text-4xl font-medium text-ink">
              {run.subtasks?.length ?? 0}
            </div>
          </div>
        </div>
        <div className="lg:col-span-3">
          <div className="ecomatrix-frame relative rounded-md border border-hairline/40 bg-gradient-to-br from-panel/60 to-canvas/40 p-4">
            <div className="font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
              WORKERS
            </div>
            <div className="ecomatrix-numeral mt-2 font-display text-4xl font-medium text-ink">
              {run.worker_results?.length ?? 0}
            </div>
          </div>
        </div>
        <div className="lg:col-span-3">
          <div className="ecomatrix-frame relative rounded-md border border-hairline/40 bg-gradient-to-br from-panel/60 to-canvas/40 p-4">
            <div className="font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
              TOKENS · USED / BUDGET
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="ecomatrix-numeral font-display text-4xl font-medium text-ink">
                {run.tokens_used}
              </span>
              <span className="font-mono text-xs text-ink-muted">
                / {run.tokens_budget}
              </span>
            </div>
            <div className="mt-2 h-1 w-full overflow-hidden rounded-full bg-hairline/40">
              <div
                className="h-full bg-gradient-to-r from-accent-cyan via-accent-violet to-accent-gold"
                style={{ width: `${tokensPct}%` }}
              />
            </div>
            <div className="mt-1 font-mono text-[10px] uppercase tracking-[0.18em] text-ink-dim">
              {tokensPct}%
            </div>
          </div>
        </div>
        <div className="lg:col-span-3">
          <div className="ecomatrix-frame relative rounded-md border border-hairline/40 bg-gradient-to-br from-panel/60 to-canvas/40 p-4">
            <div className="font-mono text-[10px] uppercase tracking-[0.28em] text-ink-dim">
              DURATION
            </div>
            <div className="mt-2 font-display text-4xl font-medium text-ink">
              {run.duration_ms ?? 0}
              <span className="ml-1 font-mono text-xs text-ink-muted">ms</span>
            </div>
          </div>
        </div>
      </section>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
        <div className="lg:col-span-12">
          <GlowingCard label="Supervisor Run · 详细信息" tone="violet">
            <SupervisorRunDetail initialRun={run} />
          </GlowingCard>
        </div>
      </section>
    </div>
  );
}
