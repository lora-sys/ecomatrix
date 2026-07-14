// Package domain holds the enterprise entities and errors.
// Domain code must not import infrastructure packages (no GORM, no Fiber).
package domain

import "errors"

var (
	ErrAgentNotFound         = errors.New("agent not found")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrSelfTrade             = errors.New("self-trade is not allowed")
	ErrDuplicateReplay       = errors.New("duplicate msg_id, returning original receipt")
	ErrSupervisorRunNotFound = errors.New("supervisor run not found")
)
