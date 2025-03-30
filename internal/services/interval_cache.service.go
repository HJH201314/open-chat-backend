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
	if err := s.SyncCacheProviders(); err != nil {
		// do nothing
	}

	if err := s.CachePresets(); err != nil {
		// do nothing
	}
}

// SyncCacheProviders 缓存 provider
func (s *CacheService) SyncCacheProviders() error {
	// 1. 查询数据库
	data, err := s.GormStore.QueryProviders()
	if err != nil {
		s.Logger.Error("provider_model failed to query store" + err.Error())
		return err
	}

	// 2. 写入Redis
	if err := s.RedisStore.CacheProviders(data); err != nil {
		s.Logger.Error("provider_model failed to save to redis: " + err.Error())
		return err
	}

	// 记录成功日志
	s.Logger.Info(fmt.Sprintf("Cache %d providers successfully", len(data)))
	return nil
}

// CachePresets 缓存 presets
func (s *CacheService) CachePresets() error {
	// 1. 查询数据库
	data, err := s.GormStore.ListPresets()
	if err != nil {
		s.Logger.Error("preset failed to query store" + err.Error())
		return err
	}

	// 2. 写入Redis
	if err := s.RedisStore.CachePresets(data); err != nil {
		s.Logger.Error("preset failed to save to redis: " + err.Error())
		return err
	}

	// 记录成功日志
	s.Logger.Info(fmt.Sprintf("Cache %d bots successfully", len(data)))
	return nil
}
