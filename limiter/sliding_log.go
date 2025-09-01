package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLog struct {
	client     *redis.Client
	limit      int
	windowSize time.Duration
	script     *redis.Script
}

func NewSlidingWindowLog(client *redis.Client, limit int, window time.Duration) *SlidingWindowLog {
	luaScript := `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

redis.call("ZREMRANGEBYSCORE", key, 0, now - window)

local count = redis.call("ZCARD", key)
if count >= limit then
  return 0
end

redis.call("ZADD", key, now, now)
redis.call("PEXPIRE", key, window * 2)
return 1
`
	return &SlidingWindowLog{
		client:     client,
		limit:      limit,
		windowSize: window,
		script:     redis.NewScript(luaScript),
	}
}

func (sw *SlidingWindowLog) Allow(key string) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixMilli()

	result, err := sw.script.Run(ctx, sw.client,
		[]string{fmt.Sprintf("swlog:%s", key)},
		now,
		sw.windowSize.Milliseconds(),
		sw.limit,
	).Int()

	if err != nil {
		return false, err
	}
	return result == 1, nil
}
