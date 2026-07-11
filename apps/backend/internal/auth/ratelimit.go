// Package auth: simple per-agent token-bucket rate limiter.
//
// One bucket per (agent_id, action). Refills at a fixed rate. When a bucket
// is empty, the request is rejected with HTTP 429. The state is in-memory
// and resets on server restart; this is fine for MVP and matches the
// "trusted operator" threat model. For multi-instance deployments, swap
// the storage for Redis.
package auth

import (
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter caps requests per agent. capacity = max burst; refillPerSec
// is the steady-state allowance.
type RateLimiter struct {
	mu            sync.Mutex
	buckets       map[string]*bucket
	capacity      float64
	refillPerSec  float64
}

func NewRateLimiter(capacity, refillPerSec float64) *RateLimiter {
	return &RateLimiter{
		buckets:      map[string]*bucket{},
		capacity:     capacity,
		refillPerSec: refillPerSec,
	}
}

// Allow returns true if a token is available for (key), and consumes one
// if so. False otherwise.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, ok := r.buckets[key]
	if !ok {
		b = &bucket{tokens: r.capacity, lastRefill: now}
		r.buckets[key] = b
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(r.capacity, b.tokens+elapsed*r.refillPerSec)
	b.lastRefill = now
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
