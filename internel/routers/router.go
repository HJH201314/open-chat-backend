package routers

import (
	"github.com/fcraft/open-chat/internel/handlers/chat"
	"github.com/fcraft/open-chat/internel/handlers/user"
	"github.com/fcraft/open-chat/internel/storage"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine, store *storage.GormStore) {
	// router for chat completion
	chatHandler := chat.NewChatHandler(store)
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

	// router for user
	userHandler := user.NewUserHandler(store)
	userGroup := r.Group("/user")
	{
		userGroup.POST("/ping", user.Ping)
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/register", userHandler.Register)
	}
}
