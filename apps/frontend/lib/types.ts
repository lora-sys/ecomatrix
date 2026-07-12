export type JobType = "miner" | "merchant" | "hacker" | "mediator";

export interface Agent {
  ID: number;
  StringID: string;
  JobType: JobType;
  Balance: number;
  Vitality: number;
  CreditScore: number;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface MetricsSnapshot {
  agent_count: number;
  total_gold: number;
  jobs_breakdown: Record<string, number>;
  recent_qps: number;
  ws_connections: number;
  last_trade_at?: string;
  generated_at: string;
}

export interface Transaction {
  ID: number;
  TxID: string;
  MsgID: string;
  FromID: number;
  ToID: number;
  Amount: number;
  CurrencyType: string;
  Status: "SETTLED" | "REJECTED";
  ErrorCode: string;
  Reasoning: string;
  CreatedAt: string;
}

export type StreamEvent =
  | { type: "trade.settled"; tx_id: string; from: string; to: string; amount: number }
  | { type: "trade.rejected"; msg_id: string; code: string; message?: string }
  | { type: "trade.idempotent_replay"; tx_id: string }
  | { type: "agent.heartbeat"; alive?: number }
  | { type: string; [k: string]: unknown };

export interface MetricsHistorySample {
  at: string;
  agent_count: number;
  total_gold: number;
  recent_qps: number;
  trade_count: number;
}

export interface MetricsHistory {
  window_seconds: number;
  capacity: number;
  count: number;
  samples: MetricsHistorySample[];
}

export interface ConversationEntry {
  id: number;
  agent_id: string;
  role: "user" | "assistant" | "tool" | "system" | "error";
  content: string;
  tool_name?: string;
  error_code?: string;
  latency_ms?: number;
  created_at: string;
}

export interface LLMCacheStats {
  total_entries: number;
  expired_entries: number;
  avg_hit_count: number;
}
