package main

import (
	"github.com/fcraft/open-chat/internel/middlewares"
	"github.com/fcraft/open-chat/internel/models"
	"github.com/fcraft/open-chat/internel/routers"
	"github.com/fcraft/open-chat/internel/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func main() {
	// 初始化 Postgres 连接
	dsn := "xxx"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 自动迁移表结构
	db.AutoMigrate(&models.Session{}, &models.Message{})
	// 初始化 GORM 存储
	store := storage.NewGormStore(db)

	r := gin.Default()
	// 初始化中间件
	r.Use(middlewares.AuthMiddleware())
	r.Use(middlewares.ResponseMiddleware())
	// 初始化路由
	routers.InitRouter(r, store)

	os.Setenv("API_KEY_DEEPSEEK", "sk-xxx")
	os.Setenv("AUTH_SECRET", "xxx")

	// 在 8080 端口启动服务
	if err := r.Run(); err != nil {
		panic("failed to run server")
	}
}
