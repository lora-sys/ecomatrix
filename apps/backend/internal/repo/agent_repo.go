package repo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

// AgentRepo reads and writes Agent rows.
type AgentRepo struct{ db *gorm.DB }

func NewAgentRepo(db *gorm.DB) *AgentRepo { return &AgentRepo{db: db} }

func (r *AgentRepo) List(ctx context.Context, limit, offset int) ([]domain.Agent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []AgentModel
	if err := r.db.WithContext(ctx).Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Agent, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainAgent(r))
	}
	return out, nil
}

func (r *AgentRepo) ByID(ctx context.Context, id int64) (domain.Agent, error) {
	var row AgentModel
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Agent{}, domain.ErrAgentNotFound
		}
		return domain.Agent{}, err
	}
	return toDomainAgent(row), nil
}

func (r *AgentRepo) ByStringID(ctx context.Context, sid string) (domain.Agent, error) {
	var row AgentModel
	if err := r.db.WithContext(ctx).Where("string_id = ?", sid).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Agent{}, domain.ErrAgentNotFound
		}
		return domain.Agent{}, err
	}
	return toDomainAgent(row), nil
}

func (r *AgentRepo) Create(ctx context.Context, a domain.Agent) (domain.Agent, error) {
	row := AgentModel{
		StringID:    a.StringID,
		JobType:     string(a.JobType),
		Balance:     a.Balance,
		Vitality:    a.Vitality,
		CreditScore: a.CreditScore,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Agent{}, err
	}
	return toDomainAgent(row), nil
}

// LockPair takes a row-level FOR UPDATE lock on the two agents with the given
// ids, always in ascending id order to avoid deadlock. Must be called inside a
// transaction. Uses raw SQL because GORM's gorm:query_option is unreliable for
// row locking. Returns ErrAgentNotFound if either row is missing.
func (r *AgentRepo) LockPair(ctx context.Context, tx *gorm.DB, a, b int64) (domain.Agent, domain.Agent, error) {
	first, second := a, b
	if first > second {
		first, second = second, first
	}
	var rows []AgentModel
	if err := tx.WithContext(ctx).
		Raw(`SELECT * FROM agents WHERE id IN (?, ?) ORDER BY id ASC FOR UPDATE`, first, second).
		Scan(&rows).Error; err != nil {
		return domain.Agent{}, domain.Agent{}, err
	}
	if len(rows) != 2 {
		return domain.Agent{}, domain.Agent{}, domain.ErrAgentNotFound
	}
	lookup := map[int64]AgentModel{}
	for _, row := range rows {
		lookup[row.ID] = row
	}
	return toDomainAgent(lookup[a]), toDomainAgent(lookup[b]), nil
}

// ApplyDelta updates balances inside a tx. Returns ErrAgentNotFound if the row
// is missing; the database CHECK constraint on balance >= 0 surfaces as the
// underlying pg error and is mapped to ErrInsufficientFunds by the caller.
func (r *AgentRepo) ApplyDelta(ctx context.Context, tx *gorm.DB, id int64, delta int64) error {
	res := tx.WithContext(ctx).
		Exec(`UPDATE agents SET balance = balance + ?, updated_at = now() WHERE id = ?`, delta, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrAgentNotFound
	}
	return nil
}

// ErrAgentNotFoundSentinel is re-exported so the transport layer can match it
// without importing the domain package directly.
var ErrAgentNotFoundSentinel = func() error { return domain.ErrAgentNotFound }

// NewAgentInput packages the create-agent payload.
func NewAgentInput(stringID, jobType string, balance int64, vitality, credit int) domain.Agent {
	return domain.Agent{
		StringID:    stringID,
		JobType:     domain.JobType(jobType),
		Balance:     balance,
		Vitality:    vitality,
		CreditScore: credit,
	}
}

// GetLongTermMemory returns the LTM JSONB column for one agent.
// GORM's raw Scan into []byte misinterprets JSONB; we Scan into a string and
// then unmarshal.
func (r *AgentRepo) GetLongTermMemory(ctx context.Context, id int64) (domain.LongTermMemory, error) {
	var raw string
	if err := r.db.WithContext(ctx).
		Raw(`SELECT long_term_memory::text FROM agents WHERE id = ?`, id).
		Scan(&raw).Error; err != nil {
		return domain.LongTermMemory{}, err
	}
	ltm := domain.LongTermMemory{}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &ltm)
	}
	return ltm, nil
}

// SetLongTermMemory overwrites the LTM JSONB column for one agent.
func (r *AgentRepo) SetLongTermMemory(ctx context.Context, id int64, ltm domain.LongTermMemory) error {
	data, err := json.Marshal(ltm)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Exec(`UPDATE agents SET long_term_memory = ?, updated_at = now() WHERE id = ?`, data, id).Error
}
