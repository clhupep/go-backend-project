package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductCache struct{ Client *redis.Client }

func NewProductCache(client *redis.Client) *ProductCache { return &ProductCache{Client: client} }
func (c *ProductCache) Version(ctx context.Context) (string, error) {
	version, err := c.Client.Get(ctx, "catalog:version").Result()
	if err == redis.Nil {
		return "0", nil
	}
	return version, err
}
func (c *ProductCache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.Client.Get(ctx, key).Bytes()
}
func (c *ProductCache) Set(ctx context.Context, key string, data []byte) error {
	return c.Client.Set(ctx, key, data, 90*time.Second).Err()
}
func (c *ProductCache) Invalidate(ctx context.Context) error {
	return c.Client.Incr(ctx, "catalog:version").Err()
}
