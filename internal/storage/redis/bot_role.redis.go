package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fcraft/open-chat/internal/schema"
)

// CacheBotRole 缓存机器人角色
func (r *RedisStore) CacheBotRole(role *schema.BotRole) error {
	ctx := context.Background()
	key := fmt.Sprintf("bot-role:%d", role.ID)

	// 序列化角色数据
	data, err := json.Marshal(role)
	if err != nil {
		return err
	}

	// 存储到 Redis，设置1小时过期
	return r.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetCachedBotRole 获取缓存的机器人角色
func (r *RedisStore) GetCachedBotRole(id uint64) (*schema.BotRole, error) {
	ctx := context.Background()
	key := fmt.Sprintf("bot-role:%d", id)

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色数据
	var role schema.BotRole
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteBotRoleCache 删除机器人角色缓存
func (r *RedisStore) DeleteBotRoleCache(id uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("bot-role:%d", id)
	return r.Client.Del(ctx, key).Err()
}

// CacheBotRoles 缓存机器人角色列表
func (r *RedisStore) CacheBotRoles(roles []schema.BotRole) error {
	ctx := context.Background()
	key := "bot-roles:list"

	// 序列化角色列表数据
	data, err := json.Marshal(roles)
	if err != nil {
		return err
	}

	// 存储到 Redis，设置1小时过期
	return r.Client.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetCachedBotRoles 获取缓存的机器人角色列表
func (r *RedisStore) GetCachedBotRoles() ([]schema.BotRole, error) {
	ctx := context.Background()
	key := "bot-roles:list"

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色列表数据
	var roles []schema.BotRole
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// DeleteBotRolesCache 删除所有机器人角色缓存
func (r *RedisStore) DeleteBotRolesCache() error {
	ctx := context.Background()
	// 获取所有机器人角色缓存键
	keys, err := r.Client.Keys(ctx, "bot-role:*").Result()
	if err != nil {
		return err
	}

	// 如果有键，则删除
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}
