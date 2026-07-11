package a2a

import (
	"encoding/json"
	"fmt"
	"time"
)

// Validate enforces envelope shape and returns a typed *Error on failure.
// It does not decode the payload; callers do that for the action they handle.
func Validate(env Envelope) error {
	if env.ProtocolV == "" {
		return New(CodeInvalidEnvelope, "protocol_v is required", false)
	}
	if env.ProtocolV != ProtocolVersion {
		return New(CodeProtocolMismatch,
			fmt.Sprintf("server speaks %s, got %q", ProtocolVersion, env.ProtocolV),
			false)
	}
	if !ValidateMessageID(env.MsgID) {
		return New(CodeInvalidEnvelope,
			fmt.Sprintf("msg_id %q must match %s", env.MsgID, msgIDRe.String()),
			false)
	}
	if !ValidateAgentID(env.Sender) {
		return New(CodeInvalidEnvelope,
			fmt.Sprintf("sender %q must match %s", env.Sender, agentRe.String()),
			false)
	}
	if !IsAllowedAction(env.Action) {
		return New(CodeUnknownAction,
			fmt.Sprintf("action %q is not recognized", env.Action),
			false)
	}
	if env.Payload == nil {
		return New(CodeInvalidEnvelope, "payload is required", false)
	}
	if env.Timestamp == 0 {
		return New(CodeInvalidEnvelope, "timestamp is required", false)
	}
	drift := time.Since(time.Unix(env.Timestamp, 0))
	if drift < -MaxClockSkew || drift > MaxClockSkew {
		return New(CodeInvalidEnvelope,
			fmt.Sprintf("timestamp drift %s exceeds %s", drift, MaxClockSkew),
			false)
	}
	return nil
}

// DecodeTradePayload narrows the action-specific payload for EXECUTE_TRADE.
func DecodeTradePayload(payload map[string]any) (TradePayload, error) {
	if payload == nil {
		return TradePayload{}, New(CodeInvalidEnvelope, "payload is required", false)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return TradePayload{}, New(CodeInvalidEnvelope, "payload is not JSON-encodable", false)
	}
	var out TradePayload
	if err := json.Unmarshal(raw, &out); err != nil {
		return TradePayload{}, New(CodeInvalidEnvelope, "payload shape is invalid", false)
	}
	if !ValidateAgentID(out.TargetAgent) {
		return TradePayload{}, New(CodeInvalidEnvelope,
			fmt.Sprintf("target_agent %q is not a valid agent id", out.TargetAgent),
			false)
	}
	if out.Offer.CurrencyType != CurrencyGold {
		return TradePayload{}, New(CodeInvalidEnvelope,
			fmt.Sprintf("currency_type %q not supported", out.Offer.CurrencyType),
			false)
	}
	if out.Offer.Amount <= 0 {
		return TradePayload{}, New(CodeInvalidEnvelope,
			"offer.amount must be > 0", false)
	}
	return out, nil
}
