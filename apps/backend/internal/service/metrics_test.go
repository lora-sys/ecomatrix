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
