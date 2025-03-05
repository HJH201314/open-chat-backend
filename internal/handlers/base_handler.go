package handlers

import (
	"github.com/fcraft/open-chat/internal/services"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
)

type BaseHandler struct {
	Store  *gormstore.GormStore
	Redis  *redisstore.RedisStore
	Cache  *services.CacheService
	Helper *HandlerHelper
}

func NewBaseHandler(store *gormstore.GormStore, redis *redisstore.RedisStore, cache *services.CacheService) *BaseHandler {
	return &BaseHandler{
		Store:  store,
		Redis:  redis,
		Cache:  cache,
		Helper: NewHandlerHelper(store, redis),
	}
}
