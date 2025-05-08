package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jekyulll/url_shortener/config"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{
		client: client,
	}, nil
}

func (c *RedisCache) SetURL(ctx context.Context, url model.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	cmd := c.client.Set(ctx, url.ShortCode, data, time.Until(url.ExpiredAt))
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *RedisCache) GetURL(ctx context.Context, shortCode string) (*model.URL, error) {
	cmd := c.client.Get(ctx, shortCode)
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

func (c *RedisCache) Close() error {
	return c.client.Close()
}
