package repo

import (
	"context"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

type TraceModel struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	AgentID    string    `gorm:"column:agent_id;index" json:"agent_id"`
	Kind       string    `gorm:"column:kind" json:"kind"`
	Content    string    `gorm:"column:content" json:"content"`
	LatencyMS  *int      `gorm:"column:latency_ms" json:"latency_ms,omitempty"`
	TokensIn   *int      `gorm:"column:tokens_in" json:"tokens_in,omitempty"`
	TokensOut  *int      `gorm:"column:tokens_out" json:"tokens_out,omitempty"`
	ToolName   string    `gorm:"column:tool_name" json:"tool_name"`
	ToolInput  []byte    `gorm:"column:tool_input;type:jsonb" json:"tool_input,omitempty"`
	ToolOutput []byte    `gorm:"column:tool_output;type:jsonb" json:"tool_output,omitempty"`
	CostUSD    *float64  `gorm:"column:cost_usd" json:"cost_usd,omitempty"`
	ErrorCode  string    `gorm:"column:error_code" json:"error_code"`
	ParentID   *int64    `gorm:"column:parent_id" json:"parent_id,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (TraceModel) TableName() string { return "traces" }

type TracesRepo struct{ db *gorm.DB }

func NewTracesRepo(db *gorm.DB) *TracesRepo {
	return &TracesRepo{db: db}
}

func (r *TracesRepo) Insert(ctx context.Context, t domain.Trace) (domain.Trace, error) {
	row := TraceModel{
		AgentID:   t.AgentID,
		Kind:      t.Kind,
		Content:   t.Content,
		LatencyMS: t.LatencyMS,
		TokensIn:  t.TokensIn,
		TokensOut: t.TokensOut,
		ToolName:  t.ToolName,
		ToolInput: t.ToolInput,
		ToolOutput: t.ToolOutput,
		CostUSD:   t.CostUSD,
		ErrorCode: t.ErrorCode,
		ParentID:  t.ParentID,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Trace{}, err
	}
	return toDomainTrace(row), nil
}

func (r *TracesRepo) Recent(ctx context.Context, agentID string, limit int) ([]domain.Trace, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var rows []TraceModel
	if err := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Trace, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainTrace(row))
	}
	return out, nil
}

func toDomainTrace(t TraceModel) domain.Trace {
	return domain.Trace{
		ID:         t.ID,
		AgentID:    t.AgentID,
		Kind:       t.Kind,
		Content:    t.Content,
		LatencyMS:  t.LatencyMS,
		TokensIn:   t.TokensIn,
		TokensOut:  t.TokensOut,
		ToolName:   t.ToolName,
		ToolInput:  t.ToolInput,
		ToolOutput: t.ToolOutput,
		CostUSD:    t.CostUSD,
		ErrorCode:  t.ErrorCode,
		ParentID:   t.ParentID,
		CreatedAt:  t.CreatedAt,
	}
}
