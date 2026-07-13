package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ecomatrix/backend/internal/domain"
	"gorm.io/gorm"
)

type SupervisorRunModel struct {
	ID           int64      `gorm:"primaryKey;column:id"`
	Goal         string     `gorm:"column:goal"`
	Status       string     `gorm:"column:status"`
	Error        string     `gorm:"column:error"`
	WarningsJSON []byte     `gorm:"column:warnings;type:jsonb"`
	SubtasksJSON []byte     `gorm:"column:subtasks;type:jsonb"`
	WorkersJSON  []byte     `gorm:"column:worker_results;type:jsonb"`
	FinalSummary string     `gorm:"column:final_summary"`
	TokensUsed   int        `gorm:"column:tokens_used"`
	TokensBudget int        `gorm:"column:tokens_budget"`
	StartedAt    time.Time  `gorm:"column:started_at"`
	FinishedAt   *time.Time `gorm:"column:finished_at"`
	DurationMs   int        `gorm:"column:duration_ms"`
}

func (SupervisorRunModel) TableName() string { return "supervisor_runs" }

type SupervisorRunsRepo struct{ db *gorm.DB }

func NewSupervisorRunsRepo(db *gorm.DB) *SupervisorRunsRepo {
	return &SupervisorRunsRepo{db: db}
}

func (r *SupervisorRunsRepo) Insert(ctx context.Context, in domain.SupervisorRun) (domain.SupervisorRun, error) {
	warnings, err := json.Marshal(in.Warnings)
	if err != nil {
		return domain.SupervisorRun{}, err
	}
	subtasks, err := json.Marshal(in.Subtasks)
	if err != nil {
		return domain.SupervisorRun{}, err
	}
	workers, err := json.Marshal(in.WorkerResults)
	if err != nil {
		return domain.SupervisorRun{}, err
	}
	row := SupervisorRunModel{
		Goal:         in.Goal,
		Status:       in.Status,
		Error:        in.Error,
		WarningsJSON: warnings,
		SubtasksJSON: subtasks,
		WorkersJSON:  workers,
		FinalSummary: in.FinalSummary,
		TokensUsed:   in.TokensUsed,
		TokensBudget: in.TokensBudget,
		StartedAt:    in.StartedAt,
		FinishedAt:   in.FinishedAt,
		DurationMs:   in.DurationMs,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.SupervisorRun{}, err
	}
	in.ID = row.ID
	return in, nil
}

func (r *SupervisorRunsRepo) Recent(ctx context.Context, limit int) ([]domain.SupervisorRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []SupervisorRunModel
	if err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.SupervisorRun, 0, len(rows))
	for _, row := range rows {
		run := domain.SupervisorRun{
			ID:           row.ID,
			Goal:         row.Goal,
			Status:       row.Status,
			Error:        row.Error,
			FinalSummary: row.FinalSummary,
			TokensUsed:   row.TokensUsed,
			TokensBudget: row.TokensBudget,
			StartedAt:    row.StartedAt,
			FinishedAt:   row.FinishedAt,
			DurationMs:   row.DurationMs,
		}
		if len(row.WarningsJSON) > 0 {
			_ = json.Unmarshal(row.WarningsJSON, &run.Warnings)
		}
		if len(row.SubtasksJSON) > 0 {
			_ = json.Unmarshal(row.SubtasksJSON, &run.Subtasks)
		}
		if len(row.WorkersJSON) > 0 {
			_ = json.Unmarshal(row.WorkersJSON, &run.WorkerResults)
		}
		out = append(out, run)
	}
	return out, nil
}
