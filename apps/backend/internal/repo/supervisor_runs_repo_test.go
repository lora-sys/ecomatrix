package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSupervisorRun(goal string, workers []map[string]any, started time.Time, finished *time.Time) domain.SupervisorRun {
	run := domain.SupervisorRun{
		Goal:          goal,
		Status:        "finished",
		Warnings:      []string{"warn-a"},
		Subtasks:      []map[string]any{{"subtask": "trade", "target_agent": "agent_miner_01"}},
		WorkerResults: workers,
		FinalSummary:  "ok",
		TokensUsed:    25,
		TokensBudget:  100,
		StartedAt:     started,
		FinishedAt:    finished,
		DurationMs:    42,
	}
	if run.WorkerResults == nil {
		run.WorkerResults = []map[string]any{}
	}
	return run
}

func TestSupervisorRunsRepo_ByIDAndByAgent(t *testing.T) {
	db := testDBRepo(t)
	wipeRepo(t, db)
	r := repo.NewSupervisorRunsRepo(db)

	t0 := time.Now().UTC().Add(-5 * time.Minute).Truncate(time.Microsecond)
	fin := t0.Add(1500 * time.Millisecond)
	inserted, err := r.Insert(context.Background(), newSupervisorRun(
		"trade",
		[]map[string]any{
			{"agent_id": "agent_miner_01", "receipt": map[string]any{"tx_id": "tx_1", "amount": 12}},
			{"agent_id": "agent_merchant_03", "receipt": map[string]any{"tx_id": "tx_2", "amount": 7}},
		},
		t0, &fin,
	))
	require.NoError(t, err)
	require.NotZero(t, inserted.ID)

	// ByID happy path.
	got, err := r.ByID(context.Background(), inserted.ID)
	require.NoError(t, err)
	assert.Equal(t, inserted.ID, got.ID)
	assert.Equal(t, "trade", got.Goal)
	require.Len(t, got.WorkerResults, 2)
	require.Len(t, got.Subtasks, 1)
	require.Len(t, got.Warnings, 1)

	// ByID not-found -> ErrSupervisorRunNotFound.
	_, err = r.ByID(context.Background(), inserted.ID+999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrSupervisorRunNotFound))

	// ByAgent filter on worker_results JSONB.
	fromMiner, err := r.ByAgent(context.Background(), "agent_miner_01", 10)
	require.NoError(t, err)
	require.Len(t, fromMiner, 1)
	assert.Equal(t, inserted.ID, fromMiner[0].ID)

	fromGhost, err := r.ByAgent(context.Background(), "agent_does_not_exist", 10)
	require.NoError(t, err)
	assert.Empty(t, fromGhost)
}
