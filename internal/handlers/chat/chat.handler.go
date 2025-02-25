package chat

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/handlers"
	"github.com/fcraft/open-chat/internal/models"
	"github.com/fcraft/open-chat/internal/shared/constant"
	"github.com/fcraft/open-chat/internal/shared/entity"
	_ "github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/fcraft/open-chat/internal/shared/util"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
)

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
//	@Success		200	{object}	entity.CommonResponse[string]
//	@Router			/chat/session/new [post]
func (h *Handler) CreateSession(c *gin.Context) {
	session := models.Session{
		UserID:        util.GetUserId(c),
		EnableContext: true, // 默认开启上下文
	}
	if err := h.Store.CreateSession(&session); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create session")
		return
	}
	util.NormalResponse(c, session.ID)
}

// DeleteSession
//
//	@Summary		删除会话
//	@Description	删除会话
//	@Tags			Session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string	true	"会话 ID"
//	@Success		200			{object}	entity.CommonResponse[bool]
//	@Router			/chat/session/del/{session_id} [post]
func (h *Handler) DeleteSession(c *gin.Context) {
	var uri struct {
		SessionId string `uri:"session_id" binding:"required"`
	}
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	// 验证用户对会话的所有权
	var session models.Session
	if err := h.Store.Db.Where(
		"id = ? AND user_id = ?",
		uri.SessionId,
		util.GetUserId(c),
	).First(&session).Error; err != nil {
		util.HttpErrorResponse(c, constant.ErrUnauthorized)
		return
	}
	// 执行删除操作
	if err := h.Store.DeleteSession(uri.SessionId); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to delete session")
		return
	}
	util.NormalResponse(c, true)
}

// GetModels
//
//	@Summary		获取所有模型
//	@Description	获取所有模型
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	entity.CommonResponse[[]models.ModelCache]
//	@Router			/chat/config/models [get]
func (h *Handler) GetModels(c *gin.Context) {
	// 从缓存中查询
	cacheModels, err := h.Redis.GetCachedModels()
	if err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to get models")
		return
	}
	// 将 config 隐藏
	slice.ForEach(
		cacheModels, func(_ int, item models.ModelCache) {
			item.Config = models.ModelConfig{}
		},
	)
	util.NormalResponse(c, cacheModels)
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
	// 从 query 和 body 中获取用户输入
	var uri struct {
		SessionId string `uri:"session_id" binding:"required"`
	}
	type userInput struct {
		Question      string  `json:"question" binding:"required"`
		Provider      string  `json:"provider_name" binding:"required"` // Provider.Name 准确的供应商名称
		ModelName     string  `json:"model_name" binding:"required"`    // Model.Name 准确的模型名称
		EnableContext *bool   `json:"enable_context" binding:"-"`
		SystemPrompt  *string `json:"system_prompt" binding:"-"` // 系统提示词
	}
	var req userInput
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Question == "" {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}

	// 读取模型信息
	modelInfo := h.Redis.FindCachedModelByName(req.Provider, req.ModelName)
	if modelInfo == nil {
		util.HttpErrorResponse(c, constant.ErrBadRequest)
		return
	}
	modelConfig := modelInfo.Config
	// 获取供应商 base_url 和 api_key
	providerInfo := h.Redis.FindProviderByName(req.Provider)
	if providerInfo == nil {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}
	providerBaseUrl := providerInfo.BaseURL
	providerKey, idx := slice.Random(providerInfo.APIKeys)
	if idx == -1 {
		util.HttpErrorResponse(c, constant.ErrInternal)
		return
	}

	// 获取会话配置
	var session models.Session
	if err := h.Store.Db.First(&session, "id = ?", uri.SessionId).Error; err != nil {
		util.CustomErrorResponse(c, http.StatusNotFound, "session not found")
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
			util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to load context")
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
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create messages")
		return
	}
	// 执行结束后，根据是否有回答进行操作
	var fullResponseContent string
	defer func() {
		if fullResponseContent != "" {
			// 完成响应，记录消息
			messages[0].Content = req.Question
			messages[1].Content = fullResponseContent
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
		util.HttpErrorResponse(c, constant.ErrInternal)
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
					"error", (&entity.CommonResponse[any]{}).WithError(event.Error).WithCode(500),
				)
				return false
			case "done":
				// 用量信息
				c.SSEvent("usage", event.Metadata)
				resp, ok := event.Metadata.(chat_utils.DoneResponse)
				if ok {
					fullResponseContent = resp.Content
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
