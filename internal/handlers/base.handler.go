package handlers

import (
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/services"
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	storehelper "github.com/fcraft/open-chat/internal/storage/helper"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
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

// GetPublicKey
//
//	@Summary		获取公钥
//	@Description	获取公钥
//	@Tags			Base
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[string]	"RSA 公钥"
//	@Router			/base/public-key [get]
func (h *BaseHandler) GetPublicKey(c *gin.Context) {
	ctx_utils.Success(c, services.GetEncryptService().PublicKeyPEM)
}
