package chat

import (
	"context"
	"github.com/fcraft/open-chat/internel/handlers"
	"github.com/fcraft/open-chat/internel/models"
	"github.com/fcraft/open-chat/internel/shared/constant"
	"github.com/fcraft/open-chat/internel/shared/util"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewChatHandler(h *handlers.BaseHandler) *Handler {
	return &Handler{BaseHandler: h}
}

// CreateSession 创建会话
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

// DeleteSession 删除会话
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
	if err := h.Store.Db.Where("id = ? AND user_id = ?", uri.SessionId, util.GetUserId(c)).First(&session).Error; err != nil {
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

// CompletionStream 流式输出聊天
func (h *Handler) CompletionStream(c *gin.Context) {
	// 设置流式响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")

	// 从 query 和 body 中获取用户输入
	var uri struct {
		SessionId string `uri:"session_id" binding:"required"`
	}
	var request struct {
		Question      string  `json:"question" binding:"required"`
		EnableContext *bool   `json:"enable_context" binding:"-"`
		Provider      *string `json:"provider" binding:"-"`      // DeepSeek or OpenAI
		ModelName     *string `json:"model_name" binding:"-"`    // 准确的模型名称
		SystemPrompt  *string `json:"system_prompt" binding:"-"` // 系统提示词
	}
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		_ = c.Error(constant.ErrBadRequest)
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Question == "" {
		_ = c.Error(constant.ErrBadRequest)
		return
	}

	// 获取会话配置
	var session models.Session
	if err := h.Store.Db.First(&session, "id = ?", uri.SessionId).Error; err != nil {
		util.CustomErrorResponse(c, http.StatusNotFound, "session not found")
		return
	}

	// 若启用（会话配置或显式传入），获取上下文消息
	var contextMessages []models.Message
	if (request.EnableContext == nil && session.EnableContext) || (request.EnableContext != nil && *request.EnableContext) {
		messages, err := h.Store.GetLatestMessages(session.ID, 50)
		if err != nil {
			util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to load context")
			return
		}
		contextMessages = messages
	}
	var chatMessages []openai.ChatCompletionMessageParamUnion
	// 系统提示
	/*var systemPrompt = ""
	if request.SystemPrompt != nil && *request.SystemPrompt != "" {
		systemPrompt = *request.SystemPrompt
	} else {
		systemPrompt = ""
	}
	const titlePrompt = "当检测到对话主题发生明显变化时，用简短的标题总结主题。生成的标题应不超过十个字，并用 [title:总结出的标题] 的格式放置在响应开头。如果主题没有变化，则正常回应用户问题。"
	fullSystemPrompt := systemPrompt + titlePrompt
	chatMessages = append(chatMessages, openai.ChatCompletionMessage{Role: "system", Content: fullSystemPrompt})*/
	for _, msg := range contextMessages {
		switch msg.Role {
		case "user":
			chatMessages = append(chatMessages, openai.ChatCompletionMessage{Role: "user", Content: msg.Content})
		case "assistant":
			chatMessages = append(chatMessages, openai.ChatCompletionMessage{Role: "assistant", Content: msg.Content})
		}
	}
	chatMessages = append(chatMessages, openai.UserMessage(request.Question))

	// 预先插入新对话，获取消息 ID
	messages := []models.Message{
		{SessionID: session.ID, Role: "user"},
		{SessionID: session.ID, Role: "assistant"},
	}
	if err := h.Store.CreateMessages(&messages); err != nil {
		util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to create messages")
		return
	}

	// 初始化 OpenAI 客户端
	var client *openai.Client
	var modelName string
	if request.Provider != nil && *request.Provider == "OpenAI" {
		client = openai.NewClient(option.WithAPIKey(os.Getenv("API_KEY_GPT")), option.WithBaseURL("https://api.chatanywhere.tech"))
		modelName = "gpt-4o"
	} else {
		client = openai.NewClient(option.WithAPIKey(os.Getenv("API_KEY_DEEPSEEK")), option.WithBaseURL("https://api.deepseek.com"))
		modelName = "deepseek-chat"
	}
	if request.ModelName != nil {
		modelName = *request.ModelName
	}

	// 创建流式请求
	stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
		Messages:    openai.F(chatMessages),
		Model:       openai.F(modelName),
		Temperature: openai.F(0.6),
	})

	// 创建一个通道来发送事件
	eventChan := make(chan string)

	acc := openai.ChatCompletionAccumulator{}

	go func() {
		// 发送消息 ID 给前端
		eventChan <- "[ID:" + strconv.FormatUint(messages[0].ID, 10) + "," + strconv.FormatUint(messages[1].ID, 10) + "]"

		// 流式传输 OpenAI 的响应
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			// 将 OpenAI 的响应通过通道发送到 Gin 的响应流
			if len(chunk.Choices) > 0 {
				eventChan <- chunk.Choices[0].Delta.Content
			}
		}

		if err := stream.Err(); err != nil {
			eventChan <- "[ERROR: API response error]"
		}
		// 清理资源，1. 发送[DONE]告知前端响应已完成 2. 关闭通道以结束当前连接 3. 关闭 OpenAI 的数据流
		if err := stream.Close(); err != nil {
			return
		}
		eventChan <- "[DONE]"
		close(eventChan)

		// 保存用户输入和响应结果
		if len(acc.Choices) > 0 {
			messages[0].Content = request.Question
			messages[1].Content = acc.Choices[0].Message.Content
			if err := h.Store.SaveMessages(&messages); err != nil {
				return
			}
		} else {
			// 如果没有响应，删除预先保存的消息
			if err := h.Store.DeleteMessages(session.ID, []uint64{messages[0].ID, messages[1].ID}); err != nil {
				return
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		if event, ok := <-eventChan; ok {
			// 显式传输换行符，避免前端处理异常
			event = strings.ReplaceAll(event, "\n", "\\n")
			// 按照 SSE 规范输出内容
			_, err := w.Write([]byte("data: " + event + "\n\n"))

			if err != nil {
				return false
			}
			// 返回 true 说明还要等待下一个事件，Stream 会进入下一次迭代
			return true
		}
		return false
	})
}
