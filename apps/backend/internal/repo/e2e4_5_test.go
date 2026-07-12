package repo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConversationsRepo_Roundtrip(t *testing.T) {
	db := testDBRepo(t)
	wipeRepo(t, db)
	agents := repo.NewAgentRepo(db)
	_, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_conv_test", JobType: domain.JobMiner,
		Balance: 100, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)

	c := repo.NewConversationsRepo(db)
	inserted, err := c.Insert(context.Background(), domain.Conversation{
		AgentID:   "agent_conv_test",
		Role:      "assistant",
		Content:   "I should trade 10 GOLD with merchant_03",
		LatencyMS: 123,
	})
	require.NoError(t, err)
	assert.NotZero(t, inserted.ID)

	// Read back.
	recent, err := c.Recent(context.Background(), "agent_conv_test", 10)
	require.NoError(t, err)
	require.NotEmpty(t, recent)
	assert.Equal(t, "assistant", recent[0].Role)
	assert.Equal(t, "I should trade 10 GOLD with merchant_03", recent[0].Content)
}

func TestLLMCacheRepo_RoundtripWithTTL(t *testing.T) {
	db := testDBRepo(t)
	wipeRepo(t, db)
	c := repo.NewLLMCacheRepo(db)

	// Set with 1-hour TTL.
	require.NoError(t, c.Set(context.Background(),
		"abc123", "gpt-4o-mini", "prompt-hash",
		`{"action":"EXECUTE_TRADE"}`, 1*time.Hour))

	// Get should hit.
	resp, ok, err := c.Get(context.Background(), "abc123")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, resp, "EXECUTE_TRADE")

	// Stats: total=1, expired=0.
	stats, err := c.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.TotalEntries)
	assert.Equal(t, int64(0), stats.ExpiredEntries)
}

func TestLLMCacheRepo_ExpiredMisses(t *testing.T) {
	db := testDBRepo(t)
	wipeRepo(t, db)
	c := repo.NewLLMCacheRepo(db)

	// Insert with an expires_at in the past via raw SQL.
	require.NoError(t, db.Exec(`INSERT INTO llm_cache (key, model, response, prompt_hash, created_at, expires_at) VALUES (?, ?, ?, ?, now() - interval '2 hours', now() - interval '1 hour')`,
		"expired-key", "m", "resp", "ph").Error)

	_, ok, err := c.Get(context.Background(), "expired-key")
	require.NoError(t, err)
	assert.False(t, ok, "expired entry should miss")

	// Evict and verify.
	n, err := c.EvictExpired(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	var count int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM llm_cache`).Scan(&count).Error)
	assert.Equal(t, int64(0), count)
}

func testDBRepo(t *testing.T) *gorm.DB {
	dsn := os.Getenv("ECOMATRIX_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(db))
	return db
}
