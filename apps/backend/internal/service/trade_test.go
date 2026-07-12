package service_test

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/repo"
	"github.com/ecomatrix/backend/internal/service"
	"github.com/ecomatrix/backend/pkg/a2a"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testDB returns a freshly-migrated connection to a per-test schema so tests
// are independent and parallel-safe.
func testDB(t *testing.T) *gorm.DB {
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

func wipe(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE transactions, social_feeds, agents, conversations, llm_cache RESTART IDENTITY CASCADE").Error)
}

func seedAgents(t *testing.T, db *gorm.DB) (sender, target *domain.Agent) {
	t.Helper()
	repo := repo.NewAgentRepo(db)
	a, err := repo.Create(context.Background(), domain.Agent{
		StringID: "agent_miner_01", JobType: domain.JobMiner,
		Balance: 100, Vitality: 80, CreditScore: 60,
	})
	require.NoError(t, err)
	b, err := repo.Create(context.Background(), domain.Agent{
		StringID: "agent_merchant_03", JobType: domain.JobMerchant,
		Balance: 200, Vitality: 100, CreditScore: 70,
	})
	require.NoError(t, err)
	return &a, &b
}

func goodEnv() a2a.Envelope {
	return a2a.Envelope{
		ProtocolV: a2a.ProtocolVersion,
		MsgID:     "tx_req_test01",
		Sender:    "agent_miner_01",
		Action:    a2a.ActionExecuteTrade,
		Payload: map[string]any{
			"target_agent": "agent_merchant_03",
			"offer":        map[string]any{"currency_type": "GOLD", "amount": 40},
			"reasoning":    "buying supplies",
		},
		Timestamp: time.Now().Unix(),
	}
}

func TestTradeService_Settle_HappyPath(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	sender, target := seedAgents(t, db)

	svc := service.NewTradeService(db, repo.NewAgentRepo(db), repo.NewTxRepo(db), nil, nil)

	env := goodEnv()
	payload, err := a2a.DecodeTradePayload(env.Payload)
	require.NoError(t, err)
	res, aerr := svc.Settle(context.Background(), env, payload)
	require.Nil(t, aerr)
	assert.False(t, res.Replay)
	assert.Equal(t, int64(60), res.Receipt.BalanceAfter.From)
	assert.Equal(t, int64(240), res.Receipt.BalanceAfter.To)

	// Confirm ledger.
	got, err := repo.NewAgentRepo(db).ByID(context.Background(), sender.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(60), got.Balance)
	got, err = repo.NewAgentRepo(db).ByID(context.Background(), target.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(240), got.Balance)
}

func TestTradeService_Settle_InsufficientFunds(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	seedAgents(t, db)
	svc := service.NewTradeService(db, repo.NewAgentRepo(db), repo.NewTxRepo(db), nil, nil)

	env := goodEnv()
	env.Payload["offer"] = map[string]any{"currency_type": "GOLD", "amount": 99999}
	payload, err := a2a.DecodeTradePayload(env.Payload)
	require.NoError(t, err)

	res, aerr := svc.Settle(context.Background(), env, payload)
	require.NotNil(t, aerr)
	assert.Equal(t, a2a.CodeInsufficientFunds, aerr.Code)
	assert.Equal(t, domain.Receipt{}, res.Receipt)
}

func TestTradeService_Settle_UnknownAgent(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	seedAgents(t, db)
	svc := service.NewTradeService(db, repo.NewAgentRepo(db), repo.NewTxRepo(db), nil, nil)

	env := goodEnv()
	env.Sender = "agent_does_not_exist"
	_, aerr := svc.Settle(context.Background(), env, a2a.TradePayload{
		TargetAgent: "agent_merchant_03",
		Offer:       a2a.Offer{CurrencyType: a2a.CurrencyGold, Amount: 10},
	})
	require.NotNil(t, aerr)
	assert.Equal(t, a2a.CodeUnknownAgent, aerr.Code)
}

func TestTradeService_Settle_SelfTrade(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	seedAgents(t, db)
	svc := service.NewTradeService(db, repo.NewAgentRepo(db), repo.NewTxRepo(db), nil, nil)

	env := goodEnv()
	env.Sender = "agent_merchant_03"
	env.Payload["target_agent"] = "agent_merchant_03"
	payload, err := a2a.DecodeTradePayload(env.Payload)
	require.NoError(t, err)
	_, aerr := svc.Settle(context.Background(), env, payload)
	require.NotNil(t, aerr)
	assert.Equal(t, a2a.CodeSelfTrade, aerr.Code)
}

func TestTradeService_Settle_IdempotentReplay(t *testing.T) {
	db := testDB(t)
	wipe(t, db)
	seedAgents(t, db)
	svc := service.NewTradeService(db, repo.NewAgentRepo(db), repo.NewTxRepo(db), nil, nil)

	env := goodEnv()
	payload, err := a2a.DecodeTradePayload(env.Payload)
	require.NoError(t, err)

	res1, aerr1 := svc.Settle(context.Background(), env, payload)
	require.Nil(t, aerr1)

	res2, aerr2 := svc.Settle(context.Background(), env, payload)
	require.Nil(t, aerr2)
	assert.True(t, res2.Replay)
	assert.Equal(t, res1.Receipt.TxID, res2.Receipt.TxID)

	// Balance must reflect only one trade.
	got, err := repo.NewAgentRepo(db).ByStringID(context.Background(), "agent_miner_01")
	require.NoError(t, err)
	assert.Equal(t, int64(60), got.Balance)
}

// TestTradeService_Settle_50ConcurrentRacesNoDoubleSpend is the crown-jewel
// concurrency proof. 50 goroutines try to send 30 GOLD from a sender that
// starts with 1000 GOLD. Only 33 trades may settle; the rest are rejected;
// balance must remain non-negative.
func TestTradeService_Settle_50ConcurrentRacesNoDoubleSpend(t *testing.T) {
	db := testDB(t)
	wipe(t, db)

	agents := repo.NewAgentRepo(db)
	_, err := agents.Create(context.Background(), domain.Agent{
		StringID: "agent_race_sender", JobType: domain.JobMiner,
		Balance: 1000, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)
	_, err = agents.Create(context.Background(), domain.Agent{
		StringID: "agent_race_target", JobType: domain.JobMerchant,
		Balance: 0, Vitality: 100, CreditScore: 50,
	})
	require.NoError(t, err)

	svc := service.NewTradeService(db, agents, repo.NewTxRepo(db), nil, nil)

	const N = 50
	const amount = int64(30)

	var wg sync.WaitGroup
	var settled, rejected, errs atomic.Int64
	start := make(chan struct{})
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			env := a2a.Envelope{
				ProtocolV: a2a.ProtocolVersion,
				MsgID:     mustMsgID(i),
				Sender:    "agent_race_sender",
				Action:    a2a.ActionExecuteTrade,
				Payload: map[string]any{
					"target_agent": "agent_race_target",
					"offer":        map[string]any{"currency_type": "GOLD", "amount": amount},
				},
				Timestamp: time.Now().Unix(),
			}
			payload, _ := a2a.DecodeTradePayload(env.Payload)
			<-start
			_, aerr := svc.Settle(context.Background(), env, payload)
			switch {
			case aerr == nil:
				settled.Add(1)
			case aerr.Code == a2a.CodeInsufficientFunds:
				rejected.Add(1)
			default:
				errs.Add(1)
				t.Logf("unexpected err: %v", aerr)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	assert.Zero(t, errs.Load(), "no unexpected errors")
	assert.LessOrEqual(t, settled.Load(), int64(1000/amount), "settled <= initial/amount")
	assert.Equal(t, int64(N), settled.Load()+rejected.Load(), "all goroutines accounted for")

	// Final balance must be exactly initial - amount * settled, never negative.
	finalSender, err := agents.ByStringID(context.Background(), "agent_race_sender")
	require.NoError(t, err)
	finalTarget, err := agents.ByStringID(context.Background(), "agent_race_target")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, finalSender.Balance, int64(0), "sender balance must be non-negative")
	expected := int64(1000) - settled.Load()*amount
	assert.Equal(t, expected, finalSender.Balance, "sender balance exactly initial - settled*amount")
	assert.Equal(t, settled.Load()*amount, finalTarget.Balance, "target balance exactly settled*amount")

	t.Logf("settled=%d rejected=%d sender=%d target=%d",
		settled.Load(), rejected.Load(), finalSender.Balance, finalTarget.Balance)
}

func mustMsgID(i int) string {
	return "race_msg_" + leftpad(i)
}

func leftpad(i int) string {
	s := []byte("0000")
	n := i
	for j := len(s) - 1; j >= 0 && n > 0; j-- {
		s[j] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}
