package main

import (
	"github.com/fcraft/open-chat/internel/middlewares"
	"github.com/fcraft/open-chat/internel/routers"
	"github.com/fcraft/open-chat/internel/storage"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	// 初始化数据库
	store := storage.InitGormStore()

	r := gin.Default()
	// 初始化中间件
	r.Use(middlewares.AuthMiddleware())
	// 初始化路由
	routers.InitRouter(r, store)

	os.Setenv("API_KEY_DEEPSEEK", "")
	os.Setenv("API_KEY_GPT", "")
	os.Setenv("AUTH_SECRET", "auth-secret")

	// 在 9033 端口启动服务
	if err := r.Run("0.0.0.0:9033"); err != nil {
		panic("failed to run server")
	}
}
