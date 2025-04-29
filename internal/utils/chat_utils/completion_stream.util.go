package chat_utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"log/slog"
	"strings"
)

// CompletionStream 流式聊天
func CompletionStream(ctx context.Context, opts CompletionOptions, eventChan chan<- StreamEvent) error {
	// 参数校验
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// 初始化客户端
	client := openai.NewClient(option.WithBaseURL(opts.Provider.BaseUrl), option.WithAPIKey(opts.Provider.ApiKey))

	// 构建请求消息
	messages := buildMessages(opts)

	// 启动协程处理流式请求
	go processStreaming(ctx, &client, messages, opts, eventChan)

	return nil
}

// 参数校验逻辑
func validateOptions(opts CompletionOptions) error {
	if len(opts.Messages) == 0 {
		return errors.New("messages cannot be empty")
	}
	if opts.Temperature < 0 || opts.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	return nil
}

// 消息预处理
func buildMessages(opts CompletionOptions) []Message {
	messages := make([]Message, 0, len(opts.Messages)+1)

	// 添加系统提示
	if opts.SystemPrompt != "" && (len(opts.Messages) <= 0 || opts.Messages[0].Role != "system") {
		messages = append(
			messages, Message{
				Role:    "system",
				Content: opts.SystemPrompt,
			},
		)
	}

	// 添加上下文消息
	messages = append(messages, opts.Messages...)
	slog.Default().Info("messages: ", messages)
	return messages
}

// 流式处理核心逻辑
func processStreaming(ctx context.Context, client *openai.Client, messages []Message, opts CompletionOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan) // 确保通道关闭

	// 将消息转换为 OpenAI 的请求格式，ChatCompletionMessage 是 ChatCompletionMessageParamUnion 的特例
	reqMessages := slice.Map(
		messages, func(_ int, m Message) openai.ChatCompletionMessageParamUnion {
			if m.Role == "user" {
				return openai.UserMessage(m.Content)
			} else {
				return openai.AssistantMessage(m.Content)
			}
		},
	)

	availableTools := slice.Map(
		opts.Tools, func(_ int, tool CompletionTool) openai.ChatCompletionToolParam {
			return tool.Param
		},
	)
	toolsMap := ConvertToolsToMap(opts.Tools)

	// 构造参数，非必要不传递
	params := openai.ChatCompletionNewParams{
		Messages: reqMessages,
		Model:    opts.Model,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Opt(true),
		},
	}
	if opts.Temperature > 0 {
		params.Temperature = openai.Opt(opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		params.MaxTokens = openai.Opt(opts.MaxTokens)
	}
	if len(opts.Tools) > 0 {
		params.Tools = availableTools
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.Opt("auto"),
		}
	}

	// 获取提供商的流式响应
	stream := client.Chat.Completions.NewStreaming(
		ctx, params,
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
	accReasoningContent := ""

	replaceMsg := "" // 用于在输出结果为空时作为结果，通常在 tool_calls 时使用
	extra := map[string]any{}

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

			// 检测到一个完整的 tool call
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason == "tool_calls" && len(acc.ChatCompletion.Choices[0].Message.ToolCalls) > 0 {
				toolcall := acc.Choices[0].Message.ToolCalls[0]
				tool, toolOk := toolsMap[toolcall.Function.Name]
				if toolOk {
					if tool.UserTip != "" {
						eventChan <- StreamEvent{
							Type:    CommandEventType,
							Content: "tooltip",
							Metadata: map[string]string{
								"tooltip": tool.UserTip,
							},
						}
					}
					res, err := tool.Handler(toolcall.Function.Arguments)
					if err != nil || res == nil {
						eventChan <- StreamEvent{
							Type:    ErrorEventType,
							Content: "tool_error",
							Error:   err,
						}
						return
					}
					replaceMsg = res.ReplaceMessage
					if acc.Choices[0].Message.Content == "" {
						// 发送 msg
						eventChan <- StreamEvent{
							Type:    ContentEventType,
							Content: replaceMsg,
						}
					}
					// 把函数处理结果存入 extra，可能被用于存入数据库
					extra[res.Type] = res.Data
					// 发送 cmd
					eventChan <- StreamEvent{
						Type:     CommandEventType,
						Content:  "tool:" + res.Type,
						Metadata: res.Data,
					}
				}
				println(
					"Tool call stream finished:",
					toolcall.ID,
					toolcall.Type,
					toolcall.Function.Name,
					toolcall.Function.Arguments,
				)
			}

			// 额外解析含有 reasoning_content 的结构
			if len(chunk.Choices) == 0 {
				continue
			}
			choiceDelta := &chunk.Choices[0].Delta
			// 发送内容事件
			if reasoningContent := choiceDelta.JSON.ExtraFields["reasoning_content"].Raw(); reasoningContent != "" && reasoningContent != "null" {
				reasoningContent, _ = strings.CutPrefix(reasoningContent, "\"")
				reasoningContent, _ = strings.CutSuffix(reasoningContent, "\"")
				accReasoningContent += reasoningContent
				eventChan <- StreamEvent{
					Type:    ReasoningContentEventType,
					Content: reasoningContent,
				}
			} else {
				eventChan <- StreamEvent{
					Type:    ContentEventType,
					Content: choiceDelta.Content,
				}
			}
		}
	}

	var finalContent string
	if len(acc.Choices) > 0 && acc.Choices[0].Message.Content != "" {
		finalContent = acc.Choices[0].Message.Content
	} else {
		finalContent = replaceMsg
	}
	// 发送完成事件
	eventChan <- StreamEvent{
		Type: DoneEventType, Metadata: DoneResponse{
			Content:          finalContent,
			ReasoningContent: accReasoningContent,
			Extra:            extra,
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
		Type:  ErrorEventType,
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

type StreamEventType string

var (
	ReasoningContentEventType StreamEventType = "reasoning_content"
	ContentEventType          StreamEventType = "content"
	ErrorEventType            StreamEventType = "error"
	DoneEventType             StreamEventType = "done"
	CommandEventType          StreamEventType = "command"
)

// StreamEvent 表示流式事件的数据结构
type StreamEvent struct {
	Type     StreamEventType // 事件类型：content/error/done
	Content  string          // 内容（当 Type=content 时有效）
	Error    error           // 错误对象（当 Type=error 时有效）
	Metadata interface{}     // 附加元数据
}

// CompletionOptions 流式请求配置
type CompletionOptions struct {
	Provider     Provider  // 服务提供商
	Model        string    // 模型名称
	Messages     []Message // 消息列表
	SystemPrompt string    // 系统提示词
	CompletionModelConfig

	Tools []CompletionTool // 工具列表
}

type CompletionToolHandlerReturn struct {
	Data           interface{}
	ReplaceMessage string // 调用工具可能模型没有回复，此时使用替代回复
	Type           string // TODO: 支持区分回发 command / 继续请求对话 等等
}

type CompletionTool struct {
	Param   openai.ChatCompletionToolParam
	UserTip string // 调用工具时给用户输出的提示
	Handler func(args ...interface{}) (*CompletionToolHandlerReturn, error)
}

type CompletionModelConfig struct {
	Temperature float64 // 温度系数
	MaxTokens   int64   // 最大 token 数
}

// DoneResponse 结果响应
type DoneResponse struct {
	Content          string            `json:"content"`
	ReasoningContent string            `json:"reasoning_content"`
	Extra            map[string]any    `json:"extra"`
	Usage            DoneResponseUsage `json:"usage"`
}
type DoneResponseUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}
