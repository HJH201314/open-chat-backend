package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fcraft/open-chat/internal/schema"
)

// CacheUserRoles 缓存用户角色信息
func (r *RedisStore) CacheUserRoles(userId uint64, roles []schema.Role) error {
	ctx := context.Background()
	key := fmt.Sprintf("user-roles:%d", userId)

	// 序列化角色数据
	data, err := json.Marshal(roles)
	if err != nil {
		return err
	}

	// 存储到 Redis，设置1小时过期
	return r.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetCachedUserRoles 获取缓存的用户角色信息
func (r *RedisStore) GetCachedUserRoles(userId uint64) ([]schema.Role, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user-roles:%d", userId)

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色数据
	var roles []schema.Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// DeleteUserRolesCache 删除用户角色缓存
func (r *RedisStore) DeleteUserRolesCache(userId uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("user-roles:%d", userId)
	return r.Client.Del(ctx, key).Err()
}

// DeleteAllUserRolesCache 删除所有用户角色缓存
func (r *RedisStore) DeleteAllUserRolesCache() error {
	ctx := context.Background()
	// 获取所有用户角色缓存键
	keys, err := r.Client.Keys(ctx, "user-roles:*").Result()
	if err != nil {
		return err
	}

	// 如果有键，则删除
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}
