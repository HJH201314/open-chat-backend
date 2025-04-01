package main

import (
	"context"
	"github.com/fcraft/open-chat/internal/storage/helper"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fcraft/open-chat/internal/middlewares"
	"github.com/fcraft/open-chat/internal/routers"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	chat_utils.TestExtractTagContent()
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stdout, slogcolor.DefaultOptions)))

	// 加载环境变量
	loadEnv()

	// 初始化数据库
	store := gorm.NewGormStore()
	rd := redis.NewRedisStore()
	hp := helper.NewHandlerHelper(store, rd)

	// 启动缓存服务
	cacheCtx, cancelCache := context.WithCancel(context.Background())
	defer cancelCache()
	baseService := services.NewBaseService(store, rd, hp)         // 基础服务
	intervalCacheService := services.NewCacheService(baseService) // 定时缓存服务
	go intervalCacheService.Start(cacheCtx, 10*time.Minute)
	services.InitEncryptService()
	services.InitPresetService(baseService) // 初始化预设缓存服务
	services.InitMakeQuestionService(baseService)
	//go services.GetMakeQuestionService().StartGenerate(context.Background(), 10*time.Second)

	r := gin.Default()
	// 初始化中间件
	r.Use(middlewares.AuthMiddleware(rd))
	r.Use(middlewares.PermissionMiddleware(hp))
	// 初始化路由
	routers.InitRouter(r, store, rd, hp, intervalCacheService)

	// 在 9033 端口启动服务
	if err := r.Run("0.0.0.0:9033"); err != nil {
		log.Fatal("Error running server", err.Error())
	}
}

// loadEnv 加载环境变量
// 加载顺序：.env.{TUE_ENV}.local > .env.local > .env.{TUE_ENV} > .env
// TUE_ENV 默认为 development
func loadEnv() {
	// 默认加载 development 变量
	env := os.Getenv("TUE_ENV")
	if "" == env {
		env = "development"
	}
	// .env.{TUE_ENV}.local
	_ = godotenv.Load(".env." + env + ".local")
	// .env.local
	_ = godotenv.Load(".env.local")
	// .env.{TUE_ENV}
	_ = godotenv.Load(".env." + env)
	// .env
	_ = godotenv.Load()
}
