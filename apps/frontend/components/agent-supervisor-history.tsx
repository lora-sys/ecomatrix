"use client";

import Link from "next/link";
import { SupervisorRun } from "../lib/types";

interface Props {
  agentId: string;
  initialRuns: SupervisorRun[];
}

export function AgentSupervisorHistory({ agentId, initialRuns }: Props) {
  if (initialRuns.length === 0) {
    return (
      <p className="font-mono text-xs text-ink-dim">
        该 Agent 还没有被 Supervisor 调度过。
      </p>
    );
  }
  return (
    <ul className="space-y-1 font-mono text-xs">
      {initialRuns.map((r) => {
        const failed = r.status === "failed" || Boolean(r.error);
        const tokenPair = `${r.tokens_used ?? 0} / ${r.tokens_budget ?? 0}`;
        return (
          <li
            key={r.id}
            className="flex items-baseline justify-between rounded border border-hairline/40 px-2 py-1"
          >
            <Link
              href={`/supervisor/${r.id}`}
              className="truncate pr-2 text-accent-cyan hover:underline"
            >
              #{r.id} · {r.goal}
            </Link>
            <span
              className={failed ? "text-accent-rose" : "text-accent-emerald"}
            >
              {failed ? "FAILED" : "OK"} · {r.duration_ms ?? 0} ms · {tokenPair}
            </span>
          </li>
        );
      })}
    </ul>
  );
}
