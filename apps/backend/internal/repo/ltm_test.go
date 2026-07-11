package repo_test

import (
	"context"
	"os"
	"testing"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAgentRepo_LongTermMemory_Roundtrip(t *testing.T) {
	db := openTestDBRepo(t)
	wipeRepo(t, db)

	agents := repo.NewAgentRepo(db)
	a, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_ltm_test_01", JobType: domain.JobMiner,
		Balance: 100, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)

	// Initial LTM is empty.
	ltm, err := agents.GetLongTermMemory(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Empty(t, ltm.Summary)
	assert.Empty(t, ltm.Facts)

	// Set LTM.
	want := domain.LongTermMemory{
		Summary: "low on vitality, traded with merchant_03",
		Facts:   []string{"bought food", "sold ore", "asked mediator for help"},
	}
	require.NoError(t, agents.SetLongTermMemory(context.Background(), a.ID, want))

	got, err := agents.GetLongTermMemory(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Equal(t, want.Summary, got.Summary)
	assert.Equal(t, want.Facts, got.Facts)
}

func openTestDBRepo(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("ECOMATRIX_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(db))
	return db
}

func wipeRepo(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE transactions, social_feeds, agents RESTART IDENTITY CASCADE").Error)
}
