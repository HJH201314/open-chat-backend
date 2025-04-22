package main

import (
	"github.com/MatusOllah/slogcolor"
	"github.com/fcraft/open-chat/internal/middlewares"
	"github.com/fcraft/open-chat/internal/routers"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/helper"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stdout, slogcolor.DefaultOptions)))

	// 加载环境变量
	loadEnv()

	// 初始化数据库
	store := gorm.NewGormStore()
	rd := redis.NewRedisStore()
	hp := helper.NewHandlerHelper(store, rd)

	baseService := services.NewBaseService(store, rd, hp)         // 基础服务
	services.InitPresetService(baseService)                       // 初始化预设缓存服务 !高优先级
	services.InitScheduleService(baseService)                     // 初始化定时任务服务 !高优先级
	services.InitSystemConfigService(baseService)                 // 初始化系统配置服务
	intervalCacheService := services.NewCacheService(baseService) // 定时缓存服务
	go services.InitEncryptService()
	go services.InitOAuthService(baseService)           // 注册OAuth服务
	go services.InitChatService(baseService)            // 注册对话服务
	go services.InitMakeQuestionService(baseService)    // 初始化题目生成服务
	go services.InitModelCollectionService(baseService) // 初始化模型集合服务

	services.GetScheduleService().StartSchedule() // 启动定时任务

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
	_ = godotenv.Load("./conf/.env." + env + ".local")
	// .env.local
	_ = godotenv.Load("./conf/.env.local")
	// .env.{TUE_ENV}
	_ = godotenv.Load("./conf/.env." + env)
	// .env
	_ = godotenv.Load("./conf/.env")
}
