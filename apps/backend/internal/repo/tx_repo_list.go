package repo

import (
	"context"

	"github.com/ecomatrix/backend/internal/domain"
)

// Recent returns the latest transactions, newest first.
func (r *TxRepo) Recent(ctx context.Context, limit int) ([]domain.Transaction, error) {
	var rows []TxModel
	if err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Transaction{
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
		})
	}
	return out, nil
}
