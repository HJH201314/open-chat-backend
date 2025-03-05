package redis

import (
	"context"
	"fmt"
)

// DeleteSessionCache 删除会话缓存
func (r *RedisStore) DeleteSessionCache(sessionId string) error {
	ctx := context.Background()
	pipe := r.Client.Pipeline()

	// 扫描并删除所有 user-session:{sessionId}:* 键
	userSessionKeys, _ := r.Client.Keys(ctx, fmt.Sprintf("user-session:%s:*", sessionId)).Result()
	if len(userSessionKeys) > 0 {
		pipe.Del(ctx, userSessionKeys...)
	}
	_, err := pipe.Exec(ctx)
	return err
}
