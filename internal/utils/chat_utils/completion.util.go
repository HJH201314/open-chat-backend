package chat_utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Completion 非流式聊天完成
func Completion(ctx context.Context, opts CompletionOptions) (*CompletionResponse, error) {
	// 参数校验
	if err := validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// 初始化客户端
	client := openai.NewClient(option.WithBaseURL(opts.Provider.BaseUrl), option.WithAPIKey(opts.Provider.ApiKey))

	// 构建请求消息
	messages := buildMessages(opts)

	// 构建请求参数
	reqMessages := slice.Map(
		messages, func(_ int, m Message) openai.ChatCompletionMessageParamUnion {
			return openai.ChatCompletionMessage{
				Role:    openai.ChatCompletionMessageRole(m.Role),
				Content: m.Content,
			}
		},
	)

	// 发送请求
	resp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages:    openai.F(reqMessages),
			Model:       openai.F(opts.Model),
			Temperature: openai.F(opts.Temperature),
			MaxTokens:   openai.F(opts.MaxTokens),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	// 构建响应
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	// 提取 reasoning_content
	reasoningContent := ""
	if _, exists := resp.Choices[0].Message.JSON.ExtraFields["reasoning_content"]; exists {
		reasoningContent = resp.Choices[0].Message.JSON.ExtraFields["reasoning_content"].Raw()
	}

	return &CompletionResponse{
		Content:          resp.Choices[0].Message.Content,
		ReasoningContent: reasoningContent,
		Usage: CompletionUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}

// CompletionResponse 非流式响应结构
type CompletionResponse struct {
	Content          string          `json:"content"`
	ReasoningContent string          `json:"reasoning_content"`
	Usage            CompletionUsage `json:"usage"`
}

// CompletionUsage 使用统计
type CompletionUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}
