package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_BurstThenRefill(t *testing.T) {
	r := NewRateLimiter(3, 1) // burst 3, refill 1/sec
	// 3 immediate requests should pass.
	assert.True(t, r.Allow("a"))
	assert.True(t, r.Allow("a"))
	assert.True(t, r.Allow("a"))
	// 4th should fail.
	assert.False(t, r.Allow("a"))
	// Wait 1.1s — bucket refills to 1.
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, r.Allow("a"))
	// Next should fail again.
	assert.False(t, r.Allow("a"))
}

func TestRateLimiter_PerKeyIsolation(t *testing.T) {
	r := NewRateLimiter(1, 0.001)
	assert.True(t, r.Allow("a"))
	assert.False(t, r.Allow("a"))
	// Different key — fresh bucket.
	assert.True(t, r.Allow("b"))
}

func TestRateLimiter_RecoversAfterIdle(t *testing.T) {
	r := NewRateLimiter(2, 100) // high refill
	assert.True(t, r.Allow("a"))
	assert.True(t, r.Allow("a"))
	assert.False(t, r.Allow("a"))
	time.Sleep(100 * time.Millisecond) // 0.1s * 100/sec = 10 tokens
	assert.True(t, r.Allow("a"))
	assert.True(t, r.Allow("a"))
}
