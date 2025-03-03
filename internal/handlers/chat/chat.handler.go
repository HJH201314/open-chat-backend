package chat

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entities"
	_ "github.com/fcraft/open-chat/internal/entities"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/models"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type PathParamSessionId struct {
	SessionId string `uri:"session_id" binding:"required"`
}

type Handler struct {
	*handlers.BaseHandler
}

func NewChatHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}

// CreateSession
//
//	@Summary		创建会话
//	@Description	创建会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entities.CommonResponse[string]
//	@Router			/chat/session/new [post]
func (h *Handler) CreateSession(c *gin.Context) {
	session := models.Session{
		UserID:        ctx_utils.GetUserId(c),
		EnableContext: true, // 默认开启上下文
	}
	if err := h.Store.CreateSession(&session); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create session")
		return
	}
	ctx_utils.Success(c, session.ID)
}

// DeleteSession
//
//	@Summary		删除会话
//	@Description	删除会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string	true	"会话 ID"
//	@Success		200			{object}	entities.CommonResponse[bool]
//	@Router			/chat/session/del/{session_id} [post]
func (h *Handler) DeleteSession(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if _, err := h.Store.FindSessionWithUser(uri.SessionId, ctx_utils.GetUserId(c)); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 执行删除操作
	if err := h.Store.DeleteSession(uri.SessionId); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to delete session")
		return
	}
	ctx_utils.Success(c, true)
}

// GetMessages
//
//	@Summary		获取消息
//	@Description	获取消息
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string												true	"会话 ID"
//	@Param			req			query		entities.ParamPagingSort							true	"分页参数"
//	@Success		200			{object}	entities.CommonResponse[chat.GetMessages.resType]	"返回数据"
//	@Router			/chat/message/list/{session_id} [get]
func (h *Handler) GetMessages(c *gin.Context) {
	var uri PathParamSessionId
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	var req entities.ParamPagingSort
	if err := c.BindQuery(&req); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	if _, err := h.Store.FindSessionWithUser(uri.SessionId, ctx_utils.GetUserId(c)); err != nil {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	// 查询消息
	size := req.GetPageSize(20, 50)
	messages, nextPage, err := h.Store.GetMessagesByPage(uri.SessionId, req.PageNum, size, req.SortParam)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	type resType struct {
		Messages []models.Message `json:"messages"`
		NextPage *int             `json:"next_page"`
	}
	ctx_utils.Success(
		c, &resType{
			Messages: messages,
			NextPage: nextPage,
		},
	)
}

// GetModels
//
//	@Summary		获取所有模型
//	@Description	获取所有模型
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entities.CommonResponse[[]models.ModelCache]
//	@Router			/chat/config/models [get]
func (h *Handler) GetModels(c *gin.Context) {
	// 从缓存中查询
	cacheModels, err := h.Redis.GetCachedModels()
	if err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to get models")
		return
	}
	// 将 config 隐藏
	slice.ForEach(
		cacheModels, func(_ int, item models.ModelCache) {
			item.Config = models.ModelConfig{}
		},
	)
	ctx_utils.Success(c, cacheModels)
}

// CompletionStream
//
//	@Summary		流式输出聊天
//	@Description	流式输出聊天
//	@Tags			Chat
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			session_id	path	string							true	"会话 ID"
//	@Param			request		body	chat.CompletionStream.userInput	true	"用户输入及参数"
//	@Router			/chat/completion/stream/{session_id} [post]
func (h *Handler) CompletionStream(c *gin.Context) {
	// 从 path 和 body 中获取用户输入
	var uri PathParamSessionId
	type userInput struct {
		Question      string  `json:"question" binding:"required"`
		Provider      string  `json:"provider_name" binding:"required"` // Provider.Name 准确的供应商名称
		ModelName     string  `json:"model_name" binding:"required"`    // Model.Name 准确的模型名称
		EnableContext *bool   `json:"enable_context" binding:"-"`
		SystemPrompt  *string `json:"system_prompt" binding:"-"` // 系统提示词
	}
	var req userInput
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Question == "" {
		ctx_utils.HttpError(c, constants.ErrBadRequest)
		return
	}

	// 读取模型信息
	modelInfo := h.Redis.FindCachedModelByName(req.Provider, req.ModelName)
	if modelInfo == nil {
		ctx_utils.CustomError(c, 404, "module not found")
		return
	}
	modelConfig := modelInfo.Config
	// 获取供应商 base_url 和 api_key
	providerInfo := h.Redis.FindProviderByName(req.Provider)
	if providerInfo == nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	providerBaseUrl := providerInfo.BaseURL
	providerKey, idx := slice.Random(providerInfo.APIKeys)
	if idx == -1 {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}

	// 获取会话配置
	var session models.Session
	if err := h.Store.Db.First(&session, "id = ?", uri.SessionId).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusNotFound, "session not found")
		return
	}

	// 上下文消息
	enableContext := session.EnableContext // 默认使用会话配置
	if req.EnableContext != nil {
		enableContext = *req.EnableContext // 请求参数优先
	}
	var contextMessages []models.Message
	if enableContext {
		messages, err := h.Store.GetLatestMessages(session.ID, 50)
		if err != nil {
			ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to load context")
			return
		}
		contextMessages = messages
	}

	// 系统提示
	var fullSystemPrompt = ""
	if modelConfig.AllowSystemPrompt {
		var systemPrompt = ""
		if req.SystemPrompt != nil && *req.SystemPrompt != "" {
			systemPrompt = *req.SystemPrompt
		}
		const titlePrompt = "当检测到对话主题发生明显变化时，用简短的标题总结主题。生成的标题应不超过十个字，并用 [title:总结出的标题] 的格式放置在响应开头。如果主题没有变化，则正常回应用户问题。"
		fullSystemPrompt = systemPrompt + titlePrompt
	}

	// 标准格式消息列表
	var chatMessages []chat_utils.Message
	chatMessages = append(
		chatMessages, slice.Map(
			contextMessages, func(_ int, m models.Message) chat_utils.Message {
				return chat_utils.Message{
					Role:    m.Role,
					Content: m.Content,
				}
			},
		)...,
	)
	chatMessages = append(chatMessages, chat_utils.UserMessage(req.Question))

	// 预先插入新对话，获取消息 ID
	messages := []models.Message{
		{SessionID: session.ID, Role: "user", ModelID: modelInfo.ID},
		{SessionID: session.ID, Role: "assistant", ModelID: modelInfo.ID},
	}
	if err := h.Store.CreateMessages(&messages); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create messages")
		return
	}
	// 执行结束后，根据是否有回答进行操作
	var fullResponseContent string
	var fullReasoningContent string
	defer func() {
		if fullResponseContent != "" {
			// 完成响应，记录消息
			messages[0].Content = req.Question
			messages[1].Content = fullResponseContent
			messages[1].ReasoningContent = fullReasoningContent
			if err := h.Store.SaveMessages(&messages); err != nil {
				// do nothing
			}
		} else {
			// 无响应，删除预插入的消息
			if err := h.Store.DeleteMessages(session.ID, []uint64{messages[0].ID, messages[1].ID}); err != nil {
				// do nothing
			}
		}
	}()

	eventChan, err := chat_utils.CompletionStream(
		c.Request.Context(), chat_utils.CompletionStreamOptions{
			Provider: chat_utils.Provider{
				BaseUrl: providerBaseUrl,
				ApiKey:  providerKey.Key,
			},
			Model:                 req.ModelName,
			Messages:              chatMessages,
			SystemPrompt:          fullSystemPrompt,
			CompletionModelConfig: getCompletionModelConfig(modelConfig),
		},
	)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	sendStreamMessageEvent(
		c,
		"[ID:"+strconv.FormatUint(messages[0].ID, 10)+","+strconv.FormatUint(messages[1].ID, 10)+"]",
		false,
	)

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 流式输出
	c.Stream(
		func(w io.Writer) bool {
			event, ok := <-eventChan
			if !ok {
				return false
			}

			switch event.Type {
			case "content":
				// 消息内容
				sendStreamMessageEvent(c, event.Content, false)
			case "reasoning_content":
				// 思考内容
				sendStreamMessageEvent(c, event.Content, true)
			case "error":
				// 错误信息（记录日志并终止流）
				c.SSEvent(
					"error", (&entities.CommonResponse[any]{}).WithError(event.Error).WithCode(500),
				)
				return false
			case "done":
				// 用量信息
				c.SSEvent("usage", event.Metadata)
				resp, ok := event.Metadata.(chat_utils.DoneResponse)
				if ok {
					fullResponseContent = resp.Content
					fullReasoningContent = resp.ReasoningContent
				}
				// 结束标记
				c.SSEvent("done", "[DONE]")
				return false
			}
			return true
		},
	)
}

func sendStreamMessageEvent(c *gin.Context, msg string, thinking bool) {
	var name string
	if thinking {
		name = "think"
	} else {
		name = "msg"
	}
	c.SSEvent(
		name, gin.H{
			"content": strings.ReplaceAll(msg, "\n", "\\n"),
		},
	)
}

func getCompletionModelConfig(config models.ModelConfig) chat_utils.CompletionModelConfig {
	temperature := config.DefaultTemperature
	if temperature == 0 {
		temperature = models.DefaultModelConfig.DefaultTemperature
	}
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = models.DefaultModelConfig.MaxTokens
	}

	return chat_utils.CompletionModelConfig{
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
}
