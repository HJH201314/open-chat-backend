package chat_utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"strings"
)

// CompletionStream 流式聊天
func CompletionStream(ctx context.Context, opts CompletionStreamOptions) (chan StreamEvent, error) {
	// 参数校验
	if err := validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// 初始化客户端
	client := openai.NewClient(option.WithBaseURL(opts.Provider.BaseUrl), option.WithAPIKey(opts.Provider.ApiKey))

	// 构建请求消息
	messages := buildMessages(opts)

	// 创建事件通道（带缓冲避免阻塞）
	eventChan := make(chan StreamEvent, 5)

	// 启动协程处理流式请求
	go processStreaming(ctx, client, messages, opts, eventChan)

	return eventChan, nil
}

// 参数校验逻辑
func validateOptions(opts CompletionStreamOptions) error {
	if len(opts.Messages) == 0 {
		return errors.New("messages cannot be empty")
	}
	if opts.Temperature < 0 || opts.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	return nil
}

// 消息预处理
func buildMessages(opts CompletionStreamOptions) []Message {
	messages := make([]Message, 0, len(opts.Messages)+1)

	// 添加系统提示
	if opts.SystemPrompt != "" {
		messages = append(
			messages, Message{
				Role:    "system",
				Content: opts.SystemPrompt,
			},
		)
	}

	// 添加上下文消息
	messages = append(messages, opts.Messages...)
	return messages
}

// 流式处理核心逻辑
func processStreaming(ctx context.Context, client *openai.Client, messages []Message, opts CompletionStreamOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan) // 确保通道关闭

	// 将消息转换为 OpenAI 的请求格式，ChatCompletionMessage 是 ChatCompletionMessageParamUnion 的特例
	reqMessages := slice.Map(
		messages, func(_ int, m Message) openai.ChatCompletionMessageParamUnion {
			return openai.ChatCompletionMessage{
				Role:    openai.ChatCompletionMessageRole(m.Role),
				Content: m.Content,
			}
		},
	)
	// 获取提供商的流式响应
	stream := client.Chat.Completions.NewStreaming(
		ctx, openai.ChatCompletionNewParams{
			Messages:    openai.F(reqMessages),
			Model:       openai.F(opts.Model),
			Temperature: openai.F(opts.Temperature),
			MaxTokens:   openai.F(opts.MaxTokens),
		},
	)
	if stream.Err() != nil {
		sendError(eventChan, fmt.Errorf("failed to create stream: %w", stream.Err()))
		return
	}
	defer func(stream *ssestream.Stream[openai.ChatCompletionChunk]) {
		err := stream.Close()
		if err != nil {

		}
	}(stream)

	acc := openai.ChatCompletionAccumulator{}

	// 处理流式响应
streamingLoop:
	for {
		select {
		case <-ctx.Done():
			sendError(eventChan, ctx.Err())
			return
		default:
			// 接收下一个响应，若流式结束，退出循环
			if !stream.Next() {
				break streamingLoop
			}
			if stream.Err() != nil {
				sendError(eventChan, fmt.Errorf("stream error: %w", stream.Err()))
				return
			}
			chunk := stream.Current()
			acc.AddChunk(chunk)

			// 额外解析含有 reasoning_content 的结构
			choiceDelta := &chunk.Choices[0].Delta
			// 发送内容事件
			if reasoningContent := choiceDelta.JSON.ExtraFields["reasoning_content"].Raw(); reasoningContent != "" {
				reasoningContent, _ = strings.CutPrefix(reasoningContent, "\"")
				reasoningContent, _ = strings.CutSuffix(reasoningContent, "\"")
				eventChan <- StreamEvent{
					Type:    "reasoning_content",
					Content: reasoningContent,
				}
			} else {
				eventChan <- StreamEvent{
					Type:    "content",
					Content: choiceDelta.Content,
				}
			}
		}
	}

	// 发送完成事件
	eventChan <- StreamEvent{
		Type: "done", Metadata: DoneResponse{
			Content: acc.Choices[0].Message.Content,
			Usage: DoneResponseUsage{
				PromptTokens:     acc.Usage.PromptTokens,
				CompletionTokens: acc.Usage.CompletionTokens,
			},
		},
	}
}

// 发送错误事件
func sendError(ch chan<- StreamEvent, err error) {
	ch <- StreamEvent{
		Type:  "error",
		Error: err,
	}
}

// Provider 提供商信息
type Provider struct {
	BaseUrl string
	ApiKey  string
}

// Message 消息结构体
type Message struct {
	Role    string
	Content string
}

func SystemMessage(content string) Message {
	return Message{
		Role:    "system",
		Content: content,
	}
}
func AssistantMessage(content string) Message {
	return Message{
		Role:    "assistant",
		Content: content,
	}
}
func UserMessage(content string) Message {
	return Message{
		Role:    "user",
		Content: content,
	}
}

// StreamEvent 表示流式事件的数据结构
type StreamEvent struct {
	Type     string      // 事件类型：content/error/done
	Content  string      // 内容（当 Type=content 时有效）
	Error    error       // 错误对象（当 Type=error 时有效）
	Metadata interface{} // 附加元数据
}

// CompletionStreamOptions 流式请求配置
type CompletionStreamOptions struct {
	Provider     Provider  // 服务提供商
	Model        string    // 模型名称
	Messages     []Message // 消息列表
	SystemPrompt string    // 系统提示词
	CompletionModelConfig
}

type CompletionModelConfig struct {
	Temperature   float64 // 温度系数
	MaxTokens     int64   // 最大 token 数
	ContextWindow int64   // 上下文窗口大小
}

// DoneResponse 结果响应
type DoneResponse struct {
	Content string            `json:"content"`
	Usage   DoneResponseUsage `json:"usage"`
}
type DoneResponseUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}

// ChoiceDelta 拓展流式输出 Delta
type ChoiceDelta struct {
	openai.ChatCompletionChunkChoicesDelta
	ReasoningContent string `json:"reasoning_content"`
}
