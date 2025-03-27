package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"time"

	"github.com/fcraft/open-chat/internal/schema"
)

// CachePreset 缓存预设
func (r *RedisStore) CachePreset(role *schema.Preset) error {
	ctx := context.Background()
	keyID := fmt.Sprintf("preset:%d", role.ID)
	keyName := fmt.Sprintf("preset:%s", role.Name)

	// 序列化角色数据
	data, err := json.Marshal(role)
	if err != nil {
		return err
	}

	// 存储到 Redis，设置1小时过期
	if err := r.Client.Set(ctx, keyID, data, 1*time.Hour).Err(); err != nil {
		return err
	}
	if err := r.Client.Set(ctx, keyName, data, 1*time.Hour).Err(); err != nil {
		return err
	}
	return nil
}

// GetCachedPresetByID 获取缓存的预设
func (r *RedisStore) GetCachedPresetByID(id uint64) (*schema.Preset, error) {
	ctx := context.Background()
	key := fmt.Sprintf("preset:%d", id)

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色数据
	var role schema.Preset
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// GetCachedPresetByName 获取缓存的预设
func (r *RedisStore) GetCachedPresetByName(name string) (*schema.Preset, error) {
	ctx := context.Background()
	key := fmt.Sprintf("preset:%s", name)

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色数据
	var role schema.Preset
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeletePresetCache 删除预设缓存
func (r *RedisStore) DeletePresetCache(id uint64, name string) error {
	ctx := context.Background()
	keyID := fmt.Sprintf("preset:%d", id)
	keyName := fmt.Sprintf("preset:%s", name)
	if err := r.Client.Del(ctx, keyID).Err(); err != nil {
		return err
	}
	if err := r.Client.Del(ctx, keyName).Err(); err != nil {
		return err
	}
	return nil
}

// CachePresets 缓存预设列表
func (r *RedisStore) CachePresets(presets []schema.Preset) error {
	ctx := context.Background()
	key := "presets:list"

	// 序列化角色列表数据
	data, err := json.Marshal(presets)
	if err != nil {
		return err
	}
	// 存储到 Redis，设置1小时过期
	if err := r.Client.Set(ctx, key, data, 1*time.Hour).Err(); err != nil {
		return err
	}

	// 按照类别进行分组
	presetsMap := slice.GroupWith(
		presets, func(item schema.Preset) string {
			return item.Module
		},
	)
	for module, presetsInModule := range presetsMap {
		// 序列化角色列表数据
		jsonData, err := json.Marshal(presetsInModule)
		if err != nil {
			return err
		}

		// 存储到 Redis，设置1小时过期
		if err := r.Client.Set(ctx, fmt.Sprintf("presets:list:%s", module), jsonData, 1*time.Hour).Err(); err != nil {
			return err
		}
	}

	return nil
}

// GetCachedPresets 获取缓存的预设列表
func (r *RedisStore) GetCachedPresets() ([]schema.Preset, error) {
	ctx := context.Background()
	key := "presets:list"

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色列表数据
	var roles []schema.Preset
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetCachedPresetsByModule 根据类别获取缓存的预设列表
func (r *RedisStore) GetCachedPresetsByModule(module string) ([]schema.Preset, error) {
	ctx := context.Background()
	key := fmt.Sprintf("presets:list:%s", module)

	// 从 Redis 获取数据
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化角色列表数据
	var roles []schema.Preset
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// DeletePresetsCache 删除所有预设缓存
func (r *RedisStore) DeletePresetsCache() error {
	ctx := context.Background()
	// 获取所有预设缓存键
	keys, err := r.Client.Keys(ctx, "preset:*").Result()
	if err != nil {
		return err
	}

	// 如果有键，则删除
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}
