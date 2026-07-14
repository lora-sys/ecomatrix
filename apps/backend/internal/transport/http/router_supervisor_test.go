package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// supervisorTestDB opens one gorm connection per test and caps the pool to
// a single connection. Sharing a single connection serializes this package's
// tests against the global `max_connections=100` Postgres that CI provides
// (so this suite doesn't compound the 50-race trade_test's footprint).
// All four packages hit the same `ecomatrix_test` schema; the wipe runs
// first inside each test, just like the rest of the suite.
func supervisorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("ECOMATRIX_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://repotwin:repotwin@localhost:5432/ecomatrix_test?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, repo.Migrate(db))
	require.NoError(t, db.Exec("TRUNCATE transactions, social_feeds, agents, conversations, llm_cache, agent_secrets, supervisor_runs RESTART IDENTITY CASCADE").Error)
	return db
}

func newSupervisorTestServer(t *testing.T, db *gorm.DB) (*fiber.App, *Server) {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	agents := repo.NewAgentRepo(db)
	supervisor := repo.NewSupervisorRunsRepo(db)
	server := &Server{
		App: app, Agents: agents,
		Supervisor: supervisor,
	}
	app.Get("/v1/supervisor/runs/:id", server.getSupervisorRun)
	app.Get("/v1/agents/by-string-id/:sid/supervisor-runs", server.getAgentSupervisorRuns)
	return app, server
}

func TestRouter_GetSupervisorRun_OK(t *testing.T) {
	db := supervisorTestDB(t)
	app, _ := newSupervisorTestServer(t, db)
	agents := repo.NewAgentRepo(db)
	supervisor := repo.NewSupervisorRunsRepo(db)
	_, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_miner_01", JobType: domain.JobMiner,
		Balance: 100, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)
	t0 := time.Now().UTC().Truncate(time.Microsecond)
	fin := t0.Add(800 * time.Millisecond)
	inserted, err := supervisor.Insert(context.Background(), domain.SupervisorRun{
		Goal:     "trade",
		Status:   "finished",
		Warnings: []string{"aggregation warning"},
		Subtasks: []map[string]any{{"subtask": "trade", "target_agent": "agent_miner_01"}},
		WorkerResults: []map[string]any{
			{"agent_id": "agent_miner_01", "receipt": map[string]any{"tx_id": "tx_1", "amount": 5}},
		},
		FinalSummary: "trade done",
		TokensUsed:   12, TokensBudget: 50, StartedAt: t0, FinishedAt: &fin, DurationMs: 800,
	})
	require.NoError(t, err)
	url := fmt.Sprintf("/v1/supervisor/runs/%d", inserted.ID)
	resp, err := app.Test(httptest.NewRequest("GET", url, nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(body, &out))
	assert.EqualValues(t, inserted.ID, out["id"])
	assert.Equal(t, "trade", out["goal"])
	workers, ok := out["worker_results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, workers)
}

func TestRouter_GetSupervisorRun_NotFound(t *testing.T) {
	db := supervisorTestDB(t)
	app, _ := newSupervisorTestServer(t, db)
	resp, err := app.Test(httptest.NewRequest("GET", "/v1/supervisor/runs/99999", nil))
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestRouter_GetAgentSupervisorRuns_FiltersByAgent(t *testing.T) {
	db := supervisorTestDB(t)
	app, _ := newSupervisorTestServer(t, db)
	agents := repo.NewAgentRepo(db)
	supervisor := repo.NewSupervisorRunsRepo(db)
	_, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_miner_01", JobType: domain.JobMiner,
		Balance: 100, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)
	t0 := time.Now().UTC().Truncate(time.Microsecond)
	fin := t0.Add(500 * time.Millisecond)
	_, err = supervisor.Insert(context.Background(), domain.SupervisorRun{
		Goal: "miner-coordinated", Status: "finished",
		Warnings: []string{}, Subtasks: []map[string]any{},
		WorkerResults: []map[string]any{{"agent_id": "agent_miner_01"}},
		FinalSummary:  "ok", TokensUsed: 10, TokensBudget: 50, StartedAt: t0, FinishedAt: &fin, DurationMs: 500,
	})
	require.NoError(t, err)
	_, err = supervisor.Insert(context.Background(), domain.SupervisorRun{
		Goal: "merchant-coordinated", Status: "finished",
		Warnings: []string{}, Subtasks: []map[string]any{},
		WorkerResults: []map[string]any{{"agent_id": "agent_merchant_03"}},
		FinalSummary:  "ok", TokensUsed: 10, TokensBudget: 50, StartedAt: t0.Add(time.Second), FinishedAt: &fin, DurationMs: 500,
	})
	require.NoError(t, err)
	resp, err := app.Test(httptest.NewRequest("GET", "/v1/agents/by-string-id/agent_miner_01/supervisor-runs?limit=10", nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	payload := map[string]any{}
	require.NoError(t, json.Unmarshal(body, &payload))
	runs, ok := payload["runs"].([]any)
	require.True(t, ok)
	require.Len(t, runs, 1)
}

func TestRouter_GetAgentSupervisorRuns_UnknownAgentReturns404(t *testing.T) {
	db := supervisorTestDB(t)
	app, _ := newSupervisorTestServer(t, db)
	resp, err := app.Test(httptest.NewRequest("GET", "/v1/agents/by-string-id/agent_ghost/supervisor-runs", nil))
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}
