package domain

import (
	"time"
)

type TxStatus string

const (
	TxSettled  TxStatus = "SETTLED"
	TxRejected TxStatus = "REJECTED"
)

// Transaction is the immutable ledger entry. Failed trades still produce a row
// with status=REJECTED so the history is auditable.
type Transaction struct {
	ID           int64
	TxID         string
	MsgID        string
	FromID       int64
	ToID         int64
	Amount       int64
	CurrencyType string
	Status       TxStatus
	ErrorCode    string
	Reasoning    string
	CreatedAt    time.Time
}

// Conversation is one entry in an agent's LLM interaction log.
type Conversation struct {
	ID         int64     `json:"id"`
	AgentID    string    `json:"agent_id"`
	Role       string    `json:"role"` // user | assistant | tool | system | error
	Content    string    `json:"content"`
	ToolName   string    `json:"tool_name"`
	ToolInput  []byte    `json:"tool_input"`
	ToolOutput []byte    `json:"tool_output"`
	ErrorCode  string    `json:"error_code"`
	LatencyMS  int       `json:"latency_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// LLMCacheStats summarises cache health for the dashboard.
type LLMCacheStats struct {
	TotalEntries   int64 `json:"total_entries"`
	ExpiredEntries int64 `json:"expired_entries"`
	AvgHitCount    int64 `json:"avg_hit_count"`
}

// FeedPost is an entry in the agent social square.
type FeedPost struct {
	ID         int64
	AgentID    int64
	Content    string
	IntentType string
	CreatedAt  time.Time
}

// Receipt is what the trade service returns to the transport layer.
type Receipt struct {
	TxID         string
	From         string
	To           string
	Amount       int64
	CurrencyType string
	BalanceAfter struct {
		From int64
		To   int64
	} `json:"balance_after"`
}

// Trace is one observability event: a plan, decision, tool call, or error.
type Trace struct {
	ID         int64     `json:"id"`
	AgentID    string    `json:"agent_id"`
	Kind       string    `json:"kind"` // plan | decision | tool_call | tool_result | error | observation | reflection
	Content    string    `json:"content"`
	LatencyMS  *int      `json:"latency_ms,omitempty"`
	TokensIn   *int      `json:"tokens_in,omitempty"`
	TokensOut  *int      `json:"tokens_out,omitempty"`
	ToolName   string    `json:"tool_name"`
	ToolInput  []byte    `json:"tool_input,omitempty"`
	ToolOutput []byte    `json:"tool_output,omitempty"`
	CostUSD    *float64  `json:"cost_usd,omitempty"`
	ErrorCode  string    `json:"error_code"`
	ParentID   *int64    `json:"parent_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
