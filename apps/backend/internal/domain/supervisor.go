package domain

import "time"

// SupervisorRun is one execution of the hierarchical supervisor.
//
// JSONB columns (subtasks, workers, warnings) are handled by the
// repository layer using encoding/json; the domain only carries the
// in-memory shape so we avoid pulling pgtype into other packages.
type SupervisorRun struct {
	ID            int64
	Goal          string
	Status        string // "started" | "finished"
	Error         string
	Warnings      []string
	Subtasks      []map[string]any
	WorkerResults []map[string]any
	FinalSummary  string
	TokensUsed    int
	TokensBudget  int
	StartedAt     time.Time
	FinishedAt    *time.Time
	DurationMs    int
}
