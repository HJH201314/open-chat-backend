package redis

import (
	"context"
	"encoding/json"
	"github.com/fcraft/open-chat/internal/models"
	"time"
)

func (r *RedisClient) CacheProviders(providers []models.Provider) error {
	pipe := r.Client.Pipeline()
	pipe.Del(context.Background(), "providers")

	for _, provider := range providers {
		data, _ := json.Marshal(provider)
		pipe.RPush(
			context.Background(),
			"providers", data,
		)
	}

	pipe.Expire(
		context.Background(),
		"providers", 1*time.Hour,
	)

	_, err := pipe.Exec(context.Background())
	return err
}

func (r *RedisClient) GetCachedProviders() ([]models.Provider, error) {
	data, err := r.Client.LRange(context.Background(), "providers", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var providers []models.Provider
	for _, v := range data {
		var provider models.Provider
		if err := json.Unmarshal([]byte(v), &provider); err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}

	return providers, nil
}
