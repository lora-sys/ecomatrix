// Command seed populates the database with deterministic test agents.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ecomatrix/backend/internal/config"
	"github.com/ecomatrix/backend/internal/domain"
	"github.com/ecomatrix/backend/internal/observability"
	"github.com/ecomatrix/backend/internal/repo"
)

type seed struct {
	stringID string
	jobType  domain.JobType
	balance  int64
	vitality int
	credit   int
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := observability.NewLogger(cfg.LogLevel)

	db, err := repo.Open(cfg.DBDSN)
	if err != nil {
		log.Error("open db", "err", err)
		os.Exit(1)
	}
	if err := repo.Migrate(db); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}

	seeds := []seed{
		// 5 miners
		{"agent_miner_01", domain.JobMiner, 100, 80, 60},
		{"agent_miner_02", domain.JobMiner, 100, 80, 60},
		{"agent_miner_03", domain.JobMiner, 100, 80, 60},
		{"agent_miner_04", domain.JobMiner, 100, 80, 60},
		{"agent_miner_05", domain.JobMiner, 100, 80, 60},
		// 3 merchants
		{"agent_merchant_01", domain.JobMerchant, 200, 100, 70},
		{"agent_merchant_02", domain.JobMerchant, 200, 100, 70},
		{"agent_merchant_03", domain.JobMerchant, 200, 100, 70},
		// 2 hackers
		{"agent_hacker_01", domain.JobHacker, 80, 60, 40},
		{"agent_hacker_02", domain.JobHacker, 80, 60, 40},
		// 1 mediator
		{"agent_mediator_01", domain.JobMediator, 300, 100, 90},
	}

	ctx := context.Background()
	for _, s := range seeds {
		// ON CONFLICT (string_id) DO UPDATE balances/vitals to seed values.
		err := db.WithContext(ctx).Exec(`
			INSERT INTO agents (string_id, job_type, balance, vitality, credit_score)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (string_id) DO UPDATE
			SET job_type = EXCLUDED.job_type,
			    balance = EXCLUDED.balance,
			    vitality = EXCLUDED.vitality,
			    credit_score = EXCLUDED.credit_score,
			    updated_at = now()
		`, s.stringID, string(s.jobType), s.balance, s.vitality, s.credit).Error
		if err != nil {
			log.Error("seed", "agent", s.stringID, "err", err)
			os.Exit(1)
		}
	}
	log.Info("seed complete", "agents", len(seeds))
}
