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
		chatGroup.POST("/completion/stream/:session_id", chatHandler.CompletionStream)
		chatGroup.GET("/session/new", chatHandler.CreateSession)
	}

	// router for user
	userGroup := r.Group("/user")
	{
		userGroup.POST("/ping", user.Ping)
	}
}
