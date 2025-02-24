package handlers

import (
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/redis"
)

type BaseHandler struct {
	Store *gorm.GormStore
	Redis *redis.RedisClient
}
