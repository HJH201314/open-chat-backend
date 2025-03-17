package handlers

import (
	"github.com/fcraft/open-chat/internal/services"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"gorm.io/gorm"
)

type BaseHandler struct {
	Store  *gormstore.GormStore
	Db     *gorm.DB
	Redis  *redisstore.RedisStore
	Cache  *services.CacheService
	Helper *HandlerHelper
}

func NewBaseHandler(store *gormstore.GormStore, redis *redisstore.RedisStore, cache *services.CacheService) *BaseHandler {
	return &BaseHandler{
		Store:  store,
		Db:     store.Db,
		Redis:  redis,
		Cache:  cache,
		Helper: NewHandlerHelper(store, redis),
	}
}
