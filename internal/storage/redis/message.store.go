package redis

import (
	"context"
	"encoding/json"
	"github.com/fcraft/open-chat/internal/schema"
	"time"
)

func (r *RedisClient) CacheMessages(sessionID string, messages []schema.Message) error {
	// 保留最近50条消息
	pipe := r.Client.Pipeline()
	pipe.Del(context.Background(), "messages:"+sessionID)

	for _, msg := range messages {
		data, _ := json.Marshal(msg)
		pipe.RPush(
			context.Background(),
			"messages:"+sessionID, data,
		)
	}

	pipe.LTrim(
		context.Background(),
		"messages:"+sessionID, -50, -1,
	)
	pipe.Expire(
		context.Background(),
		"messages:"+sessionID, 1*time.Hour,
	)

	_, err := pipe.Exec(context.Background())
	return err
}

func (r *RedisClient) GetCachedMessages(sessionID string) ([]schema.Message, error) {
	data, err := r.Client.LRange(context.Background(), "messages:"+sessionID, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []schema.Message
	for _, v := range data {
		var msg schema.Message
		if err := json.Unmarshal([]byte(v), &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
