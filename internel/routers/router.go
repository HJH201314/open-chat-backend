package routers

import (
	"github.com/fcraft/open-chat/internel/handlers"
	"github.com/fcraft/open-chat/internel/handlers/chat"
	"github.com/fcraft/open-chat/internel/handlers/manage"
	"github.com/fcraft/open-chat/internel/handlers/user"
	"github.com/fcraft/open-chat/internel/storage/gorm"
	"github.com/gin-gonic/gin"
)

type Router struct {
	Engine *gin.Engine
}

func InitRouter(r *gin.Engine, store *gorm.GormStore) Router {
	baseHandler := &handlers.BaseHandler{Store: store}

	// routes for chat completion
	chatHandler := chat.NewChatHandler(baseHandler)
	chatGroup := r.Group("/chat")
	{
		chatSessionGroup := chatGroup.Group("/session")
		{
			chatSessionGroup.GET("/new", chatHandler.CreateSession)
			chatSessionGroup.POST("/del/:session_id", chatHandler.DeleteSession)
		}
		chatCompletionGroup := chatGroup.Group("/completion")
		{
			chatCompletionGroup.POST("/stream/:session_id", chatHandler.CompletionStream)
		}
	}

	// routes for user
	userHandler := user.NewUserHandler(baseHandler)
	userGroup := r.Group("/user")
	{
		userGroup.POST("/ping", userHandler.Ping)
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/register", userHandler.Register)
	}

	// routes for management
	manageHandler := manage.NewManageHandler(baseHandler)
	manageGroup := r.Group("/manage")
	{
		manageProviderGroup := manageGroup.Group("/provider")
		{
			manageProviderGroup.POST("/create", manageHandler.CreateProvider)
			manageProviderGroup.GET("/:provider_id", manageHandler.GetProvider)
			manageProviderGroup.GET("/list", manageHandler.GetProviders)
			manageProviderGroup.POST("/update", manageHandler.UpdateProvider)
			manageProviderGroup.POST("/delete/:provider_id", manageHandler.DeleteProvider)
		}
		manageModelGroup := manageGroup.Group("/model")
		{
			manageModelGroup.POST("/create", manageHandler.CreateModel)
			manageModelGroup.GET("/:model_id", manageHandler.GetModel)
			manageModelGroup.GET("/list/:provider_id", manageHandler.GetModelsByProvider)
			manageModelGroup.POST("/update", manageHandler.UpdateModel)
			manageModelGroup.POST("/delete/:model_id", manageHandler.DeleteModel)
		}
	}

	return Router{Engine: r}
}
