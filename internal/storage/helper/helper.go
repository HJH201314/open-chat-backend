package helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type HandlerHelper struct {
	Gorm       *gorm.DB
	GormStore  *gormstore.GormStore
	Redis      *redis.Client
	RedisStore *redisstore.RedisStore
}

func NewHandlerHelper(gormStore *gormstore.GormStore, redisStore *redisstore.RedisStore) *HandlerHelper {
	return &HandlerHelper{Gorm: gormStore.Db, GormStore: gormStore, Redis: redisStore.Client, RedisStore: redisStore}
}

// CheckUserSession 检查用户是否拥有会话权限
func (s *HandlerHelper) CheckUserSession(userId uint64, sessionId string) bool {
	ctx := context.Background()
	// 查询 Redis user-session:{userId}:{sessionId} 是否存在
	_, err := s.Redis.Get(ctx, fmt.Sprintf("user-session:%s:%d", sessionId, userId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		// 查询异常
		return false
	}
	// 查询 GORM 是否存在
	var userSession schema.UserSession
	if err = s.Gorm.First(&userSession, "user_id = ? AND session_id = ?", userId, sessionId).Error; err != nil {
		// 查询不到或查询异常
		return false
	}
	// 存入 Redis
	s.Redis.Set(ctx, fmt.Sprintf("user-session:%s:%d", sessionId, userId), 1, 1*time.Hour)
	return true
}

// DeleteSession 删除会话
func (s *HandlerHelper) DeleteSession(sessionId string) error {
	// 删除数据
	if err := s.GormStore.DeleteSession(sessionId); err != nil {
		return err
	}
	// 删除缓存
	if err := s.RedisStore.DeleteSessionCache(sessionId); err != nil {
		return err
	}

	return nil
}
