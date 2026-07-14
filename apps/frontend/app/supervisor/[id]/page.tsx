import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchSupervisorRun } from "../../../lib/api";
import { GlowingCard } from "../../../components/glowing-card";
import { SupervisorRunDetail } from "../../../components/supervisor-run-detail";
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
  return (
    <div className="mx-auto max-w-[1440px] px-6 py-6">
      <Link
        href="/"
        className="font-mono text-xs text-accent-cyan hover:underline"
      >
        ← 返回仪表板
      </Link>
      <header className="mt-3 mb-6">
        <h1 className="font-display text-2xl font-semibold tracking-tight text-ink">
          Supervisor 运行 #{run.id}
        </h1>
        <p className="mt-1 font-mono text-xs uppercase tracking-[0.18em] text-ink-muted">
          Hierarchical supervisor · detailed trace
        </p>
      </header>
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
