package cache

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
}

func NewCache() *Cache {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "redis:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Cache{Client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value string) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}
