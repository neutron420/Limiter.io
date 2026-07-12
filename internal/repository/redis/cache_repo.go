package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	internalredis "limiter.io/internal/redis"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type cacheRepo struct {
	rc *internalredis.RedisClient
}

func NewCacheRepository(rc *internalredis.RedisClient) repository.CacheRepository {
	return &cacheRepo{rc: rc}
}

func (r *cacheRepo) GetAPIKey(ctx context.Context, hash string) (*models.APIKey, error) {
	key := fmt.Sprintf("rate_limit:cache:api_key:%s", hash)
	val, err := r.rc.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.New("api key cache miss")
	} else if err != nil {
		return nil, err
	}

	var apiKey models.APIKey
	if err := json.Unmarshal([]byte(val), &apiKey); err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *cacheRepo) SetAPIKey(ctx context.Context, hash string, apiKey *models.APIKey, ttl time.Duration) error {
	key := fmt.Sprintf("rate_limit:cache:api_key:%s", hash)
	data, err := json.Marshal(apiKey)
	if err != nil {
		return err
	}
	return r.rc.Client.Set(ctx, key, data, ttl).Err()
}

func (r *cacheRepo) DeleteAPIKey(ctx context.Context, hash string) error {
	key := fmt.Sprintf("rate_limit:cache:api_key:%s", hash)
	return r.rc.Client.Del(ctx, key).Err()
}

func (r *cacheRepo) GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	key := fmt.Sprintf("rate_limit:cache:subscription:%s", userID.String())
	val, err := r.rc.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.New("subscription cache miss")
	} else if err != nil {
		return nil, err
	}

	var sub models.Subscription
	if err := json.Unmarshal([]byte(val), &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *cacheRepo) SetSubscription(ctx context.Context, userID uuid.UUID, sub *models.Subscription, ttl time.Duration) error {
	key := fmt.Sprintf("rate_limit:cache:subscription:%s", userID.String())
	data, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	return r.rc.Client.Set(ctx, key, data, ttl).Err()
}

func (r *cacheRepo) DeleteSubscription(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("rate_limit:cache:subscription:%s", userID.String())
	return r.rc.Client.Del(ctx, key).Err()
}
