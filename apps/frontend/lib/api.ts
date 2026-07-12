import { Agent, MetricsHistory, MetricsSnapshot, Transaction } from "./types";

const BASE = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

export async function fetchAgents(): Promise<Agent[]> {
  const r = await fetch(`${BASE}/v1/agents?limit=100`, { cache: "no-store" });
  if (!r.ok) throw new Error(`agents: ${r.status}`);
  const d = await r.json();
  return d.agents as Agent[];
}

export async function fetchAgent(id: string): Promise<Agent> {
  const r = await fetch(`${BASE}/v1/agents/by-string-id/${id}`, { cache: "no-store" });
  if (!r.ok) throw new Error(`agent: ${r.status}`);
  return (await r.json()) as Agent;
}

export async function fetchMetrics(): Promise<MetricsSnapshot> {
  const r = await fetch(`${BASE}/v1/metrics`, { cache: "no-store" });
  if (!r.ok) throw new Error(`metrics: ${r.status}`);
  return (await r.json()) as MetricsSnapshot;
}

export async function fetchTransactions(limit = 50): Promise<Transaction[]> {
  const r = await fetch(`${BASE}/v1/transactions?limit=${limit}`, { cache: "no-store" });
  if (!r.ok) throw new Error(`transactions: ${r.status}`);
  const d = await r.json();
  return d.transactions as Transaction[];
}

export interface LongTermMemory {
  summary: string;
  facts: string[];
}

export async function fetchLongTermMemory(id: string): Promise<LongTermMemory> {
  const r = await fetch(`${BASE}/v1/agents/by-string-id/${id}/long-term-memory`, {
    cache: "no-store",
  });
  if (!r.ok) throw new Error(`ltm: ${r.status}`);
  const d = await r.json();
  return d.long_term_memory as LongTermMemory;
}

export async function fetchMetricsHistory(): Promise<MetricsHistory> {
  const r = await fetch(`${BASE}/v1/metrics/history`, { cache: "no-store" });
  if (!r.ok) throw new Error(`history: ${r.status}`);
  return (await r.json()) as MetricsHistory;
}
