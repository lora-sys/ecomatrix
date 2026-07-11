// Package a2a encodes and validates the Agent-to-Agent protocol v1.1.
//
// The codec is the single source of truth shared with the Python agent and the
// WebSocket fan-out. Any change here is an API break and MUST bump protocol_v.
package a2a

import (
	"regexp"
	"time"
)

// ProtocolVersion is the wire-level protocol version this codec supports.
const ProtocolVersion = "1.1"

// Action is a verb in the A2A protocol.
type Action string

const (
	ActionExecuteTrade Action = "EXECUTE_TRADE"
)

// Allowed actions; anything else is rejected as UNKNOWN_ACTION.
var allowedActions = map[Action]struct{}{
	ActionExecuteTrade: {},
}

// CurrencyType is the unit of account. MVP supports GOLD only.
type CurrencyType string

const (
	CurrencyGold CurrencyType = "GOLD"
)

// Offer describes what an agent is willing to transfer.
type Offer struct {
	CurrencyType CurrencyType `json:"currency_type"`
	Amount       int64        `json:"amount"`
}

// Envelope is the outer wrapper for every A2A message.
type Envelope struct {
	ProtocolV string         `json:"protocol_v"`
	MsgID     string         `json:"msg_id"`
	Sender    string         `json:"sender"`
	Action    Action         `json:"action"`
	Payload   map[string]any `json:"payload"`
	Timestamp int64          `json:"timestamp"`
}

// TradePayload is the typed shape of an EXECUTE_TRADE envelope payload.
type TradePayload struct {
	TargetAgent string `json:"target_agent"`
	Offer       Offer  `json:"offer"`
	Reasoning   string `json:"reasoning,omitempty"`
}

var (
	msgIDRe = regexp.MustCompile(`^[A-Za-z0-9_]{6,64}$`)
	agentRe = regexp.MustCompile(`^agent_[a-z0-9_]{2,32}$`)
)

// MaxClockSkew is the largest acceptable drift between client timestamp and
// server time. The server always wins; clients outside the window are rejected
// unless they pass X-Allow-Old-Timestamp: 1 (admin tooling).
const MaxClockSkew = 300 * time.Second

// IsAllowedAction reports whether the action verb is recognized.
func IsAllowedAction(a Action) bool {
	_, ok := allowedActions[a]
	return ok
}

// ValidateMessageID returns true when the msg_id shape is acceptable.
func ValidateMessageID(s string) bool { return msgIDRe.MatchString(s) }

// ValidateAgentID returns true when the string_id shape is acceptable.
func ValidateAgentID(s string) bool { return agentRe.MatchString(s) }
