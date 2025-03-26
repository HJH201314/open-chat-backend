package chat

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	_ "github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

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
		BotID         *uint64 `json:"bot_id" binding:"-"`
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
	// 验证用户对会话的所有权
	if !h.Helper.CheckUserSession(ctx_utils.GetUserId(c), uri.SessionId) {
		ctx_utils.BizError(c, constants.ErrNoPermission)
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
	var session schema.Session
	if err := h.Db.First(&session, "id = ?", uri.SessionId).Error; err != nil {
		ctx_utils.CustomError(c, http.StatusNotFound, "session not found")
		return
	}

	// 获取 bot 的提示词会话
	var bot *schema.Preset
	if req.BotID != nil && *req.BotID > 0 {
		botRole, err := h.Helper.GetPreset(*req.BotID)
		if err != nil || botRole == nil || botRole.PromptSession == nil {
			ctx_utils.CustomError(c, http.StatusNotFound, "bot role not found")
			return
		}
		bot = botRole
	}

	// 上下文消息
	enableContext := session.EnableContext // 默认使用会话配置
	if req.EnableContext != nil {
		enableContext = *req.EnableContext // 请求参数优先
	}
	if bot != nil {
		enableContext = bot.PromptSession.EnableContext // bot 配置优先
	}
	var contextMessages []schema.Message
	if enableContext {
		contextSize := 50
		if bot != nil && bot.PromptSession.ContextSize > 0 {
			// bot 上下文窗口配置优先
			contextSize = bot.PromptSession.ContextSize
		}
		messages, err := h.Store.GetLatestMessages(session.ID, contextSize)
		if err != nil {
			ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to load context")
			return
		}
		contextMessages = messages
	}

	// 系统提示
	var systemPrompt = ""
	if session.SystemPrompt != "" {
		systemPrompt = session.SystemPrompt
	}
	if modelConfig.AllowSystemPrompt {
		if req.SystemPrompt != nil && *req.SystemPrompt != "" {
			systemPrompt = *req.SystemPrompt
		}
	}
	if bot != nil {
		// bot 提示词优先覆盖
		systemPrompt = bot.PromptSession.SystemPrompt
	}

	// 标准格式消息列表
	var chatMessages []chat_utils.Message
	// 此处不加入系统提示，系统提示由工具函数内部处理
	// 标准格式消息 - bot 提示词消息
	if bot != nil && bot.PromptSession != nil {
		chatMessages = append(
			chatMessages, slice.Map(
				bot.PromptSession.Messages, func(_ int, m schema.Message) chat_utils.Message {
					return chat_utils.Message{
						Role:    m.Role,
						Content: m.Content,
					}
				},
			)...,
		)
	}
	// 标准格式消息 - 上下文消息
	chatMessages = append(
		chatMessages, slice.Map(
			contextMessages, func(_ int, m schema.Message) chat_utils.Message {
				return chat_utils.Message{
					Role:    m.Role,
					Content: m.Content,
				}
			},
		)...,
	)
	// 标准格式消息 - 用户输入
	chatMessages = append(chatMessages, chat_utils.UserMessage(req.Question))

	// 预先插入新对话，获取消息 ID
	messages := []schema.Message{
		{SessionID: session.ID, Role: "user", ModelID: modelInfo.ID},
		{SessionID: session.ID, Role: "assistant", ModelID: modelInfo.ID},
	}
	if err := h.Store.CreateMessages(&messages); err != nil {
		ctx_utils.CustomError(c, http.StatusInternalServerError, "failed to create messages")
		return
	}
	// 执行结束后，根据是否有回答进行操作
	var doneResp *chat_utils.DoneResponse
	defer func() {
		if doneResp != nil && doneResp.Content != "" {
			// 完成响应，记录消息
			messages[0].Content = req.Question
			messages[0].TokenUsage = doneResp.Usage.PromptTokens
			messages[1].Content = doneResp.Content
			messages[1].ReasoningContent = doneResp.ReasoningContent
			messages[1].TokenUsage = doneResp.Usage.CompletionTokens
			if err := h.Store.SaveMessages(&messages); err != nil {
				// do nothing
			}
			usageTokens := doneResp.Usage.PromptTokens + doneResp.Usage.CompletionTokens*4
			if err := h.Store.UpdateUserUsage(ctx_utils.GetUserId(c), -usageTokens); err != nil {
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
		c.Request.Context(), chat_utils.CompletionOptions{
			Provider: chat_utils.Provider{
				BaseUrl: providerBaseUrl,
				ApiKey:  providerKey.Key,
			},
			Model:                 req.ModelName,
			Messages:              chatMessages,
			SystemPrompt:          systemPrompt,
			CompletionModelConfig: getCompletionModelConfig(modelConfig),
		},
	)
	if err != nil {
		ctx_utils.HttpError(c, constants.ErrInternal)
		return
	}
	sendStreamCommandEvent(
		c, "ID", map[string]uint64{
			"q": messages[0].ID,
			"a": messages[1].ID,
		},
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
			case chat_utils.ContentEventType:
				// 消息内容
				sendStreamMessageEvent(c, event.Content, false)
				return true
			case chat_utils.ReasoningContentEventType:
				// 思考内容
				sendStreamMessageEvent(c, event.Content, true)
				return true
			case chat_utils.ErrorEventType:
				// 错误信息（记录日志并终止流）
				c.SSEvent(
					"error", (&entity.CommonResponse[any]{}).WithError(event.Error).WithCode(500),
				)
				return false
			case chat_utils.DoneEventType:
				// 用量信息
				c.SSEvent("usage", event.Metadata)
				resp, ok := event.Metadata.(chat_utils.DoneResponse)
				if ok {
					doneResp = &resp
				}
				// 结束标记
				c.SSEvent("done", "[DONE]")
				return false
			default:
				return true
			}
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
			"content": msg,
		},
	)
}

// SendStreamCommandEvent 发送流式命令事件
func sendStreamCommandEvent(c *gin.Context, cmd string, data interface{}) {
	dataString, ok := data.(string)
	if ok {
		// 字符串格式
		c.SSEvent("cmd", fmt.Sprintf("[%s,%s]", cmd, dataString))
	} else {
		// JSON 格式
		c.SSEvent(
			"cmd", gin.H{
				"name": cmd,
				"data": data,
			},
		)
	}
}

func getCompletionModelConfig(config schema.ModelConfig) chat_utils.CompletionModelConfig {
	temperature := config.DefaultTemperature
	if temperature == 0 {
		temperature = schema.DefaultModelConfig.DefaultTemperature
	}
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = schema.DefaultModelConfig.MaxTokens
	}

	return chat_utils.CompletionModelConfig{
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
}
