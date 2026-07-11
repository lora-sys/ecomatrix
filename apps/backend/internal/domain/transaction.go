package domain

import "time"

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

// Receipt is the success response returned by the trade service.
type Receipt struct {
	TxID         string `json:"tx_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       int64  `json:"amount"`
	CurrencyType string `json:"currency_type"`
	BalanceAfter struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	} `json:"balance_after"`
}
