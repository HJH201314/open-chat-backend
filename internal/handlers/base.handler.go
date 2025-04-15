package handlers

import (
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/services"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	storehelper "github.com/fcraft/open-chat/internal/storage/helper"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"gorm.io/gorm"
)

type BaseHandler struct {
	Store  *gormstore.GormStore
	Db     *gorm.DB
	Redis  *redisstore.RedisStore
	Cache  *services.CacheService
	Helper *storehelper.QueryHelper
}

func NewBaseHandler(store *gormstore.GormStore, redis *redisstore.RedisStore, helper *storehelper.QueryHelper, cache *services.CacheService) *BaseHandler {
	return &BaseHandler{
		Store:  store,
		Db:     store.Db,
		Redis:  redis,
		Cache:  cache,
		Helper: helper,
	}
}
