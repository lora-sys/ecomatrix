package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
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

// We isolate this test binary into its own schema so the http test pool
// doesn't compete with repo/ and service/ packages' tests on the shared
// `ecomatrix_test` database during `go test -race ./...`.
var supervisorTestDSN string
var supervisorSchemaName string

func TestMain(m *testing.M) {
	os.Exit(runSupervisorTests(m))
}

func runSupervisorTests(m *testing.M) (code int) {
	baseDSN := os.Getenv("ECOMATRIX_TEST_DSN")
	if baseDSN == "" {
		baseDSN = "postgres://repotwin:repotwin@localhost:5432/ecomatrix_test?sslmode=disable"
	}
	baseDB, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open supervisor test database:", err)
		return 1
	}
	supervisorSchemaName = fmt.Sprintf("ecomatrix_http_test_%d", os.Getpid())
	if err := baseDB.Exec("CREATE SCHEMA " + supervisorSchemaName).Error; err != nil {
		fmt.Fprintln(os.Stderr, "create supervisor test schema:", err)
		return 1
	}
	defer func() {
		if err := baseDB.Exec("DROP SCHEMA " + supervisorSchemaName + " CASCADE").Error; err != nil {
			fmt.Fprintln(os.Stderr, "drop supervisor test schema:", err)
			if code == 0 {
				code = 1
			}
		}
	}()
	parsed, err := url.Parse(baseDSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse supervisor test DSN:", err)
		return 1
	}
	q := parsed.Query()
	q.Set("search_path", supervisorSchemaName)
	parsed.RawQuery = q.Encode()
	supervisorTestDSN = parsed.String()

	testDB, err := gorm.Open(postgres.Open(supervisorTestDSN), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open isolated supervisor test schema:", err)
		return 1
	}
	if err := repo.Migrate(testDB); err != nil {
		fmt.Fprintln(os.Stderr, "migrate isolated supervisor test schema:", err)
		return 1
	}
	return m.Run()
}

func supervisorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(supervisorTestDSN), &gorm.Config{})
	require.NoError(t, err)
	// Cap this package's pool at 1 connection: the supervisor router tests
	// are independent and don't need concurrency. Sharing a single conn
	// also keeps the package from saturating CI's default Postgres
	// `max_connections=100` when the broader test suite runs in parallel.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
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
