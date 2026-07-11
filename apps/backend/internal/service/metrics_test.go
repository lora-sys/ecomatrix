package service_test

import (
	"context"
	"testing"

	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_Collect_EmptyDB(t *testing.T) {
	db := testDB(t)
	wipe(t, db)

	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	m := service.NewMetricsService(db, agents, txs)

	snap, err := m.Collect(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 0, snap.AgentCount)
	assert.Equal(t, int64(0), snap.TotalGold)
	assert.NotNil(t, snap.JobsBreakdown)
	assert.Empty(t, snap.LastTradeAt)
	assert.NotEmpty(t, snap.GeneratedAt)
}

func TestMetricsService_NoteTrade_QPSWindow(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	m := service.NewMetricsService(db, agents, txs)

	for i := 0; i < 5; i++ {
		m.NoteTrade()
	}
	snap, err := m.Collect(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, 2, snap.WSConnections)
	assert.Greater(t, snap.RecentQPS, 0.0)
	assert.NotEmpty(t, snap.LastTradeAt)
}

func TestMetricsService_Collect_JobsBreakdown(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	agents := repo.NewAgentRepo(db)

	for i, sid := range []string{"agent_test_a", "agent_test_b", "agent_test_c"} {
		jt := "miner"
		if i == 1 {
			jt = "merchant"
		}
		if i == 2 {
			jt = "hacker"
		}
		_, err := agents.Create(context.Background(), (repo.NewAgentInput)(sid, jt, 100, 100, 50))
		require.NoError(t, err)
	}
	m := service.NewMetricsService(db, agents, repo.NewTxRepo(db))
	snap, err := m.Collect(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 3, snap.AgentCount)
	assert.Equal(t, int64(300), snap.TotalGold)
	assert.Equal(t, 1, snap.JobsBreakdown["miner"])
	assert.Equal(t, 1, snap.JobsBreakdown["merchant"])
	assert.Equal(t, 1, snap.JobsBreakdown["hacker"])
}
