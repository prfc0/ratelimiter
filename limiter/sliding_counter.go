package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowCounter struct {
	client     *redis.Client
	limit      int
	windowSize time.Duration
	script     *redis.Script
}

func NewSlidingWindowCounter(client *redis.Client, limit int, window time.Duration) *SlidingWindowCounter {
	luaScript := `
local key    = KEYS[1]
local now    = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit  = tonumber(ARGV[3])

local curr_window = math.floor(now / window)
local prev_window = curr_window - 1

local curr_key = key .. ":" .. curr_window
local prev_key = key .. ":" .. prev_window

local curr_count = tonumber(redis.call("GET", curr_key) or "0")
local prev_count = tonumber(redis.call("GET", prev_key) or "0")

local elapsed = now % window
local weight = (window - elapsed) / window

local effective_count = curr_count + (prev_count * weight)

if effective_count < limit then
  curr_count = redis.call("INCR", curr_key)
	if curr_count == 1 then
	  redis.call("EXPIRE", curr_key, window * 2)
	end
	return 1
end

return 0
`

	return &SlidingWindowCounter{
		client:     client,
		limit:      limit,
		windowSize: window,
		script:     redis.NewScript(luaScript),
	}
}

func (sw *SlidingWindowCounter) Allow(key string) (bool, error) {
	ctx := context.Background()
	now := time.Now().Unix()

	result, err := sw.script.Run(ctx, sw.client,
		[]string{fmt.Sprintf("swcounter:%s", key)},
		now,
		int64(sw.windowSize.Seconds()),
		sw.limit,
	).Int()

	if err != nil {
		return false, err
	}
	return result == 1, nil
}
