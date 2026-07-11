package domain

import "time"

type JobType string

const (
	JobMiner    JobType = "miner"
	JobMerchant JobType = "merchant"
	JobHacker   JobType = "hacker"
	JobMediator JobType = "mediator"
)

func IsValidJobType(j string) bool {
	switch JobType(j) {
	case JobMiner, JobMerchant, JobHacker, JobMediator:
		return true
	}
	return false
}

// Agent is the world's citizen.
type Agent struct {
	ID            int64
	StringID      string
	JobType       JobType
	Balance       int64
	Vitality      int
	CreditScore   int
	LongTermMemory LongTermMemory `json:"long_term_memory"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LongTermMemory is the structured per-agent memory. Free-form in MVP,
// constrained by contract: at most 50 facts, summary <= 500 chars.
type LongTermMemory struct {
	Summary string   `json:"summary"`
	Facts   []string `json:"facts"`
}
