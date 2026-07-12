package ratelimiter

import (
	"context"
	"fmt"
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
	rc *internalredis.RedisClient
}

func NewRedisRateLimiter(rc *internalredis.RedisClient) RateLimiter {
	return &redisRateLimiter{rc: rc}
}

func (rl *redisRateLimiter) Allow(ctx context.Context, key string, policy Policy) (Result, error) {
	nowMs := time.Now().UnixNano() / int64(time.Millisecond)

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
