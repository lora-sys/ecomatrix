// Package repo contains data access. Repositories take a *gorm.DB and never
// the global connection. Services are expected to wrap state-mutating calls
// in a transaction (see internal/service/trade.go).
package repo

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations_fs/*.sql
var migrationsFS embed.FS

// Open returns a tuned *gorm.DB ready for the service layer.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

// AgentModel is the GORM projection of agents.
type AgentModel struct {
	ID             int64     `gorm:"primaryKey"`
	StringID       string    `gorm:"column:string_id;uniqueIndex"`
	JobType        string    `gorm:"column:job_type"`
	Balance        int64     `gorm:"column:balance"`
	Vitality       int       `gorm:"column:vitality"`
	CreditScore    int       `gorm:"column:credit_score"`
	LongTermMemory []byte    `gorm:"column:long_term_memory;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (AgentModel) TableName() string { return "agents" }

func toDomainAgent(a AgentModel) domain.Agent {
	ltm := domain.LongTermMemory{}
	if len(a.LongTermMemory) > 0 {
		_ = json.Unmarshal(a.LongTermMemory, &ltm)
	}
	return domain.Agent{
		ID:             a.ID,
		StringID:       a.StringID,
		JobType:        domain.JobType(a.JobType),
		Balance:        a.Balance,
		Vitality:       a.Vitality,
		CreditScore:    a.CreditScore,
		LongTermMemory: ltm,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

// TxModel is the GORM projection of transactions.
type TxModel struct {
	ID           int64     `gorm:"primaryKey"`
	TxID         string    `gorm:"column:tx_id;uniqueIndex"`
	MsgID        string    `gorm:"column:msg_id;uniqueIndex"`
	FromID       int64     `gorm:"column:from_id"`
	ToID         int64     `gorm:"column:to_id"`
	Amount       int64     `gorm:"column:amount"`
	CurrencyType string    `gorm:"column:currency_type"`
	Status       string    `gorm:"column:status"`
	ErrorCode    string    `gorm:"column:error_code"`
	Reasoning    string    `gorm:"column:reasoning"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (TxModel) TableName() string { return "transactions" }

// Migrate applies the embedded migrations in order. Tracks state in
// schema_migrations(version INT PK, applied_at TIMESTAMPTZ).
func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations_fs")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type mig struct {
		version int
		upSQL   string
		downSQL string
	}
	ups := map[int]string{}
	downs := map[int]string{}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			n := strings.TrimSuffix(name, ".up.sql")
			v, err := strconv.Atoi(strings.SplitN(n, "_", 2)[0])
			if err != nil {
				return fmt.Errorf("bad migration name %q: %w", name, err)
			}
			b, err := migrationsFS.ReadFile("migrations_fs/" + name)
			if err != nil {
				return err
			}
			ups[v] = string(b)
		case strings.HasSuffix(name, ".down.sql"):
			n := strings.TrimSuffix(name, ".down.sql")
			v, err := strconv.Atoi(strings.SplitN(n, "_", 2)[0])
			if err != nil {
				return fmt.Errorf("bad migration name %q: %w", name, err)
			}
			b, err := migrationsFS.ReadFile("migrations_fs/" + name)
			if err != nil {
				return err
			}
			downs[v] = string(b)
		}
	}

	versions := make([]int, 0, len(ups))
	for v := range ups {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	for _, v := range versions {
		var existing int64
		err := db.Raw(`SELECT version FROM schema_migrations WHERE version = ?`, v).Scan(&existing).Error
		if err != nil {
			return fmt.Errorf("check migration %d: %w", v, err)
		}
		if existing == int64(v) {
			continue
		}
		if err := db.Exec(ups[v]).Error; err != nil {
			return fmt.Errorf("apply migration %d: %w", v, err)
		}
		if err := db.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, v).Error; err != nil {
			return fmt.Errorf("record migration %d: %w", v, err)
		}
	}
	return nil
}

// MigrateDown reverses the latest applied migration. Intended for tests.
func MigrateDown(db *gorm.DB) error {
	entries, err := migrationsFS.ReadDir("migrations_fs")
	if err != nil {
		return err
	}
	downs := map[int]string{}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".down.sql") {
			continue
		}
		n := strings.TrimSuffix(e.Name(), ".down.sql")
		v, err := strconv.Atoi(strings.SplitN(n, "_", 2)[0])
		if err != nil {
			return err
		}
		b, err := migrationsFS.ReadFile("migrations_fs/" + e.Name())
		if err != nil {
			return err
		}
		downs[v] = string(b)
	}
	var latest int64
	if err := db.Raw(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&latest).Error; err != nil {
		return err
	}
	if latest == 0 {
		return nil
	}
	if err := db.Exec(downs[int(latest)]).Error; err != nil {
		return err
	}
	return db.Exec(`DELETE FROM schema_migrations WHERE version = ?`, latest).Error
}
