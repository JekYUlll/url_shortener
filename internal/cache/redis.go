package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	clent *redis.Client
}

func NewRedisCache() *RedisCache {
	return &RedisCache{
		redis.NewClient()
	}
}

func (c *RedisCache) SetURL(ctx context.Context, url model.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	cmd := c.clent.Set(ctx, url.ShortCode, data, time.Until(url.ExpiredAt))
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *RedisCache) GetURL(ctx context.Context, shortCode string) (*model.URL, error) {
	cmd := c.clent.Get(ctx, shortCode)
	if err := cmd.Err(); err != nil {
		// Get 不到值时会返回 redis.Nil，这是「缓存未命中」的正常情况
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	data := cmd.Val()
	var u model.URL
	if err := json.Unmarshal([]byte(data), &u); err != nil {
		return nil, err
	}
	return &u, nil
}
