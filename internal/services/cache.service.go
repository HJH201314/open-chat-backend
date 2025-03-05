package services

import (
	"context"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"time"
)

type CacheService struct {
	Gorm       *gorm.DB
	GormStore  *gormstore.GormStore
	Redis      *redis.Client
	RedisStore *redisstore.RedisStore
	Logger     *log.Logger
}

func NewCacheService(gormStore *gormstore.GormStore, redisClient *redisstore.RedisStore) *CacheService {
	return &CacheService{
		Gorm:       gormStore.Db,
		GormStore:  gormStore,
		Redis:      redisClient.Client,
		RedisStore: redisClient,
		Logger:     log.New(log.Writer(), "Cache", log.LstdFlags),
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
	if err := s.RedisStore.CacheProviders(data); err != nil {
		s.Logger.Println("failed to save to redis", err)
		return
	}

	// 记录成功日志
	s.Logger.Printf("Cache %d providers successfully", len(data))
}
