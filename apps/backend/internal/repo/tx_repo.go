package repo

import (
	"context"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

// TxRepo writes transaction rows.
type TxRepo struct{ db *gorm.DB }

func NewTxRepo(db *gorm.DB) *TxRepo { return &TxRepo{db: db} }

// Insert writes a transaction inside the given tx.
func (r *TxRepo) Insert(ctx context.Context, tx *gorm.DB, t domain.Transaction) (domain.Transaction, error) {
	row := TxModel{
		TxID:         t.TxID,
		MsgID:        t.MsgID,
		FromID:       t.FromID,
		ToID:         t.ToID,
		Amount:       t.Amount,
		CurrencyType: t.CurrencyType,
		Status:       string(t.Status),
		ErrorCode:    t.ErrorCode,
		Reasoning:    t.Reasoning,
	}
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Transaction{}, err
	}
	return domain.Transaction{
		ID:           row.ID,
		TxID:         row.TxID,
		MsgID:        row.MsgID,
		FromID:       row.FromID,
		ToID:         row.ToID,
		Amount:       row.Amount,
		CurrencyType: row.CurrencyType,
		Status:       domain.TxStatus(row.Status),
		ErrorCode:    row.ErrorCode,
		Reasoning:    row.Reasoning,
		CreatedAt:    row.CreatedAt,
	}, nil
}

// ByMsgID returns a previously recorded transaction by msg_id (idempotency).
func (r *TxRepo) ByMsgID(ctx context.Context, msgID string) (domain.Transaction, error) {
	var row TxModel
	if err := r.db.WithContext(ctx).Where("msg_id = ?", msgID).First(&row).Error; err != nil {
		return domain.Transaction{}, err
	}
	return domain.Transaction{
		ID:           row.ID,
		TxID:         row.TxID,
		MsgID:        row.MsgID,
		FromID:       row.FromID,
		ToID:         row.ToID,
		Amount:       row.Amount,
		CurrencyType: row.CurrencyType,
		Status:       domain.TxStatus(row.Status),
		ErrorCode:    row.ErrorCode,
		Reasoning:    row.Reasoning,
		CreatedAt:    row.CreatedAt,
	}, nil
}
