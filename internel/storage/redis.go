package storage

import (
	"context"
	"encoding/json"
	"github.com/fcraft/open-chat/internel/models"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisStore struct {
	client *redis.Client
}

func (r *RedisStore) CacheMessages(sessionID string, messages []models.Message) error {
	// 保留最近50条消息
	pipe := r.client.Pipeline()
	pipe.Del(context.Background(), "messages:"+sessionID)

	for _, msg := range messages {
		data, _ := json.Marshal(msg)
		pipe.RPush(context.Background(),
			"messages:"+sessionID, data)
	}

	pipe.LTrim(context.Background(),
		"messages:"+sessionID, -50, -1)
	pipe.Expire(context.Background(),
		"messages:"+sessionID, 1*time.Hour)

	_, err := pipe.Exec(context.Background())
	return err
}
