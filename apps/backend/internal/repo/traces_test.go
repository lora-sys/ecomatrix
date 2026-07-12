package repo_test

import (
	"context"
	"testing"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracesRepo_Roundtrip(t *testing.T) {
	db := testDBRepo(t)
	wipeRepo(t, db)
	agents := repo.NewAgentRepo(db)
	_, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_trace_test", JobType: domain.JobMiner,
		Balance: 100, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)

	r := repo.NewTracesRepo(db)
	inserted, err := r.Insert(context.Background(), domain.Trace{
		AgentID:   "agent_trace_test",
		Kind:      "plan",
		Content:   "1. check balance 2. trade 3. report",
		LatencyMS: intPtr(15),
		TokensIn:  intPtr(120),
		TokensOut: intPtr(80),
	})
	require.NoError(t, err)
	assert.NotZero(t, inserted.ID)

	recent, err := r.Recent(context.Background(), "agent_trace_test", 10)
	require.NoError(t, err)
	require.NotEmpty(t, recent)
	assert.Equal(t, "plan", recent[0].Kind)
	assert.Equal(t, "1. check balance 2. trade 3. report", recent[0].Content)
	assert.NotNil(t, recent[0].LatencyMS)
	assert.Equal(t, 15, *recent[0].LatencyMS)
}

func intPtr(i int) *int { return &i }
