package handlers

import (
	"context"
	"fmt"
	"order-create/dao"
	"time"
)

func nowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// Token bucket defaults (per user).
// capacity: max burst; rate: tokens refilled per second.
const (
	tokenBucketCapacity = 5
	tokenBucketRate     = 1.0 // 1 token / s steady-state
	tokenBucketCost     = 1
	// Idle keys expire after this many seconds (capacity/rate + slack).
	tokenBucketKeyTTLSec = 120
)

// tokenBucketLua implements a Redis token bucket atomically.
//
// KEYS[1] = tokens key, KEYS[2] = last refill timestamp key (ms)
// ARGV[1] = rate (tokens/sec), ARGV[2] = capacity, ARGV[3] = now_ms,
// ARGV[4] = requested tokens, ARGV[5] = key TTL seconds
//
// Returns: {allowed (1|0), tokens_left (number)}
const tokenBucketLua = `
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

local tokens = tonumber(redis.call('GET', KEYS[1]))
local last = tonumber(redis.call('GET', KEYS[2]))

if tokens == nil or last == nil then
  tokens = capacity
  last = now
end

local delta = (now - last) / 1000.0
if delta < 0 then
  delta = 0
end
tokens = math.min(capacity, tokens + delta * rate)

local allowed = 0
if tokens >= requested then
  tokens = tokens - requested
  allowed = 1
end

redis.call('SET', KEYS[1], tokens)
redis.call('SET', KEYS[2], now)
redis.call('EXPIRE', KEYS[1], ttl)
redis.call('EXPIRE', KEYS[2], ttl)

return {allowed, tostring(tokens)}
`

func tokenBucketKeys(userID string) (tokensKey, tsKey string) {
	return "tb:tokens:" + userID, "tb:ts:" + userID
}

// allowTokenBucket returns true if the request is admitted under the per-user token bucket.
func allowTokenBucket(ctx context.Context, userID string) (allowed bool, tokensLeft float64, err error) {
	tk, tsk := tokenBucketKeys(userID)
	nowMs := float64(nowUnixMilli())
	res, err := dao.Rdb.Eval(ctx, tokenBucketLua,
		[]string{tk, tsk},
		tokenBucketRate,
		tokenBucketCapacity,
		nowMs,
		tokenBucketCost,
		tokenBucketKeyTTLSec,
	).Result()
	if err != nil {
		return false, 0, err
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, 0, fmt.Errorf("token bucket: unexpected redis reply %T", res)
	}
	var allowInt int64
	switch v := arr[0].(type) {
	case int64:
		allowInt = v
	case string:
		fmt.Sscan(v, &allowInt)
	default:
		return false, 0, fmt.Errorf("token bucket: bad allowed type %T", arr[0])
	}
	switch v := arr[1].(type) {
	case string:
		fmt.Sscan(v, &tokensLeft)
	case int64:
		tokensLeft = float64(v)
	case float64:
		tokensLeft = v
	}
	return allowInt == 1, tokensLeft, nil
}
