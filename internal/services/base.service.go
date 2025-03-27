package services

import (
	gormstore "github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/helper"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

type BaseService struct {
	Gorm       *gorm.DB
	GormStore  *gormstore.GormStore
	Redis      *redis.Client
	RedisStore *redisstore.RedisStore
	Helper     *helper.QueryHelper
	Logger     *slog.Logger
}

func NewBaseService(gormStore *gormstore.GormStore, redisClient *redisstore.RedisStore, handlerHelper *helper.QueryHelper) *BaseService {
	return &BaseService{
		Gorm:       gormStore.Db,
		GormStore:  gormStore,
		Redis:      redisClient.Client,
		RedisStore: redisClient,
		Helper:     handlerHelper,
		Logger:     slog.Default(),
	}
}
