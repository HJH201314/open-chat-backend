package services

import (
	"context"
	"fmt"
	"time"
)

type CacheService struct {
	BaseService
}

func NewCacheService(baseService *BaseService) *CacheService {
	return &CacheService{
		BaseService: *baseService,
	}
}

func (s *CacheService) Start(ctx context.Context, interval time.Duration) {
	s.syncAll()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-ctx.Done():
			s.Logger.Info("Cache Service stopped")
			return
		}
	}
}

func (s *CacheService) syncAll() {
	s.syncCacheProviders()
	s.cachePresets()
}

// syncCacheProviders 缓存 provider
func (s *CacheService) syncCacheProviders() {
	// 1. 查询数据库
	data, err := s.GormStore.QueryProviders()
	if err != nil {
		s.Logger.Error("provider_model failed to query store" + err.Error())
		return
	}

	// 2. 写入Redis
	if err := s.RedisStore.CacheProviders(data); err != nil {
		s.Logger.Error("provider_model failed to save to redis: " + err.Error())
		return
	}

	// 记录成功日志
	s.Logger.Info(fmt.Sprintf("Cache %d providers successfully", len(data)))
}

// cachePresets 缓存 presets
func (s *CacheService) cachePresets() {
	// 1. 查询数据库
	data, err := s.GormStore.ListPresets()
	if err != nil {
		s.Logger.Error("preset failed to query store" + err.Error())
		return
	}

	// 2. 写入Redis
	if err := s.RedisStore.CachePresets(data); err != nil {
		s.Logger.Error("preset failed to save to redis: " + err.Error())
		return
	}

	// 记录成功日志
	s.Logger.Info(fmt.Sprintf("Cache %d bots successfully", len(data)))
}
