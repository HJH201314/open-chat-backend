package routers

import (
	"reflect"
	"runtime"
	"strings"

	"gorm.io/gorm/clause"

	"github.com/fcraft/open-chat/internal/storage/helper"

	_ "github.com/fcraft/open-chat/docs"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/handlers/chat"
	"github.com/fcraft/open-chat/internal/handlers/course"
	"github.com/fcraft/open-chat/internal/handlers/exam"
	"github.com/fcraft/open-chat/internal/handlers/manage"
	"github.com/fcraft/open-chat/internal/handlers/user"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/services"
	"github.com/fcraft/open-chat/internal/storage/gorm"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouteInfo 存储路由信息
type RouteInfo struct {
	Method      string // HTTP 方法
	Path        string // 路由路径
	Name        string // 权限名称
	Description string // 权限描述
	Module      string // 所属模块
}

type Router struct {
	Engine     *gin.Engine
	store      *gorm.GormStore
	routeInfos []RouteInfo
}

var (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

// getHandlerName 通过反射获取处理函数的名称
func getHandlerName(handler gin.HandlerFunc) string {
	// 获取处理函数的指针
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}

	// 获取函数名称
	handlerName := runtime.FuncForPC(handlerValue.Pointer()).Name()

	// 提取函数名
	parts := strings.Split(handlerName, ".")
	if len(parts) >= 1 {
		// 返回最后一部分（函数名）
		funcName, _ := strings.CutSuffix(parts[len(parts)-1], "-fm")
		return funcName
	}

	return handlerName
}

// getModuleName 通过反射获取处理函数的模块名
func getModuleName(handler gin.HandlerFunc) string {
	// 获取处理函数的指针
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}

	// 获取函数名称
	handlerName := runtime.FuncForPC(handlerValue.Pointer()).Name()

	// 提取包名和函数名
	parts := strings.Split(handlerName, ".")
	if len(parts) >= 3 {
		moduleParts := strings.Split(parts[len(parts)-3], "/")
		// 返回倒数第三个部分（模块路径）的最后一部分（模块名）
		return moduleParts[len(moduleParts)-1]
	}

	return ""
}

// registerRoute 收集路由信息并注册路由
func (r *Router) registerRoute(group *gin.RouterGroup, method, path string, description string, handler gin.HandlerFunc) {
	// 注册路由
	switch method {
	case GET:
		group.GET(path, handler)
	case POST:
		group.POST(path, handler)
	case PUT:
		group.PUT(path, handler)
	case DELETE:
		group.DELETE(path, handler)
	case PATCH:
		group.PATCH(path, handler)
	}

	// 收集路由信息
	moduleName := getModuleName(handler)
	funcName := getHandlerName(handler)
	r.routeInfos = append(
		r.routeInfos, RouteInfo{
			Method:      method,
			Path:        group.BasePath() + path,
			Name:        strings.Join([]string{moduleName, funcName}, "."),
			Description: description,
			Module:      moduleName,
		},
	)
}

// saveRoutesToDB 将收集到的路由信息保存到数据库
func (r *Router) saveRoutesToDB() error {
	var permissions []schema.Permission
	for _, route := range r.routeInfos {
		permissions = append(
			permissions, schema.Permission{
				Name:        route.Name,
				Path:        route.Method + ":" + route.Path,
				Description: route.Description,
				Module:      route.Module,
			},
		)
	}
	// 使用 Upsert 功能，当 Path 已存在时更新，不存在时创建
	if err := r.store.Db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			UpdateAll: true,
		},
	).CreateInBatches(&permissions, 100).Error; err != nil {
		return err
	}
	return nil
}

func InitRouter(r *gin.Engine, store *gorm.GormStore, redis *redis.RedisStore, helper *helper.QueryHelper, cache *services.CacheService) Router {
	router := Router{
		Engine: r,
		store:  store,
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.DeepLinking(true)))

	baseHandler := handlers.NewBaseHandler(store, redis, helper, cache)

	// routes for chat completion
	chatHandler := chat.NewChatHandler(baseHandler)
	chatGroup := r.Group("/chat")
	{
		chatConfigGroup := chatGroup.Group("/config")
		{
			router.registerRoute(
				chatConfigGroup,
				GET,
				"/models",
				"获取可用的聊天模型列表",

				chatHandler.GetModelConfig,
			)
			router.registerRoute(
				chatConfigGroup,
				GET,
				"/bots",
				"获取可用的 bot 角色列表",

				chatHandler.GetBotConfig,
			)
		}
		// routes for preset
		botRoleGroup := r.Group("/preset")
		{
			router.registerRoute(
				botRoleGroup,
				POST,
				"/create",
				"创建新的预设",
				chatHandler.CreatePreset,
			)
			router.registerRoute(
				botRoleGroup,
				GET,
				"/list",
				"获取预设列表",
				chatHandler.ListPresets,
			)
			router.registerRoute(
				botRoleGroup,
				GET,
				"/:id",
				"获取指定预设的详细信息",
				chatHandler.GetPreset,
			)
			router.registerRoute(
				botRoleGroup,
				POST,
				"/:id/update",
				"更新预设信息",
				chatHandler.UpdatePreset,
			)
			router.registerRoute(
				botRoleGroup,
				POST,
				"/:id/delete",
				"删除指定的预设",
				chatHandler.DeletePreset,
			)
		}

		chatSessionGroup := chatGroup.Group("/session")
		{
			router.registerRoute(
				chatSessionGroup,
				POST,
				"/new",
				"创建新的聊天会话",

				chatHandler.CreateSession,
			)
			router.registerRoute(
				chatSessionGroup,
				GET,
				"/list",
				"获取当前用户的聊天会话列表",

				chatHandler.GetSessions,
			)
			router.registerRoute(
				chatSessionGroup,
				GET,
				"/sync",
				"同步聊天会话列表",
				chatHandler.SyncSessions,
			)
			router.registerRoute(
				chatSessionGroup,
				GET,
				"/:session_id",
				"获取指定会话的详细信息",

				chatHandler.GetSession,
			)
			router.registerRoute(
				chatSessionGroup,
				GET,
				"/user/:session_id",
				"获取指定用户的会话信息",

				chatHandler.GetUserSession,
			)
			router.registerRoute(
				chatSessionGroup,
				POST,
				"/update/:session_id",
				"更新指定会话的信息",

				chatHandler.UpdateSession,
			)
			router.registerRoute(
				chatSessionGroup,
				POST,
				"/flag/:session_id",
				"更新指定聊天会话标记",

				chatHandler.UpdateSessionFlag,
			)
			router.registerRoute(
				chatSessionGroup,
				POST,
				"/share/:session_id",
				"分享指定的聊天会话",

				chatHandler.ShareSession,
			)
			router.registerRoute(
				chatSessionGroup,
				POST,
				"/del/:session_id",
				"删除指定的聊天会话",

				chatHandler.DeleteSession,
			)
		}
		chatMessageGroup := chatGroup.Group("/message")
		{
			router.registerRoute(
				chatMessageGroup,
				GET,
				"/list/:session_id",
				"获取指定会话的消息列表",

				chatHandler.GetMessages,
			)
		}
		chatCompletionGroup := chatGroup.Group("/completion")
		{
			router.registerRoute(
				chatCompletionGroup,
				POST,
				"/stream/:session_id",
				"与AI进行流式对话",

				chatHandler.CompletionStream,
			)
		}
	}

	// routes for user
	userHandler := user.NewUserHandler(baseHandler)
	userGroup := r.Group("/user")
	{
		router.registerRoute(userGroup, POST, "/ping", "检查用户登录状态", userHandler.Ping)
		router.registerRoute(userGroup, GET, "/refresh", "刷新用户的访问令牌", userHandler.Refresh)
		router.registerRoute(userGroup, POST, "/login", "用户登录接口", userHandler.Login)
		router.registerRoute(userGroup, GET, "/current", "当前用户信息", userHandler.Current)
		router.registerRoute(userGroup, POST, "/logout", "用户登出接口", userHandler.Logout)
		router.registerRoute(userGroup, POST, "/register", "新用户注册接口", userHandler.Register)
	}

	// routes for management
	manageHandler := manage.NewManageHandler(baseHandler)
	manageGroup := r.Group("/manage")
	{
		manageProviderGroup := manageGroup.Group("/provider")
		{
			router.registerRoute(
				manageProviderGroup,
				POST,
				"/create",
				"创建新的AI提供商",

				manageHandler.CreateProvider,
			)
			router.registerRoute(
				manageProviderGroup,
				GET,
				"/:provider_id",
				"获取指定提供商的详细信息",

				manageHandler.GetProvider,
			)
			router.registerRoute(
				manageProviderGroup,
				GET,
				"/list",
				"分页获取AI提供商列表",

				manageHandler.GetProviders,
			)
			router.registerRoute(
				manageProviderGroup,
				GET,
				"/all",
				"获取所有AI提供商列表",

				manageHandler.GetAllProviders,
			)
			router.registerRoute(
				manageProviderGroup,
				POST,
				"/:id/update",
				"更新AI提供商信息",

				manageHandler.UpdateProvider,
			)
			router.registerRoute(
				manageProviderGroup,
				POST,
				"/:id/delete",
				"删除指定的AI提供商",

				manageHandler.DeleteProvider,
			)
		}
		manageApiKeyGroup := manageGroup.Group("/key")
		{
			router.registerRoute(
				manageApiKeyGroup,
				POST,
				"/create",
				"创建新的API访问密钥",

				manageHandler.CreateAPIKey,
			)
			router.registerRoute(
				manageApiKeyGroup,
				POST,
				"/:id/delete",
				"删除指定的API访问密钥",

				manageHandler.DeleteAPIKey,
			)
			router.registerRoute(
				manageApiKeyGroup,
				GET,
				"/list/provider/:id",
				"分页获取 API Key",

				manageHandler.GetAPIKeyByProvider,
			)
		}
		manageModelGroup := manageGroup.Group("/model")
		{
			router.registerRoute(
				manageModelGroup,
				POST,
				"/create",
				"创建新的AI模型",

				manageHandler.CreateModel,
			)
			router.registerRoute(
				manageModelGroup,
				GET,
				"/:model_id",
				"获取指定模型的详细信息",

				manageHandler.GetModel,
			)
			router.registerRoute(
				manageModelGroup,
				GET,
				"/list",
				"分页获取所有模型列表",

				manageHandler.GetModels,
			)
			router.registerRoute(
				manageModelGroup,
				GET,
				"/provider/:provider_id",
				"获取指定提供商的所有模型列表",

				manageHandler.GetModelsByProvider,
			)
			router.registerRoute(
				manageModelGroup,
				POST,
				"/update",
				"更新AI模型信息",

				manageHandler.UpdateModel,
			)
			router.registerRoute(
				manageModelGroup,
				POST,
				"/delete/:model_id",
				"删除指定的AI模型",

				manageHandler.DeleteModel,
			)
		}
		manageCollectionGroup := manageGroup.Group("/collection")
		{
			router.registerRoute(
				manageCollectionGroup,
				POST,
				"/create",
				"创建新的模型集合",

				manageHandler.CreateModelCollection,
			)
			router.registerRoute(
				manageCollectionGroup,
				GET,
				"/:collection_id",
				"获取指定模型的详细信息",

				manageHandler.GetModelCollection,
			)
			router.registerRoute(
				manageCollectionGroup,
				GET,
				"/list",
				"分页获取所有模型列表",

				manageHandler.GetModelCollections,
			)
			router.registerRoute(
				manageCollectionGroup,
				POST,
				"/delete/:collection_id",
				"删除指定的AI模型",

				manageHandler.DeleteModelCollection,
			)
		}
	}

	// routes for tue
	tueHandler := course.NewCourseHandler(baseHandler)
	tueGroup := r.Group("/tue")
	{
		tueProblemGroup := tueGroup.Group("/problem")
		{
			router.registerRoute(
				tueProblemGroup,
				GET,
				"/:id",
				"获取指定题目的详细信息",

				tueHandler.GetProblem,
			)
			router.registerRoute(
				tueProblemGroup,
				POST,
				"/create",
				"创建新的题目",

				tueHandler.CreateProblem,
			)
			router.registerRoute(
				tueProblemGroup,
				GET,
				"/list",
				"获取所有题目列表",

				tueHandler.GetProblems,
			)
		}
		tueExamGroup := tueGroup.Group("/exam")
		{
			router.registerRoute(tueExamGroup, GET, "/:id", "获取指定考试的详细信息", tueHandler.GetExam)
			router.registerRoute(tueExamGroup, POST, "/create", "创建新的考试", tueHandler.CreateExam)
			// 考试提交
			examHandler := exam.NewExamHandler(baseHandler)
			router.registerRoute(
				tueExamGroup,
				POST,
				"/:id/submit",
				"提交考试答案",
				examHandler.SubmitExam,
			)
			router.registerRoute(
				tueExamGroup,
				GET,
				"/:id/records",
				"获取考试结果",
				examHandler.GetExamResult,
			)
			router.registerRoute(
				tueExamGroup,
				POST,
				"/:id/rescore",
				"重新评分考试",
				examHandler.RescoreExam,
			)
		}
		tueCourseGroup := tueGroup.Group("/course")
		{
			router.registerRoute(
				tueCourseGroup,
				GET,
				"/:id",
				"获取指定课程的详细信息",

				tueHandler.GetCourse,
			)
			router.registerRoute(tueCourseGroup, POST, "/create", "创建新的课程", tueHandler.CreateCourse)
			router.registerRoute(tueCourseGroup, POST, "/update", "更新课程信息", tueHandler.UpdateCourse)
			router.registerRoute(
				tueCourseGroup,
				POST,
				"/delete/:id",
				"删除指定的课程",

				tueHandler.DeleteCourse,
			)
			router.registerRoute(
				tueCourseGroup,
				GET,
				"/list",
				"获取所有课程列表",

				tueHandler.GetCourses,
			)
		}
	}

	// 保存路由信息到数据库
	if err := router.saveRoutesToDB(); err != nil {
		panic("Failed to save routes to database: " + err.Error())
	}

	return router
}
