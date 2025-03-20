package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/redis/go-redis/v9"
	"time"
)

// CacheProviders 缓存供应商及关联的模型数据（使用哈希表）
func (r *RedisStore) CacheProviders(providers []schema.Provider) error {
	// 删除所有旧的 Provider 和 Model 键
	ctx := context.Background()
	pipe := r.Client.Pipeline()

	// 1. 扫描并删除所有 provider:* 和 schema:* 键
	providerKeys, _ := r.Client.Keys(ctx, "provider:*").Result()
	modelKeys, _ := r.Client.Keys(ctx, "schema:*").Result()
	if len(providerKeys) > 0 {
		pipe.Del(ctx, providerKeys...)
	}
	if len(modelKeys) > 0 {
		pipe.Del(ctx, modelKeys...)
	}

	// 2. 缓存新的 Providers 和 Models
	for _, provider := range providers {
		// 缓存 Provider 基本信息
		providerKey := fmt.Sprintf("provider:%s", provider.Name)
		providerData, _ := json.Marshal(provider)
		pipe.HSet(
			ctx, providerKey, map[string]interface{}{
				"data": providerData, // 整个对象序列化存储
			},
		)
		pipe.Expire(ctx, providerKey, 1*time.Hour)

		// 缓存关联的 Models
		for _, model := range provider.Models {
			modelKey := fmt.Sprintf("schema:%s:%s", provider.Name, model.Name)
			cacheModel := schema.ModelCache{
				Model:               model,
				ProviderName:        provider.Name,
				ProviderDisplayName: provider.DisplayName,
			}
			modelData, _ := json.Marshal(cacheModel)
			pipe.HSet(
				ctx, modelKey, map[string]interface{}{
					"data": modelData,
				},
			)
			pipe.Expire(ctx, modelKey, 1*time.Hour)
		}
	}

	// 执行管道命令
	_, err := pipe.Exec(ctx)
	return err
}

// GetCachedProviders 获取所有缓存的 Provider
func (r *RedisStore) GetCachedProviders() ([]schema.Provider, error) {
	ctx := context.Background()

	// 1. 获取所有 Provider 键
	providerKeys, err := r.Client.Keys(ctx, "provider:*").Result()
	if err != nil {
		return nil, err
	}

	// 2. 批量获取 Provider 数据
	pipe := r.Client.Pipeline()
	cmds := make([]*redis.StringCmd, len(providerKeys))
	for i, key := range providerKeys {
		cmds[i] = pipe.HGet(ctx, key, "data")
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	// 3. 反序列化数据
	var providers []schema.Provider
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil || data == "" {
			continue
		}
		var provider schema.Provider
		if err := json.Unmarshal([]byte(data), &provider); err != nil {
			continue // 或记录错误
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetCachedModels 获取所有缓存的 Model
func (r *RedisStore) GetCachedModels() ([]schema.ModelCache, error) {
	ctx := context.Background()

	// 1. 获取所有 Model 键
	modelKeys, err := r.Client.Keys(ctx, "schema:*").Result()
	if err != nil {
		return nil, err
	}

	// 2. 批量获取 Model 数据
	pipe := r.Client.Pipeline()
	cmds := make([]*redis.StringCmd, len(modelKeys))
	for i, key := range modelKeys {
		cmds[i] = pipe.HGet(ctx, key, "data")
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	// 3. 反序列化数据
	var modelCached []schema.ModelCache
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil || data == "" {
			continue
		}
		var mc schema.ModelCache
		if err := json.Unmarshal([]byte(data), &mc); err != nil {
			continue // 或记录错误
		}
		modelCached = append(modelCached, mc)
	}

	return modelCached, nil
}

// FindProviderByName 根据 ProviderName 获取供应商
func (r *RedisStore) FindProviderByName(providerName string) *schema.Provider {
	ctx := context.Background()

	// 构造符合哈希表存储规则的 key
	key := fmt.Sprintf("provider:%s", providerName)

	// 直接通过 HGet 获取数据
	data, err := r.Client.HGet(ctx, key, "data").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			// 可选：记录非"key不存在"类错误日志
			// log.Printf("HGet error: %v", err)
		}
		return nil
	}

	// 反序列化数据
	var providerCache schema.Provider
	if err := json.Unmarshal([]byte(data), &providerCache); err != nil {
		// 可选：记录数据损坏错误
		// log.Printf("Unmarshal error: %v", err)
		return nil
	}

	return &providerCache
}

// FindCachedModelByName 根据 ProviderName 和 ModelName 直接定位缓存模型
func (r *RedisStore) FindCachedModelByName(providerName string, modelName string) *schema.ModelCache {
	ctx := context.Background()

	// 构造符合哈希表存储规则的 key
	key := fmt.Sprintf("schema:%s:%s", providerName, modelName)

	// 直接通过 HGet 获取数据
	data, err := r.Client.HGet(ctx, key, "data").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			// 可选：记录非"key不存在"类错误日志
			// log.Printf("HGet error: %v", err)
		}
		return nil
	}

	// 反序列化数据
	var modelCache schema.ModelCache
	if err := json.Unmarshal([]byte(data), &modelCache); err != nil {
		// 可选：记录数据损坏错误
		// log.Printf("Unmarshal error: %v", err)
		return nil
	}

	return &modelCache
}
