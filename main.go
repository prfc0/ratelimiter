package main

import (
	"fmt"
	"log"
	"net/http"
	"ratelimiter/config"
	"ratelimiter/limiter"
	"ratelimiter/proxy"
	"ratelimiter/store"
	"time"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	// Setup store (Redis)
	redisStore := store.NewRedisStore("localhost:6379")

	rule := cfg.Rules[3]
	var rl limiter.RateLimiter

	switch rule.Algorithm {
	case "fixed_window":
		rl = limiter.NewFixedWindow(redisStore, rule.Limit, time.Duration(rule.WindowSeconds)*time.Second)
	case "token_bucket":
		rl = limiter.NewTokenBucket(redisStore, rule.Capacity, rule.RefillRate)
	case "sliding_log":
		rl = limiter.NewSlidingWindowLog(redisStore.Client(), rule.Limit, time.Duration(rule.WindowSeconds)*time.Second)
	case "sliding_counter":
		rl = limiter.NewSlidingWindowCounter(redisStore.Client(), rule.Limit, time.Duration(rule.WindowSeconds)*time.Second)
	default:
		log.Fatalf("unsupported algorithm: %s", rule.Algorithm)
	}

	// Setup proxy
	p, err := proxy.NewProxy(cfg.Backend.URL, rule, rl)
	if err != nil {
		log.Fatal("failed to create proxy:", err)
	}

	// Start server
	fmt.Println("Rate limiter proxy listening on :8080")
	http.HandleFunc("/", p.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
