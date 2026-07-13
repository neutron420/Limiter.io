package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	internalredis "limiter.io/internal/redis"
)

type AlgorithmType string

const (
	TokenBucket          AlgorithmType = "token_bucket"
	FixedWindow          AlgorithmType = "fixed_window"
	SlidingWindowCounter AlgorithmType = "sliding_window_counter"
	SlidingWindowLog     AlgorithmType = "sliding_window_log"
	LeakyBucket          AlgorithmType = "leaky_bucket"
)

type Policy struct {
	Limit     int           // Max requests
	Period    time.Duration // Time window
	Burst     int           // Max capacity (Token Bucket / Leaky Bucket)
	Algorithm AlgorithmType
}

type Result struct {
	Allowed   bool
	Remaining int
	Limit     int
	Reset     time.Duration
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, policy Policy) (Result, error)
}

type redisRateLimiter struct {
	rc          *internalredis.RedisClient
	mu          sync.Mutex
	localTokens map[string]float64
	localLast   map[string]time.Time
}

func NewRedisRateLimiter(rc *internalredis.RedisClient) RateLimiter {
	return &redisRateLimiter{
		rc:          rc,
		localTokens: make(map[string]float64),
		localLast:   make(map[string]time.Time),
	}
}

func (rl *redisRateLimiter) Allow(ctx context.Context, key string, policy Policy) (Result, error) {
	nowMs := time.Now().UnixNano() / int64(time.Millisecond)

	// Try evaluating in Redis
	res, err := rl.evalRedis(ctx, key, policy, nowMs)
	if err == nil {
		return res, nil
	}

	// Redis connection failed! Failover to Local RAM Cache (Failover Mode)
	fmt.Printf("[FAILOVER] Redis offline: evaluating rate limit for key %s in local memory\n", key)
	return rl.evalLocal(key, policy)
}

func (rl *redisRateLimiter) evalRedis(ctx context.Context, key string, policy Policy, nowMs int64) (Result, error) {

	switch policy.Algorithm {
	case TokenBucket:
		burst := policy.Burst
		if burst <= 0 {
			burst = policy.Limit // default fallback if burst not set
		}

		refillRate := float64(policy.Limit) / float64(policy.Period.Milliseconds())
		ttl := int64(policy.Period.Seconds() * 2)

		redisKey := fmt.Sprintf("rate_limit:token_bucket:%s", key)
		res, err := rl.rc.Client.EvalSha(ctx, rl.rc.TokenBucketSHA, []string{redisKey},
			burst, refillRate, nowMs, 1, ttl).Result()
		if err != nil {
			return Result{}, fmt.Errorf("evaluating token bucket failed: %w", err)
		}

		slice, ok := res.([]interface{})
		if !ok || len(slice) < 2 {
			return Result{}, fmt.Errorf("unexpected script return format: %v", res)
		}

		allowed := slice[0].(int64) == 1
		remaining := int(slice[1].(int64))

		return Result{
			Allowed:   allowed,
			Remaining: remaining,
			Limit:     policy.Limit,
			Reset:     policy.Period,
		}, nil

	case FixedWindow:
		periodSec := int64(policy.Period.Seconds())
		// Generate bucket based on timestamp to make it fully atomic & clean
		bucket := time.Now().Unix() / periodSec
		redisKey := fmt.Sprintf("rate_limit:fixed_window:%s:%d", key, bucket)

		res, err := rl.rc.Client.EvalSha(ctx, rl.rc.FixedWindowSHA, []string{redisKey},
			policy.Limit, periodSec).Result()
		if err != nil {
			return Result{}, fmt.Errorf("evaluating fixed window failed: %w", err)
		}

		slice, ok := res.([]interface{})
		if !ok || len(slice) < 2 {
			return Result{}, fmt.Errorf("unexpected script return format: %v", res)
		}

		allowed := slice[0].(int64) == 1
		remaining := int(slice[1].(int64))

		// Calculate reset time
		nowSec := time.Now().Unix()
		resetSec := ((nowSec / periodSec) + 1) * periodSec
		resetDuration := time.Duration(resetSec-nowSec) * time.Second

		return Result{
			Allowed:   allowed,
			Remaining: remaining,
			Limit:     policy.Limit,
			Reset:     resetDuration,
		}, nil

	case SlidingWindowCounter:
		nowSec := time.Now().Unix()
		periodSec := int64(policy.Period.Seconds())

		currentWindow := nowSec / periodSec
		prevWindow := currentWindow - 1
		fraction := float64(nowSec%periodSec) / float64(periodSec)

		redisKeyPrefix := fmt.Sprintf("rate_limit:sliding_window_ctr:%s", key)
		res, err := rl.rc.Client.EvalSha(ctx, rl.rc.SlidingWindowCtrSHA, []string{redisKeyPrefix},
			currentWindow, prevWindow, policy.Limit, periodSec, fraction).Result()
		if err != nil {
			return Result{}, fmt.Errorf("evaluating sliding window counter failed: %w", err)
		}

		slice, ok := res.([]interface{})
		if !ok || len(slice) < 2 {
			return Result{}, fmt.Errorf("unexpected script return format: %v", res)
		}

		allowed := slice[0].(int64) == 1
		remaining := int(slice[1].(int64))

		// Calculate time left in current window
		nextWindowStart := (currentWindow + 1) * periodSec
		resetDuration := time.Duration(nextWindowStart-nowSec) * time.Second

		return Result{
			Allowed:   allowed,
			Remaining: remaining,
			Limit:     policy.Limit,
			Reset:     resetDuration,
		}, nil

	case SlidingWindowLog:
		periodMs := policy.Period.Milliseconds()
		redisKey := fmt.Sprintf("rate_limit:sliding_window_log:%s", key)

		res, err := rl.rc.Client.EvalSha(ctx, rl.rc.SlidingWindowLogSHA, []string{redisKey},
			nowMs, periodMs, policy.Limit).Result()
		if err != nil {
			return Result{}, fmt.Errorf("evaluating sliding window log failed: %w", err)
		}

		slice, ok := res.([]interface{})
		if !ok || len(slice) < 2 {
			return Result{}, fmt.Errorf("unexpected script return format: %v", res)
		}

		allowed := slice[0].(int64) == 1
		remaining := int(slice[1].(int64))

		return Result{
			Allowed:   allowed,
			Remaining: remaining,
			Limit:     policy.Limit,
			Reset:     policy.Period,
		}, nil

	case LeakyBucket:
		capacity := policy.Burst
		if capacity <= 0 {
			capacity = policy.Limit
		}

		leakRate := float64(policy.Limit) / float64(policy.Period.Milliseconds())
		ttl := int64(policy.Period.Seconds() * 2)

		redisKey := fmt.Sprintf("rate_limit:leaky_bucket:%s", key)
		res, err := rl.rc.Client.EvalSha(ctx, rl.rc.LeakyBucketSHA, []string{redisKey},
			capacity, leakRate, nowMs, ttl).Result()
		if err != nil {
			return Result{}, fmt.Errorf("evaluating leaky bucket failed: %w", err)
		}

		slice, ok := res.([]interface{})
		if !ok || len(slice) < 2 {
			return Result{}, fmt.Errorf("unexpected script return format: %v", res)
		}

		allowed := slice[0].(int64) == 1
		remaining := int(slice[1].(int64))

		return Result{
			Allowed:   allowed,
			Remaining: remaining,
			Limit:     policy.Limit,
			Reset:     policy.Period,
		}, nil

	default:
		return Result{}, fmt.Errorf("unsupported algorithm: %s", policy.Algorithm)
	}
}

func (rl *redisRateLimiter) evalLocal(key string, policy Policy) (Result, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	capacity := float64(policy.Limit)
	if policy.Burst > 0 {
		capacity = float64(policy.Burst)
	}

	// Local token bucket simulation
	tokens, exists := rl.localTokens[key]
	lastTime, timeExists := rl.localLast[key]

	if !exists || !timeExists {
		tokens = capacity
		lastTime = now
	}

	elapsed := now.Sub(lastTime).Seconds()
	refillRate := float64(policy.Limit) / policy.Period.Seconds()

	tokens = tokens + (elapsed * refillRate)
	if tokens > capacity {
		tokens = capacity
	}

	rl.localLast[key] = now
	allowed := false

	if tokens >= 1.0 {
		tokens -= 1.0
		allowed = true
	}
	rl.localTokens[key] = tokens

	return Result{
		Allowed:   allowed,
		Remaining: int(tokens),
		Limit:     policy.Limit,
		Reset:     policy.Period,
	}, nil
}

func ParseAlgorithm(algo string) AlgorithmType {
	switch algo {
	case "token_bucket":
		return TokenBucket
	case "fixed_window":
		return FixedWindow
	case "sliding_window_counter":
		return SlidingWindowCounter
	case "sliding_window_log":
		return SlidingWindowLog
	case "leaky_bucket":
		return LeakyBucket
	default:
		return TokenBucket
	}
}
