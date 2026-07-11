package a2a

import (
	"errors"
	"fmt"
)

// Code is a machine-readable A2A error code.
type Code string

const (
	CodeInvalidEnvelope   Code = "INVALID_ENVELOPE"
	CodeUnknownAction     Code = "UNKNOWN_ACTION"
	CodeProtocolMismatch  Code = "PROTOCOL_MISMATCH"
	CodeUnknownAgent      Code = "UNKNOWN_AGENT"
	CodeInsufficientFunds Code = "INSUFFICIENT_FUNDS"
	CodeSelfTrade         Code = "SELF_TRADE"
	CodeRateLimited       Code = "RATE_LIMITED"
	CodeInternal          Code = "INTERNAL"
	CodeDuplicateReplayOK Code = "IDEMPOTENT_REPLAY"
)

// Error is a typed A2A error suitable for transport mapping.
type Error struct {
	Code      Code
	Message   string
	Retryable bool
}

func (e *Error) Error() string {
	return fmt.Sprintf("a2a: %s: %s", e.Code, e.Message)
}

// New constructs a typed A2A error.
func New(code Code, msg string, retryable bool) *Error {
	return &Error{Code: code, Message: msg, Retryable: retryable}
}

// As extracts a typed *Error if present.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
