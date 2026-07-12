package repo

import (
	"context"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

type ConversationModel struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	AgentID    string    `gorm:"column:agent_id;index" json:"agent_id"`
	Role       string    `gorm:"column:role" json:"role"`
	Content    string    `gorm:"column:content" json:"content"`
	ToolName   string    `gorm:"column:tool_name" json:"tool_name"`
	ToolInput  []byte    `gorm:"column:tool_input;type:jsonb" json:"tool_input"`
	ToolOutput []byte    `gorm:"column:tool_output;type:jsonb" json:"tool_output"`
	ErrorCode  string    `gorm:"column:error_code" json:"error_code"`
	LatencyMS  int       `gorm:"column:latency_ms" json:"latency_ms"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ConversationModel) TableName() string { return "conversations" }

func toDomainConv(c ConversationModel) domain.Conversation {
	return domain.Conversation{
		ID:         c.ID,
		AgentID:    c.AgentID,
		Role:       c.Role,
		Content:    c.Content,
		ToolName:   c.ToolName,
		ToolInput:  c.ToolInput,
		ToolOutput: c.ToolOutput,
		ErrorCode:  c.ErrorCode,
		LatencyMS:  c.LatencyMS,
		CreatedAt:  c.CreatedAt,
	}
}

type ConversationsRepo struct{ db *gorm.DB }

func NewConversationsRepo(db *gorm.DB) *ConversationsRepo {
	return &ConversationsRepo{db: db}
}

func (r *ConversationsRepo) Insert(ctx context.Context, c domain.Conversation) (domain.Conversation, error) {
	row := ConversationModel{
		AgentID:    c.AgentID,
		Role:       c.Role,
		Content:    c.Content,
		ToolName:   c.ToolName,
		ToolInput:  c.ToolInput,
		ToolOutput: c.ToolOutput,
		ErrorCode:  c.ErrorCode,
		LatencyMS:  c.LatencyMS,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Conversation{}, err
	}
	return toDomainConv(row), nil
}

// Recent returns the last N conversations for an agent, newest first.
func (r *ConversationsRepo) Recent(ctx context.Context, agentID string, limit int) ([]domain.Conversation, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []ConversationModel
	if err := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Conversation, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainConv(r))
	}
	return out, nil
}
