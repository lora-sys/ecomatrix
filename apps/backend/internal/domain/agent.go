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
	ID          int64
	StringID    string
	JobType     JobType
	Balance     int64
	Vitality    int
	CreditScore int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
