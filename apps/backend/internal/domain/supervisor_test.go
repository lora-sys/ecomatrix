package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSupervisorRun_WarningsRoundTrip(t *testing.T) {
	run := SupervisorRun{
		Goal:     "trade",
		Status:   "finished",
		Warnings: []string{"aggregation returned an empty summary"},
		Subtasks: []map[string]any{
			{"subtask": "trade", "target_agent": "agent_miner_01"},
		},
		WorkerResults: []map[string]any{
			{"agent_id": "agent_miner_01", "error": nil},
		},
	}
	warnings, err := json.Marshal(run.Warnings)
	require.NoError(t, err)
	require.Contains(t, string(warnings), "aggregation returned an empty summary")

	subtasks, err := json.Marshal(run.Subtasks)
	require.NoError(t, err)
	require.Contains(t, string(subtasks), "agent_miner_01")
}
