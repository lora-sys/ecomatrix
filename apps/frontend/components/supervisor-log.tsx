"use client";

import { useMemo } from "react";
import { useStore } from "../hooks/store";
import Link from "next/link";
import { SupervisorRun } from "../lib/types";

interface SupervisorLogProps {
  initialRuns?: SupervisorRun[];
}

export function SupervisorLog({ initialRuns = [] }: SupervisorLogProps) {
  const liveRuns = useStore((s) => s.supervisorRuns);
  const liveLatest = useStore((s) => s.supervisorLatest);
  const isHydrated = useStore((s) => s.supervisorHydrated);

  const runs = useMemo<SupervisorRun[]>(() => {
    const byId = new Map<number, SupervisorRun>();
    for (const r of initialRuns) if (r.id !== undefined) byId.set(r.id, r);
    for (const r of liveRuns) if (r.id !== undefined) byId.set(r.id, r);
    const sorted = Array.from(byId.values()).sort(
      (a, b) => (b.id ?? 0) - (a.id ?? 0)
    );
    return sorted.slice(0, 6);
  }, [initialRuns, liveRuns]);

  const latest: SupervisorRun | undefined = liveLatest ?? runs[0];

  if (!isHydrated && runs.length === 0) {
    return (
      <div className="font-mono text-xs text-ink-dim">
        等待 Supervisor 的第一条调度…
      </div>
    );
  }
  if (!latest) {
    return (
      <div className="font-mono text-xs text-ink-dim">
        等待 Supervisor 的第一条调度…
      </div>
    );
  }

  const used = latest.tokens_used ?? 0;
  const budget = latest.tokens_budget ?? 0;
  const pct = budget > 0 ? Math.min(100, Math.round((used / budget) * 100)) : 0;
  const failed = latest.status === "failed" || Boolean(latest.error);
  const subtaskCount = latest.subtasks?.length ?? 0;
  const workerCount = latest.worker_results?.length ?? 0;
  const duration = latest.duration_ms ?? 0;
  const warnings = latest.warnings ?? [];

  return (
    <div
      aria-label="supervisor task log"
      aria-live="polite"
      className="font-mono text-xs space-y-3"
    >
      <div className="rounded border border-hairline/60 bg-panel/40 p-3">
        <div className="flex items-baseline justify-between">
          <Link
            href={`/supervisor/${latest.id ?? 0}`}
            className="truncate pr-2 text-ink hover:text-accent-cyan"
            data-supervisor-link
          >
            {latest.goal || "(无目标)"}
          </Link>
          <span className="flex items-center gap-2">
            <span
              className={failed ? "text-accent-rose" : "text-accent-emerald"}
              data-status={failed ? "failed" : "ok"}
            >
              {failed ? "FAILED" : "OK"}
            </span>
            <Link
              href={`/supervisor/${latest.id ?? 0}`}
              className="font-mono text-[10px] text-accent-cyan hover:underline"
              data-supervisor-link
            >
              详情 #{latest.id ?? "?"}
            </Link>
          </span>
        </div>
        <div className="mt-1 flex items-baseline justify-between text-ink-dim">
          <span>
            {subtaskCount} subtasks · {workerCount} workers · {duration} ms
          </span>
          <span>
            {used} / {budget} tokens
          </span>
        </div>
        <div className="mt-2 h-1 w-full bg-hairline/40">
          <div
            className={failed ? "h-1 bg-accent-rose" : "h-1 bg-accent-emerald"}
            style={{ width: `${pct}%` }}
          />
        </div>
        {latest.final_summary && (
          <p className="mt-2 text-ink-muted">{latest.final_summary}</p>
        )}
        {warnings.length > 0 && (
          <ul className="mt-1 space-y-0.5 text-accent-gold">
            {warnings.map((w, i) => (
              <li key={i}>· {w}</li>
            ))}
          </ul>
        )}
      </div>
      {runs.length > 0 && (
        <div className="space-y-1 text-ink-dim">
          <div className="flex items-baseline justify-between font-mono text-[10px] uppercase tracking-[0.18em]">
            <span>近期运行</span>
            <span className="text-accent-cyan">详情 →</span>
          </div>
          <ul className="space-y-1">
            {runs.slice(1).map((r) => (
              <li
                key={r.id}
                className="flex items-baseline justify-between rounded border border-hairline/40 px-2 py-1 hover:border-accent-cyan/60"
              >
                <Link
                  href={`/supervisor/${r.id}`}
                  className="flex-1 truncate pr-2 font-mono text-xs hover:text-accent-cyan"
                  data-supervisor-link
                >
                  #{r.id} · {r.goal}
                </Link>
                <span className="font-mono text-xs">
                  {r.duration_ms ?? 0} ms
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
