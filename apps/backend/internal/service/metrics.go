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
	// Count settled trades in the last 60s and record the snapshot.
	since := time.Now().Add(-60 * time.Second)
	tc, _ := m.TradeCountInWindow(ctx, since)
	recordHistorySample(snap, tc)
	return snap, nil
}

// HistorySample is one snapshot of the world's metrics at a given moment.
type HistorySample struct {
	At         time.Time `json:"at"`
	AgentCount int       `json:"agent_count"`
	TotalGold  int64     `json:"total_gold"`
	RecentQPS  float64   `json:"recent_qps"`
	TradeCount int       `json:"trade_count"` // count of settled trades in this sample window
}

// History keeps a ring buffer of the last `HistoryCapacity` snapshots.
const HistoryCapacity = 120 // ~ 2 minutes at 1s cadence

// historyMu guards the history ring. Sample cadence is 1s.
var (
	historyMu sync.Mutex
	history   = make([]HistorySample, 0, HistoryCapacity)
	historyCh = make(chan struct{}, 1)
)

// History returns a copy of the ring buffer in chronological order (oldest
// first). The slice is always <= HistoryCapacity in length.
func (m *MetricsService) History() []HistorySample {
	historyMu.Lock()
	defer historyMu.Unlock()
	out := make([]HistorySample, len(history))
	copy(out, history)
	return out
}

// recordHistorySample is called from the main Collect path; it appends a
// new sample to the ring buffer.
func recordHistorySample(snap Snapshot, tradeCount int) {
	historyMu.Lock()
	defer historyMu.Unlock()
	sample := HistorySample{
		At:         time.Now().UTC(),
		AgentCount: snap.AgentCount,
		TotalGold:  snap.TotalGold,
		RecentQPS:  snap.RecentQPS,
		TradeCount: tradeCount,
	}
	if len(history) >= HistoryCapacity {
		// Drop the oldest.
		history = history[1:]
	}
	history = append(history, sample)
	// Non-blocking notify (so a test could await the next sample).
	select {
	case historyCh <- struct{}{}:
	default:
	}
}

// TradeCountInWindow counts settled trades whose created_at >= since.
func (m *MetricsService) TradeCountInWindow(ctx context.Context, since time.Time) (int, error) {
	var n int64
	err := m.db.WithContext(ctx).
		Raw(`SELECT COUNT(*) FROM transactions WHERE status = 'SETTLED' AND created_at >= ?`, since).
		Scan(&n).Error
	return int(n), err
}

// StartHistoryTicker launches a goroutine that records a history sample
// every `interval`. Call cancel() to stop. Safe to call once at boot.
func (m *MetricsService) StartHistoryTicker(interval time.Duration) (cancel func()) {
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				ctx, c := context.WithTimeout(context.Background(), 2*time.Second)
				_, _ = m.Collect(ctx, 0) // records a sample as a side effect
				c()
			}
		}
	}()
	return func() { close(stop) }
}
