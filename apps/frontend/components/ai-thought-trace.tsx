"use client";

import { useEffect, useState } from "react";
import clsx from "clsx";
import { fetchConversations } from "../lib/api";
import type { ConversationEntry } from "../lib/types";

const ROLE_COLOR: Record<string, string> = {
  user: "text-accent-cyan",
  assistant: "text-accent-emerald",
  tool: "text-accent-violet",
  system: "text-ink-muted",
  error: "text-accent-rose",
};

const ROLE_LABEL: Record<string, string> = {
  user: "用户",
  assistant: "助手",
  tool: "工具",
  system: "系统",
  error: "错误",
};

export function AIThoughtTrace({ agentId }: { agentId: string }) {
  const [entries, setEntries] = useState<ConversationEntry[]>([]);

  useEffect(() => {
    let stopped = false;
    const tick = async () => {
      try {
        const convs = await fetchConversations(agentId, 20);
        if (!stopped) setEntries(convs);
      } catch {}
    };
    tick();
    const id = setInterval(tick, 3000);
    return () => {
      stopped = true;
      clearInterval(id);
    };
  }, [agentId]);

  if (entries.length === 0) {
    return (
      <div className="font-mono text-xs text-ink-dim">
        等待 LLM 思考…
      </div>
    );
  }

  return (
    <ol
      className="space-y-2 font-mono text-xs"
      role="log"
      aria-live="polite"
      aria-label="AI thought trace"
    >
      {entries.slice(0, 8).map((e) => (
        <li key={e.id} className="border-l border-hairline pl-2">
          <div className="flex items-baseline justify-between gap-2">
            <span className={clsx(ROLE_COLOR[e.role] || "text-ink-muted")}>
              [{ROLE_LABEL[e.role] || e.role}]
            </span>
            {e.tool_name ? (
              <span className="text-ink-muted">→ {e.tool_name}</span>
            ) : null}
            {e.error_code ? (
              <span className="text-accent-rose">{e.error_code}</span>
            ) : null}
            {e.latency_ms != null ? (
              <span className="text-ink-dim">{e.latency_ms}ms</span>
            ) : null}
            <time className="ml-auto text-ink-dim">
              {new Date(e.created_at).toLocaleTimeString("zh-CN", { hour12: false })}
            </time>
          </div>
          <div className="mt-0.5 line-clamp-2 text-ink-muted">
            {e.content}
          </div>
        </li>
      ))}
    </ol>
  );
}
