package routers

import (
	_ "github.com/fcraft/open-chat/docs"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/handlers/chat"
	"github.com/fcraft/open-chat/internal/handlers/course"
	"github.com/fcraft/open-chat/internal/handlers/manage"
	"github.com/fcraft/open-chat/internal/handlers/user"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	Engine *gin.Engine
}

func InitRouter(r *gin.Engine, store *gorm.GormStore, redis *redis.RedisStore, cache *services.CacheService) Router {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.DeepLinking(true)))

	baseHandler := handlers.NewBaseHandler(store, redis, cache)

	// routes for chat completion
	chatHandler := chat.NewChatHandler(baseHandler)
	chatGroup := r.Group("/chat")
	{
		chatConfigGroup := chatGroup.Group("/config")
		{
			chatConfigGroup.GET("/models", chatHandler.GetModels)
		}
		chatSessionGroup := chatGroup.Group("/session")
		{
			chatSessionGroup.POST("/new", chatHandler.CreateSession)
			chatSessionGroup.GET("/list", chatHandler.GetSessions)
			chatSessionGroup.GET("/:session_id", chatHandler.GetSession)
			chatSessionGroup.GET("/user/:session_id", chatHandler.GetUserSession)
			chatSessionGroup.POST("/update/:session_id", chatHandler.UpdateSession)
			chatSessionGroup.POST("/share/:session_id", chatHandler.ShareSession)
			chatSessionGroup.POST("/del/:session_id", chatHandler.DeleteSession)
		}
		chatMessageGroup := chatGroup.Group("/message")
		{
			chatMessageGroup.GET("/list/:session_id", chatHandler.GetMessages)
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
		userGroup.GET("/refresh", userHandler.Refresh)
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/logout", userHandler.Logout)
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
		manageApiKeyGroup := manageGroup.Group("/key")
		{
			manageApiKeyGroup.POST("/create", manageHandler.CreateAPIKey)
			manageApiKeyGroup.POST("/delete/:key_id", manageHandler.DeleteAPIKey)
		}
		manageModelGroup := manageGroup.Group("/schema")
		{
			manageModelGroup.POST("/create", manageHandler.CreateModel)
			manageModelGroup.GET("/:model_id", manageHandler.GetModel)
			manageModelGroup.GET("/list/:provider_id", manageHandler.GetModelsByProvider)
			manageModelGroup.POST("/update", manageHandler.UpdateModel)
			manageModelGroup.POST("/delete/:model_id", manageHandler.DeleteModel)
		}
	}

	// routes for tue
	tueHandler := course.NewCourseHandler(baseHandler)
	tueGroup := r.Group("/tue")
	{
		tueProblemGroup := tueGroup.Group("/problem")
		{
			tueProblemGroup.GET("/:id", tueHandler.GetProblem)
			tueProblemGroup.POST("/create", tueHandler.CreateProblem)
			tueProblemGroup.GET("/list", tueHandler.GetProblems)
		}
		tueExamGroup := tueGroup.Group("/exam")
		{
			tueExamGroup.GET("/:id", tueHandler.GetExam)
			tueExamGroup.POST("/create", tueHandler.CreateExam)
		}
	}

	return Router{Engine: r}
}
