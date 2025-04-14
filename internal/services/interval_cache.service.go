package services

import (
	"fmt"
	"time"
)

type CacheService struct {
	BaseService
}

func NewCacheService(baseService *BaseService) *CacheService {
	cacheService := &CacheService{
		BaseService: *baseService,
	}
	err := GetScheduleService().RegisterSchedule(
		"cache_providers", "缓存接入点和模型", 10*time.Minute, func() error {
			return cacheService.syncAll()
		},
	)
	if err != nil {
		return nil
	}
	return cacheService
}

func (s *CacheService) syncAll() error {
	if err := s.SyncCacheProviders(); err != nil {
		return err
	}

	if err := s.CachePresets(); err != nil {
		return err
	}

	return nil
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
