// Package service contains business logic. Services own the transaction
// boundary; repos never open a tx on their own.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/pkg/a2a"
	"gorm.io/gorm"
)

// Publisher is anything that can broadcast events; satisfied by ws.Hub.
type Publisher interface {
	Publish(event any)
}

// TradeService orchestrates A2A EXECUTE_TRADE requests.
type TradeService struct {
	db        *gorm.DB
	agents    *repo.AgentRepo
	txs       *repo.TxRepo
	publisher Publisher
	metrics   *MetricsService
}

func NewTradeService(db *gorm.DB, agents *repo.AgentRepo, txs *repo.TxRepo, pub Publisher, metrics *MetricsService) *TradeService {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &TradeService{db: db, agents: agents, txs: txs, publisher: pub, metrics: metrics}
}

type noopPublisher struct{}

func (noopPublisher) Publish(any) {}

// SettleResult is what the trade service returns to the transport.
type SettleResult struct {
	Receipt domain.Receipt
	Replay  bool
}

// Settle executes the trade atomically. Concurrency guarantees:
//   - Two agent rows are locked in ascending id order (deadlock-safe).
//   - The transactions table has a UNIQUE constraint on msg_id, providing
//     idempotency at the DB layer even if the service is restarted mid-call.
//   - Balance deltas are applied only when validation passes.
func (s *TradeService) Settle(ctx context.Context, env a2a.Envelope, payload a2a.TradePayload) (SettleResult, *a2a.Error) {
	if env.Sender == payload.TargetAgent {
		return SettleResult{}, a2a.New(a2a.CodeSelfTrade, "sender and target are the same agent", false)
	}

	// Idempotency: short-circuit on duplicate msg_id.
	if existing, err := s.txs.ByMsgID(ctx, env.MsgID); err == nil {
		if existing.Status == domain.TxSettled {
			r, lerr := s.receiptForTx(ctx, existing)
			if lerr != nil {
				return SettleResult{}, a2a.New(a2a.CodeInternal, "rebuild receipt: "+lerr.Error(), true)
			}
			s.publisher.Publish(map[string]any{"type": "trade.idempotent_replay", "tx_id": existing.TxID})
			return SettleResult{Receipt: r, Replay: true}, nil
		}
		// Previously rejected; replay the rejection.
		return SettleResult{}, a2a.New(a2a.Code(existing.ErrorCode), existing.Reasoning, false)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return SettleResult{}, a2a.New(a2a.CodeInternal, "lookup msg_id: "+err.Error(), true)
	}

	var receipt domain.Receipt
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Resolve ids.
		sender, err := s.agents.ByStringID(ctx, env.Sender)
		if err != nil {
			if errors.Is(err, domain.ErrAgentNotFound) {
				s.recordRejected(ctx, tx, env, payload, a2a.CodeUnknownAgent, "sender unknown")
				return a2a.New(a2a.CodeUnknownAgent, "sender unknown", false)
			}
			return err
		}
		target, err := s.agents.ByStringID(ctx, payload.TargetAgent)
		if err != nil {
			if errors.Is(err, domain.ErrAgentNotFound) {
				s.recordRejected(ctx, tx, env, payload, a2a.CodeUnknownAgent, "target unknown")
				return a2a.New(a2a.CodeUnknownAgent, "target unknown", false)
			}
			return err
		}

		// Lock both rows in ascending id order.
		senderLocked, targetLocked, err := s.agents.LockPair(ctx, tx, sender.ID, target.ID)
		if err != nil {
			if errors.Is(err, domain.ErrAgentNotFound) {
				s.recordRejected(ctx, tx, env, payload, a2a.CodeUnknownAgent, "agent row vanished")
				return a2a.New(a2a.CodeUnknownAgent, "agent row vanished", true)
			}
			return err
		}

		if senderLocked.Balance < payload.Offer.Amount {
			s.recordRejected(ctx, tx, env, payload, a2a.CodeInsufficientFunds,
				fmt.Sprintf("%s balance %d < offer.amount %d",
					senderLocked.StringID, senderLocked.Balance, payload.Offer.Amount))
			return a2a.New(a2a.CodeInsufficientFunds,
				fmt.Sprintf("%s balance %d < offer.amount %d",
					senderLocked.StringID, senderLocked.Balance, payload.Offer.Amount),
				false)
		}

		if err := s.agents.ApplyDelta(ctx, tx, senderLocked.ID, -payload.Offer.Amount); err != nil {
			return err
		}
		if err := s.agents.ApplyDelta(ctx, tx, targetLocked.ID, payload.Offer.Amount); err != nil {
			return err
		}

		txID, err := newTxID()
		if err != nil {
			return err
		}
		row, err := s.txs.Insert(ctx, tx, domain.Transaction{
			TxID:         txID,
			MsgID:        env.MsgID,
			FromID:       senderLocked.ID,
			ToID:         targetLocked.ID,
			Amount:       payload.Offer.Amount,
			CurrencyType: string(payload.Offer.CurrencyType),
			Status:       domain.TxSettled,
			Reasoning:    payload.Reasoning,
		})
		if err != nil {
			return err
		}

		receipt = domain.Receipt{
			TxID:         row.TxID,
			From:         senderLocked.StringID,
			To:           targetLocked.StringID,
			Amount:       payload.Offer.Amount,
			CurrencyType: row.CurrencyType,
		}
		receipt.BalanceAfter.From = senderLocked.Balance - payload.Offer.Amount
		receipt.BalanceAfter.To = targetLocked.Balance + payload.Offer.Amount
		return nil
	})
	if err != nil {
		if a2aErr, ok := a2a.As(err); ok {
			s.publisher.Publish(map[string]any{
				"type":    "trade.rejected",
				"msg_id":  env.MsgID,
				"code":    string(a2aErr.Code),
				"message": a2aErr.Message,
			})
			return SettleResult{}, a2aErr
		}
		return SettleResult{}, a2a.New(a2a.CodeInternal, err.Error(), true)
	}

	s.publisher.Publish(map[string]any{
		"type":   "trade.settled",
		"tx_id":  receipt.TxID,
		"from":   receipt.From,
		"to":     receipt.To,
		"amount": receipt.Amount,
	})
	if s.metrics != nil {
		s.metrics.NoteTrade()
	}
	return SettleResult{Receipt: receipt}, nil
}

func (s *TradeService) recordRejected(ctx context.Context, tx *gorm.DB, env a2a.Envelope, payload a2a.TradePayload, code a2a.Code, msg string) {
	// Rejected trades still produce a row so the audit log is complete. We
	// resolve ids best-effort; if either is unknown we record with zero ids.
	fromID, toID := int64(0), int64(0)
	if sender, err := s.agents.ByStringID(ctx, env.Sender); err == nil {
		fromID = sender.ID
	}
	if target, err := s.agents.ByStringID(ctx, payload.TargetAgent); err == nil {
		toID = target.ID
	}
	_, _ = s.txs.Insert(ctx, tx, domain.Transaction{
		TxID:         "rej_" + env.MsgID,
		MsgID:        env.MsgID,
		FromID:       fromID,
		ToID:         toID,
		Amount:       payload.Offer.Amount,
		CurrencyType: string(payload.Offer.CurrencyType),
		Status:       domain.TxRejected,
		ErrorCode:    string(code),
		Reasoning:    msg,
	})
}

func (s *TradeService) receiptForTx(ctx context.Context, tx domain.Transaction) (domain.Receipt, error) {
	from, err := s.agents.ByID(ctx, tx.FromID)
	if err != nil {
		return domain.Receipt{}, err
	}
	to, err := s.agents.ByID(ctx, tx.ToID)
	if err != nil {
		return domain.Receipt{}, err
	}
	r := domain.Receipt{
		TxID:         tx.TxID,
		From:         from.StringID,
		To:           to.StringID,
		Amount:       tx.Amount,
		CurrencyType: tx.CurrencyType,
	}
	r.BalanceAfter.From = from.Balance
	r.BalanceAfter.To = to.Balance
	return r, nil
}

func newTxID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "tx_" + hex.EncodeToString(b[:]), nil
}
