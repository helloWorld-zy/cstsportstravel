package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/travel-booking/server/internal/common/config"
)

// Redis wraps the go-redis client with convenience methods.
type Redis struct {
	client *redis.Client
}

// NewRedis creates a new Redis client with connection pool and health check.
func NewRedis(cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{client: client}, nil
}

// Client returns the underlying redis.Client.
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// HealthCheck verifies Redis connectivity.
func (r *Redis) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
