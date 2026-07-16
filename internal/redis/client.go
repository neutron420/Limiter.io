package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"limiter.io/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client                *redis.Client
	TokenBucketSHA        string
	FixedWindowSHA        string
	SlidingWindowCtrSHA   string
	SlidingWindowLogSHA   string
	LeakyBucketSHA        string
}

//nosec
const (
	tokenBucketLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2]) -- tokens per millisecond
local now = tonumber(ARGV[3]) -- now in millisecond
local requested = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5]) -- key TTL in seconds

local state = redis.call('HMGET', key, 'tokens', 'last_updated')
local tokens = tonumber(state[1])
local last_updated = tonumber(state[2])

if not tokens then
    tokens = capacity
    last_updated = now
else
    local elapsed = math.max(0, now - last_updated)
    tokens = math.min(capacity, tokens + (elapsed * refill_rate))
end

local allowed = 0
if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
end

redis.call('HMSET', key, 'tokens', tokens, 'last_updated', now)
redis.call('EXPIRE', key, ttl)

return {allowed, math.floor(tokens)}
`

	fixedWindowLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local period = tonumber(ARGV[2]) -- period in seconds

local count = redis.call('INCR', key)
if count == 1 then
    redis.call('EXPIRE', key, period)
end

local allowed = 0
if count <= limit then
    allowed = 1
end

return {allowed, limit - count}
`

	slidingWindowCtrLua = `
local key_prefix = KEYS[1]
local current_window = ARGV[1]
local prev_window = ARGV[2]
local limit = tonumber(ARGV[3])
local period = tonumber(ARGV[4]) -- window period in seconds
local fraction = tonumber(ARGV[5]) -- fraction of current window elapsed

local cur_key = key_prefix .. ":" .. current_window
local prev_key = key_prefix .. ":" .. prev_window

local cur_count = tonumber(redis.call('GET', cur_key) or "0")
local prev_count = tonumber(redis.call('GET', prev_key) or "0")

local estimated_count = math.floor(prev_count * (1 - fraction) + cur_count)

local allowed = 0
if estimated_count < limit then
    redis.call('INCR', cur_key)
    redis.call('EXPIRE', cur_key, period * 2)
    allowed = 1
    cur_count = cur_count + 1
end

return {allowed, limit - (cur_count + math.floor(prev_count * (1 - fraction)))}
`

	slidingWindowLogLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1]) -- now in millisecond
local window = tonumber(ARGV[2]) -- window size in millisecond
local limit = tonumber(ARGV[3])

local clear_before = now - window
redis.call('ZREMRANGEBYSCORE', key, 0, clear_before)

local current_requests = redis.call('ZCARD', key)
local allowed = 0
if current_requests < limit then
    redis.call('ZADD', key, now, now)
    redis.call('EXPIRE', key, math.ceil(window / 1000) * 2)
    allowed = 1
    current_requests = current_requests + 1
end

return {allowed, limit - current_requests}
`

	leakyBucketLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local leak_rate = tonumber(ARGV[2]) -- water leaked per millisecond
local now = tonumber(ARGV[3]) -- now in millisecond
local ttl = tonumber(ARGV[4]) -- TTL in seconds

local state = redis.call('HMGET', key, 'water_level', 'last_leak')
local water_level = tonumber(state[1]) or 0
local last_leak = tonumber(state[2]) or now

local elapsed = math.max(0, now - last_leak)
local leaked = elapsed * leak_rate
water_level = math.max(0, water_level - leaked)

local allowed = 0
if water_level + 1 <= capacity then
    water_level = water_level + 1
    last_leak = now
    allowed = 1
end

redis.call('HMSET', key, 'water_level', water_level, 'last_leak', last_leak)
redis.call('EXPIRE', key, ttl)

return {allowed, math.ceil(capacity - water_level)}
`
)

func ConnectRedis(cfg *config.Config) (*RedisClient, error) {
	var rdb *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		rdb = redis.NewClient(opt)
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	for i := 1; i <= 5; i++ {
		_, err = rdb.Ping(ctx).Result()
		if err == nil {
			break
		}
		log.Printf("Failed to connect to Redis (attempt %d/5): %v. Retrying in 3 seconds...", i, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to Redis: %w", err)
	}

	// Load Lua scripts into Redis
	tbSHA, err := rdb.ScriptLoad(ctx, tokenBucketLua).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Token Bucket Lua script: %w", err)
	}

	fwSHA, err := rdb.ScriptLoad(ctx, fixedWindowLua).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Fixed Window Lua script: %w", err)
	}

	swcSHA, err := rdb.ScriptLoad(ctx, slidingWindowCtrLua).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Sliding Window Counter Lua script: %w", err)
	}

	swlSHA, err := rdb.ScriptLoad(ctx, slidingWindowLogLua).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Sliding Window Log Lua script: %w", err)
	}

	lbSHA, err := rdb.ScriptLoad(ctx, leakyBucketLua).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Leaky Bucket Lua script: %w", err)
	}

	return &RedisClient{
		Client:                rdb,
		TokenBucketSHA:        tbSHA,
		FixedWindowSHA:        fwSHA,
		SlidingWindowCtrSHA:   swcSHA,
		SlidingWindowLogSHA:   swlSHA,
		LeakyBucketSHA:        lbSHA,
	}, nil
}
