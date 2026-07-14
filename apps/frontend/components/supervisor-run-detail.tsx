"use client";

import Link from "next/link";
import { useMemo } from "react";
import { SupervisorRun } from "../lib/types";

interface Props {
  initialRun: SupervisorRun;
}

export function SupervisorRunDetail({ initialRun }: Props) {
  const workerCount = initialRun.worker_results?.length ?? 0;
  const subtaskCount = initialRun.subtasks?.length ?? 0;
  const tokensUsed = initialRun.tokens_used ?? 0;
  const tokensBudget = initialRun.tokens_budget ?? 0;
  const warnings = initialRun.warnings ?? [];
  const failed = initialRun.status === "failed" || Boolean(initialRun.error);
  const pct = tokensBudget > 0
    ? Math.min(100, Math.round((tokensUsed / tokensBudget) * 100))
    : 0;

  const started = useMemo(() => formatTs(initialRun.started_at), [initialRun.started_at]);
  const finished = useMemo(
    () => (initialRun.finished_at ? formatTs(initialRun.finished_at) : null),
    [initialRun.finished_at],
  );

  return (
    <div className="font-mono text-xs space-y-4">
      <div className="flex items-baseline justify-between">
        <h1 className="text-base text-ink truncate pr-2">{initialRun.goal}</h1>
        <span
          className={failed ? "text-accent-rose" : "text-accent-emerald"}
          data-status={failed ? "failed" : "ok"}
        >
          {failed ? "FAILED" : "OK"} · #{initialRun.id}
        </span>
      </div>
      <dl className="grid grid-cols-2 gap-2 text-ink-dim md:grid-cols-4">
        <Field label="subtasks" value={String(subtaskCount)} />
        <Field label="workers" value={String(workerCount)} />
        <Field label="duration" value={`${initialRun.duration_ms ?? 0} ms`} />
        <Field
          label="cost"
          value={`${tokensUsed} / ${tokensBudget}`}
          hint={`${pct}%`}
        />
      </dl>
      <div className="h-1 w-full bg-hairline/40">
        <div
          className={failed ? "h-1 bg-accent-rose" : "h-1 bg-accent-emerald"}
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="text-ink-muted">{initialRun.final_summary}</p>
      {warnings.length > 0 && (
        <ul className="space-y-0.5 text-accent-gold">
          {warnings.map((w, i) => (
            <li key={i}>· {w}</li>
          ))}
        </ul>
      )}
      <p className="text-ink-dim">
        {started} → {finished ?? "…"}
      </p>
      <p className="text-ink-muted">{subtaskCount} subtasks</p>
      <ol className="space-y-1">
        {(initialRun.subtasks ?? []).map((s, i) => (
          <li key={i} className="rounded border border-hairline/40 p-2">
            <div className="text-ink">
              {String(s.subtask ?? "(unnamed subtask)")}
            </div>
            <div className="text-ink-dim">
              {String(s.target_job_type ?? "—")} ·{" "}
              {String(s.target_agent ?? "—")}
              {s.reasoning ? ` — ${String(s.reasoning)}` : null}
            </div>
          </li>
        ))}
      </ol>
      <p className="text-ink-muted">{workerCount} workers</p>
      <ol className="space-y-2">
        {(initialRun.worker_results ?? []).map((w, i) => {
          const errorMsg = typeof w.error === "string" ? w.error : null;
          const receipt = w.receipt as { tx_id?: string; amount?: number } | undefined;
          return (
            <li key={i} className="rounded border border-hairline/40 p-2">
              <div className="flex items-baseline justify-between">
                <span className="text-ink">
                  #{String(w.agent_id ?? "—")}
                </span>
                <span
                  className={errorMsg ? "text-accent-rose" : "text-accent-emerald"}
                >
                  {errorMsg ? "ERROR" : "OK"}
                </span>
              </div>
              {receipt?.tx_id && (
                <div className="text-ink-dim">
                  tx {receipt.tx_id} · {receipt.amount ?? "?"} GOLD
                </div>
              )}
              {errorMsg && (
                <div className="text-accent-rose">{errorMsg}</div>
              )}
            </li>
          );
        })}
      </ol>
      <Link
        href="/"
        className="inline-block rounded border border-hairline/60 px-3 py-1 text-accent-cyan hover:underline"
      >
        ← 返回仪表板
      </Link>
    </div>
  );
}

function Field({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="rounded border border-hairline/40 p-2">
      <div className="text-ink-muted">{label}</div>
      <div className="text-ink">
        {value}
        {hint && <span className="ml-1 text-ink-dim">({hint})</span>}
      </div>
    </div>
  );
}

function formatTs(iso: string): string {
  try {
    return new Date(iso).toLocaleString("zh-CN", { hour12: false });
  } catch {
    return iso;
  }
}
