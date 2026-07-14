package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func domainAgent(sid string, jt domain.JobType, balance int64, vit, credit int) domain.Agent {
	return domain.Agent{StringID: sid, JobType: jt, Balance: balance, Vitality: vit, CreditScore: credit}
}

func TestMetricsService_History_GrowsAndCaps(t *testing.T) {
	db := testDB(t)
	wipe(t, db)

	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	m := service.NewMetricsService(db, agents, txs)

	assert.Empty(t, m.History())

	for i := 0; i < 5; i++ {
		_, err := m.Collect(context.Background(), 0)
		require.NoError(t, err)
	}
	assert.Len(t, m.History(), 5)
}

func TestMetricsService_TradeCountInWindow(t *testing.T) {
	db := testDB(t)
	wipe(t, db)

	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	_, err := agents.Create(context.Background(), domainAgent("hist_sender", domain.JobMerchant, 100, 100, 50))
	require.NoError(t, err)
	_, err = agents.Create(context.Background(), domainAgent("hist_target", domain.JobMiner, 100, 100, 50))
	require.NoError(t, err)

	m := service.NewMetricsService(db, agents, txs)

	n, err := m.TradeCountInWindow(context.Background(), time.Now().Add(-60*time.Second))
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestMetricsService_Snapshot_IncludesSupervisorFields(t *testing.T) {
	db := testDB(t)
	wipe(t, db)

	agents := repo.NewAgentRepo(db)
	txs := repo.NewTxRepo(db)
	supervisor := repo.NewSupervisorRunsRepo(db)
	m := service.NewMetricsService(db, agents, txs)

	// Empty state.
	snap, err := m.Collect(context.Background(), 0)
	require.NoError(t, err)
	assert.EqualValues(t, 0, snap.SupervisorRunsCount)
	assert.Empty(t, snap.SupervisorLastRunAt)

	// Insert one finished run.
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, err = supervisor.Insert(context.Background(), domain.SupervisorRun{
		Goal:          "trade",
		Status:        "finished",
		Warnings:      []string{},
		Subtasks:      []map[string]any{{"subtask": "trade", "target_agent": "agent_miner_01"}},
		WorkerResults: []map[string]any{{"agent_id": "agent_miner_01"}},
		FinalSummary:  "ok", TokensUsed: 10, TokensBudget: 50, StartedAt: now,
		FinishedAt: func() *time.Time { t := now.Add(1500 * time.Millisecond); return &t }(),
		DurationMs: 1500,
	})
	require.NoError(t, err)

	snap, err = m.Collect(context.Background(), 0)
	require.NoError(t, err)
	assert.EqualValues(t, 1, snap.SupervisorRunsCount)
	require.NotEmpty(t, snap.SupervisorLastRunAt)
	parsed, perr := time.Parse(time.RFC3339Nano, snap.SupervisorLastRunAt)
	require.NoError(t, perr)
	assert.WithinDuration(t, now.Add(1500*time.Millisecond), parsed, 2*time.Second)
}
