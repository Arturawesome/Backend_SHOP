package redisCashe

import (
	"BE2_API_Shop/shop/cfg"
	"BE2_API_Shop/shop/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

var _ repository.ImageCashe = (*RedisClient)(nil)

func NewRedisClient(ctx context.Context, cfg cfg.RedisConfig) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return RedisClient{}, err
	}

	return RedisClient{client: client}, nil
}

func (r *RedisClient) GetByID(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

func (r *RedisClient) Set(ctx context.Context, key string, image []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, image, ttl).Err()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}
