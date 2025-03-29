package redis

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"time"
)

// CacheUserToken 缓存用户 token
func (r *RedisStore) CacheUserToken(userId uint64, token string, duration time.Duration) error {
	// token-user:{token} -> userId; user-tokens:{userId} -> [...token]
	ctx := context.Background()
	pipe := r.Client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("token-user:%s", token), userId, duration)
	pipe.SAdd(ctx, fmt.Sprintf("user-tokens:%d", userId), token)
	_, err := pipe.Exec(ctx)
	return err
}

// InvalidUserToken 取消缓存用户 token
func (r *RedisStore) InvalidUserToken(userId uint64, token string) error {
	ctx := context.Background()
	pipe := r.Client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("token-user:%s", token))
	pipe.SRem(ctx, fmt.Sprintf("user-tokens:%d", userId), token)
	_, err := pipe.Exec(ctx)
	return err
}

// InvalidUserAllToken 取消缓存用户所有 token
//
//	Returns:
//		int   用户 token 数量
func (r *RedisStore) InvalidUserAllToken(userId uint64) (int, error) {
	ctx := context.Background()
	tokens, err := r.Client.SMembers(ctx, fmt.Sprintf("user-tokens:%d", userId)).Result()
	pipe := r.Client.Pipeline()
	if len(tokens) > 0 {
		pipe.Del(ctx, slice.Map(tokens, func(_ int, token string) string { return "token-user:" + token })...)
		pipe.Del(ctx, fmt.Sprintf("user-tokens:%d", userId))
	}
	_, err = pipe.Exec(ctx)
	return len(tokens), err
}

// FindUserIdByToken 根据 token 获取用户 ID
func (r *RedisStore) FindUserIdByToken(token string) (uint64, error) {
	userId, err := r.Client.Get(context.Background(), "token-user:"+token).Uint64()
	if err != nil {
		return 0, err
	}
	return userId, nil
}
