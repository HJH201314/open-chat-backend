package chat

import (
	"context"
	"github.com/fcraft/open-chat/internel/models"
	"github.com/fcraft/open-chat/internel/shared/constant"
	"github.com/fcraft/open-chat/internel/shared/util"
	"github.com/fcraft/open-chat/internel/storage"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"io"
	"net/http"
	"os"
	"strings"
)

type Handler struct {
	store *storage.GormStore
}

func NewChatHandler(store *storage.GormStore) *Handler {
	return &Handler{store: store}
}

// CreateSession 创建会话
func (h *Handler) CreateSession(c *gin.Context) {
	session := models.Session{
		UserID:        util.GetUserId(c),
		EnableContext: true, // 默认开启上下文
	}
	if err := h.store.CreateSession(&session); err != nil {
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
	if err := h.store.Db.Where("id = ? AND user_id = ?", uri.SessionId, util.GetUserId(c)).First(&session).Error; err != nil {
		util.HttpErrorResponse(c, constant.ErrUnauthorized)
		return
	}
	// 执行删除操作
	if err := h.store.DeleteSession(uri.SessionId); err != nil {
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
		Question string `json:"question"`
	}
	if err := c.BindUri(&uri); err != nil || uri.SessionId == "" {
		_ = c.Error(constant.ErrBadRequest)
		return
	}
	if err := c.BindJSON(&request); err != nil || request.Question == "" {
		_ = c.Error(constant.ErrBadRequest)
		return
	}

	// 获取会话配置
	var session models.Session
	if err := h.store.Db.First(&session, "id = ?", uri.SessionId).Error; err != nil {
		util.CustomErrorResponse(c, http.StatusNotFound, "session not found")
		return
	}

	// 获取上下文消息 (当启用时)
	var contextMessages []models.Message
	// 强制关闭上下文
	session.EnableContext = false
	if session.EnableContext {
		messages, err := h.store.GetLatestMessages(session.ID, 50)
		if err != nil {
			util.CustomErrorResponse(c, http.StatusInternalServerError, "failed to load context")
			return
		}
		contextMessages = messages
	}
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range contextMessages {
		switch msg.Role {
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}
	chatMessages = append(chatMessages, openai.UserMessage(request.Question))

	// 初始化 OpenAI 客户端
	client := openai.NewClient(option.WithAPIKey(os.Getenv("API_KEY_DEEPSEEK")), option.WithBaseURL("https://api.deepseek.com"))

	// 创建流式请求
	stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F(chatMessages),
		Model:    openai.F("deepseek-chat"),
	})

	acc := openai.ChatCompletionAccumulator{}

	// 创建一个通道来发送事件
	eventChan := make(chan string)

	go func() {
		// 流式传输 OpenAI 的响应
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			// 将 OpenAI 的响应通过通道发送到 Gin 的响应流
			if len(chunk.Choices) > 0 {
				eventChan <- chunk.Choices[0].Delta.Content
			}
		}

		// 清理资源，1. 发送[DONE]告知前端响应已完成 2. 关闭通道以结束当前连接 3. 关闭 OpenAI 的数据流
		eventChan <- "[DONE]"
		close(eventChan)
		if err := stream.Close(); err != nil {
			return
		}
		if err := stream.Err(); err != nil {
			panic(err)
		}

		// 保存用户输入和响应结果
		if len(acc.Choices) > 0 {
			userMsg := models.Message{
				SessionID: session.ID,
				Role:      "user",
				Content:   request.Question,
			}
			assistantMsg := models.Message{
				SessionID: session.ID,
				Role:      "assistant",
				Content:   acc.Choices[0].Message.Content,
			}
			if err := h.store.SaveMessage(&userMsg); err != nil {
				return
			}
			if err := h.store.SaveMessage(&assistantMsg); err != nil {
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
