package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) Set(key, value string) error {
	err := r.Client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key: %v", err)
	}
	return nil
}

func (r *RedisClient) SetIfNotExists(key, value string) (bool, error) {
	success, err := r.Client.SetNX(ctx, key, value, 0).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set key if not exists: %v", err)
	}
	return success, nil
}

func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key: %v", err)
	}
	return val, nil
}
