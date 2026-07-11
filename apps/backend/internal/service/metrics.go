package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ecomatrix/backend/internal/repo"
	"gorm.io/gorm"
)

// MetricsService computes the dashboard-facing aggregates. It also tracks
// real-time counters (recent QPS, last trade timestamp) in memory because
// those are not derivable from a single SQL query without sampling work.
type MetricsService struct {
	db     *gorm.DB
	agents *repo.AgentRepo
	txs    *repo.TxRepo

	mu            sync.RWMutex
	recentCounts  []int64 // 1-second buckets for the last 60 seconds
	lastBucketSec int64
	lastTradeAtNS atomic.Int64 // unix nano of the last settled trade; 0 if none
}

func NewMetricsService(db *gorm.DB, agents *repo.AgentRepo, txs *repo.TxRepo) *MetricsService {
	return &MetricsService{db: db, agents: agents, txs: txs}
}

// Snapshot is the dashboard-facing aggregate. Stable JSON shape; do not
// rename fields without bumping the A2A-style version.
type Snapshot struct {
	AgentCount    int            `json:"agent_count"`
	TotalGold     int64          `json:"total_gold"`
	JobsBreakdown map[string]int `json:"jobs_breakdown"`
	RecentQPS     float64        `json:"recent_qps"`
	WSConnections int            `json:"ws_connections"`
	LastTradeAt   string         `json:"last_trade_at,omitempty"`
	GeneratedAt   string         `json:"generated_at"`
}

// NoteTrade increments the rolling counter and updates last-trade timestamp.
// Called from the trade service after a settled trade.
func (m *MetricsService) NoteTrade() {
	now := time.Now()
	m.lastTradeAtNS.Store(now.UnixNano())

	m.mu.Lock()
	defer m.mu.Unlock()
	sec := now.Unix()
	if sec != m.lastBucketSec {
		// Shift left so the newest bucket is at the end.
		if m.lastBucketSec != 0 && sec-m.lastBucketSec >= int64(len(m.recentCounts)) {
			m.recentCounts = m.recentCounts[:0]
		} else if sec-m.lastBucketSec > 1 {
			drop := int(sec - m.lastBucketSec - 1)
			if drop >= len(m.recentCounts) {
				m.recentCounts = m.recentCounts[:0]
			} else {
				m.recentCounts = m.recentCounts[drop:]
			}
		}
		m.lastBucketSec = sec
		m.recentCounts = append(m.recentCounts, 0)
	}
	if len(m.recentCounts) == 0 {
		m.recentCounts = append(m.recentCounts, 0)
	}
	m.recentCounts[len(m.recentCounts)-1]++
}

// SetWSConnections is called by the HTTP layer when conn count changes.
func (m *MetricsService) SetWSConnections(n int) {
	// Stored as a metric; cheap to compute from the hub, but we cache for the
	// snapshot to avoid extra contention.
	m.mu.Lock()
	// Inject into the most recent bucket's metadata? Easier: just keep a side
	// field. For MVP, the snapshot reads from the hub directly.
	m.mu.Unlock()
}

// Collect produces a Snapshot for the given ws connection count.
func (m *MetricsService) Collect(ctx context.Context, wsConns int) (Snapshot, error) {
	agents, err := m.agents.List(ctx, 200, 0)
	if err != nil {
		return Snapshot{}, err
	}
	jobs := map[string]int{}
	var total int64
	for _, a := range agents {
		jobs[string(a.JobType)]++
		total += a.Balance
	}

	m.mu.RLock()
	// Sum counts in the last 10 seconds for a reasonable "recent" QPS.
	window := 10
	if len(m.recentCounts) < window {
		window = len(m.recentCounts)
	}
	var sum int64
	for i := len(m.recentCounts) - window; i < len(m.recentCounts); i++ {
		if i < 0 {
			continue
		}
		sum += m.recentCounts[i]
	}
	var qps float64
	if window > 0 {
		qps = float64(sum) / float64(window)
	}
	lastNS := m.lastTradeAtNS.Load()
	m.mu.RUnlock()

	snap := Snapshot{
		AgentCount:    len(agents),
		TotalGold:     total,
		JobsBreakdown: jobs,
		RecentQPS:     qps,
		WSConnections: wsConns,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}
	if lastNS > 0 {
		snap.LastTradeAt = time.Unix(0, lastNS).UTC().Format(time.RFC3339Nano)
	}
	return snap, nil
}
