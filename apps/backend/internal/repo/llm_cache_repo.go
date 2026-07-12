package repo

import (
	"context"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

type LLMCacheModel struct {
	Key       string    `gorm:"primaryKey;column:key"`
	Model     string    `gorm:"column:model"`
	Response  string    `gorm:"column:response"`
	PromptHash string    `gorm:"column:prompt_hash"`
	CreatedAt time.Time `gorm:"column:created_at"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	HitCount  int       `gorm:"column:hit_count"`
}

func (LLMCacheModel) TableName() string { return "llm_cache" }

type LLMCacheRepo struct{ db *gorm.DB }

func NewLLMCacheRepo(db *gorm.DB) *LLMCacheRepo {
	return &LLMCacheRepo{db: db}
}

// Get returns the cached response for a key, or ("", false, nil) on miss.
// Expired entries are treated as misses.
func (r *LLMCacheRepo) Get(ctx context.Context, key string) (string, bool, error) {
	var row LLMCacheModel
	if err := r.db.WithContext(ctx).Where("key = ? AND expires_at > now()", key).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	// Increment hit_count.
	_ = r.db.WithContext(ctx).Exec(`UPDATE llm_cache SET hit_count = hit_count + 1 WHERE key = ?`, key).Error
	return row.Response, true, nil
}

// Set inserts or refreshes a cache entry.
func (r *LLMCacheRepo) Set(ctx context.Context, key, modelName, promptHash, response string, ttl time.Duration) error {
	now := time.Now()
	row := LLMCacheModel{
		Key:        key,
		Model:      modelName,
		PromptHash: promptHash,
		Response:   response,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO llm_cache (key, model, response, prompt_hash, created_at, expires_at, hit_count)
		VALUES (?, ?, ?, ?, ?, ?, 0)
		ON CONFLICT (key) DO UPDATE SET
			model = EXCLUDED.model,
			response = EXCLUDED.response,
			prompt_hash = EXCLUDED.prompt_hash,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at,
			hit_count = 0
	`, row.Key, row.Model, row.Response, row.PromptHash, row.CreatedAt, row.ExpiresAt).Error
}

// EvictExpired deletes expired entries. Called by a periodic task in main.
func (r *LLMCacheRepo) EvictExpired(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Exec(`DELETE FROM llm_cache WHERE expires_at < now()`)
	return res.RowsAffected, res.Error
}

// Stats returns hit/miss counters for the dashboard.
func (r *LLMCacheRepo) Stats(ctx context.Context) (domain.LLMCacheStats, error) {
	var total, expired, avgHits int64
	_ = r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM llm_cache`).Scan(&total).Error
	_ = r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM llm_cache WHERE expires_at < now()`).Scan(&expired).Error
	_ = r.db.WithContext(ctx).Raw(`SELECT COALESCE(AVG(hit_count), 0)::bigint FROM llm_cache`).Scan(&avgHits).Error
	return domain.LLMCacheStats{TotalEntries: total, ExpiredEntries: expired, AvgHitCount: avgHits}, nil
}
