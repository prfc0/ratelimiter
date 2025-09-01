package limiter

import (
	"context"
	"fmt"
	"time"

	"ratelimiter/store"
)

type TokenBucket struct {
	store      store.Store
	capacity   int
	refillRate int // tokens per second
}

func NewTokenBucket(store store.Store, capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		store:      store,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (tb *TokenBucket) Allow(key string) (bool, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("tb:%s", key)
	now := time.Now().Unix()

	// We use a secondary key to track last refill timestamp
	lastRefillKey := redisKey + ":ts"

	// Fetch current tokens
	tokens, err := tb.store.Get(ctx, redisKey)
	if err != nil && err.Error() != "redis: nil" {
		return false, err
	}
	if tokens < 0 {
		tokens = int64(tb.capacity)
	}

	// Fetch last refill time
	lastRefill, err := tb.store.Get(ctx, lastRefillKey)
	if err != nil && err.Error() != "redis: nil" {
		return false, err
	}

	// Calculate elapsed time
	var elapsed int64 = 0
	if lastRefill > 0 {
		elapsed = now - lastRefill
	}

	// Refill tokens
	newTokens := tokens + int64(elapsed*int64(tb.refillRate))
	if newTokens > int64(tb.capacity) {
		newTokens = int64(tb.capacity)
	}

	// Update last refill timestamp
	if err := tb.store.Set(ctx, lastRefillKey, now, 3600); err != nil {
		return false, err
	}

	// Consume token if available
	if newTokens > 0 {
		newTokens--
		if err := tb.store.Set(ctx, redisKey, newTokens, 3600); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
