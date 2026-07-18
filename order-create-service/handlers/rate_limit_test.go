package handlers

import (
	"math"
	"testing"
)

// Pure-logic simulation of the Lua refill math (no Redis).
func refill(tokens, capacity, rate float64, lastMs, nowMs int64) float64 {
	delta := float64(nowMs-lastMs) / 1000.0
	if delta < 0 {
		delta = 0
	}
	t := tokens + delta*rate
	if t > capacity {
		t = capacity
	}
	return t
}

func TestTokenBucketRefillMath(t *testing.T) {
	const cap, rate = 5.0, 1.0
	// empty after burst
	tokens := 0.0
	last := int64(1_000_000)
	// after 3s → 3 tokens
	tokens = refill(tokens, cap, rate, last, last+3000)
	if math.Abs(tokens-3) > 1e-9 {
		t.Fatalf("want 3 got %v", tokens)
	}
	// take 1
	tokens--
	// after 10s should cap at 5
	tokens = refill(tokens, cap, rate, last+3000, last+3000+10_000)
	if math.Abs(tokens-cap) > 1e-9 {
		t.Fatalf("want capacity %v got %v", cap, tokens)
	}
}

func TestTokenBucketBurstThenThrottle(t *testing.T) {
	const cap, rate, cost = 5.0, 1.0, 1.0
	tokens := cap
	var last int64 = 0
	now := int64(0)
	allowed := 0
	for i := 0; i < 10; i++ {
		tokens = refill(tokens, cap, rate, last, now)
		last = now
		if tokens >= cost {
			tokens -= cost
			allowed++
		}
		now += 10 // 10ms between tries — no meaningful refill
	}
	if allowed != 5 {
		t.Fatalf("burst allow want 5 got %d", allowed)
	}
}
