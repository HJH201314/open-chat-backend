package services

import (
	"context"
	"fmt"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/helper"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

type CacheService struct {
	Gorm       *gorm.DB
	GormStore  *gormstore.GormStore
	Redis      *redis.Client
	RedisStore *redisstore.RedisStore
	Helper     *helper.HandlerHelper
	Logger     *slog.Logger
}

func NewCacheService(gormStore *gormstore.GormStore, redisClient *redisstore.RedisStore, handlerHelper *helper.HandlerHelper) *CacheService {
	return &CacheService{
		Gorm:       gormStore.Db,
		GormStore:  gormStore,
		Redis:      redisClient.Client,
		RedisStore: redisClient,
		Helper:     handlerHelper,
		Logger:     slog.Default(),
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
