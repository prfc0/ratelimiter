package limiter

import (
	"context"
	"fmt"
	"time"

	"ratelimiter/store"
)

type FixedWindow struct {
	store  store.Store
	limit  int
	window time.Duration
}

func NewFixedWindow(store store.Store, limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		store:  store,
		limit:  limit,
		window: window,
	}
}

func (fw *FixedWindow) Allow(key string) (bool, error) {
	ctx := context.Background()

	// Calculate window key
	windowKey := fmt.Sprintf("fw:%s:%d", key, time.Now().Unix()/int64(fw.window.Seconds()))

	// Increment counter
	count, err := fw.store.Incr(ctx, windowKey)
	if err != nil {
		return false, err
	}

	// Expire after window size
	if count == 1 {
		if err := fw.store.Expire(ctx, windowKey, int64(fw.window.Seconds())); err != nil {
			return false, err
		}
	}

	if int(count) > fw.limit {
		return false, nil
	}
	return true, nil
}
