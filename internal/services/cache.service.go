package services

import (
	"context"
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"log"
	"time"
)

type CacheService struct {
	GormStore   *gorm.GormStore
	RedisClient *redis.RedisClient
	Logger      *log.Logger
}

func NewCacheService(gormStore *gorm.GormStore, redisClient *redis.RedisClient) *CacheService {
	return &CacheService{
		GormStore:   gormStore,
		RedisClient: redisClient,
		Logger:      log.New(log.Writer(), "CacheService", log.LstdFlags),
	}
}

func (s *CacheService) Start(ctx context.Context, interval time.Duration) {
	s.syncCacheProviders()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncCacheProviders()
		case <-ctx.Done():
			if s.Logger != nil {
				s.Logger.Println("Cache Service stopped")
			}
			return
		}
	}
}

func (s *CacheService) syncCacheProviders() {
	// 1. 查询数据库
	data, err := s.GormStore.QueryProviders()
	if err != nil {
		s.Logger.Println("failed to query store", err)
		return
	}

	// 2. 写入Redis
	if err := s.RedisClient.CacheProviders(data); err != nil {
		s.Logger.Println("failed to save to redis", err)
		return
	}

	// 记录成功日志
	s.Logger.Printf("Cache %d providers successfully", len(data))
}
