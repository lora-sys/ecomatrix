package repo

import (
	"context"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

// FeedModel is the GORM projection of social_feeds.
type FeedModel struct {
	ID         int64     `gorm:"primaryKey"`
	AgentID    int64     `gorm:"column:agent_id"`
	Content    string    `gorm:"column:content"`
	IntentType string    `gorm:"column:intent_type"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (FeedModel) TableName() string { return "social_feeds" }

func toDomainFeed(f FeedModel) domain.FeedPost {
	return domain.FeedPost{
		ID:         f.ID,
		AgentID:    f.AgentID,
		Content:    f.Content,
		IntentType: f.IntentType,
		CreatedAt:  f.CreatedAt,
	}
}

// FeedRepo writes and reads social_feeds rows.
type FeedRepo struct{ db *gorm.DB }

func NewFeedRepo(db *gorm.DB) *FeedRepo { return &FeedRepo{db: db} }

// Insert writes a feed post.
func (r *FeedRepo) Insert(ctx context.Context, post domain.FeedPost) (domain.FeedPost, error) {
	row := FeedModel{
		AgentID:    post.AgentID,
		Content:    post.Content,
		IntentType: post.IntentType,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.FeedPost{}, err
	}
	return toDomainFeed(row), nil
}

// Recent returns the latest feed posts, newest first.
func (r *FeedRepo) Recent(ctx context.Context, limit int) ([]domain.FeedPost, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []FeedModel
	if err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.FeedPost, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainFeed(r))
	}
	return out, nil
}

// ByAgent returns posts by one agent, newest first.
func (r *FeedRepo) ByAgent(ctx context.Context, agentID int64, limit int) ([]domain.FeedPost, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []FeedModel
	if err := r.db.WithContext(ctx).Where("agent_id = ?", agentID).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.FeedPost, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainFeed(r))
	}
	return out, nil
}
