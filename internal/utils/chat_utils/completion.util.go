package chat_utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
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
	params := openai.ChatCompletionNewParams{
		Messages: reqMessages,
		Model:    opts.Model,
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

	// 发送请求
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	toolCalls := completion.Choices[0].Message.ToolCalls

	if len(toolCalls) != 0 {
		// If there is a was a function call, continue the conversation
		params.Messages = append(params.Messages, completion.Choices[0].Message.ToParam())
		for _, toolCall := range toolCalls {
			if toolCall.Function.Name == "gen_single_choice_problem" {
				// Extract the location from the function call arguments
				var args map[string]interface{}
				err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				if err != nil {
					panic(err)
				}
				location := args["topic"].(string)

				questionData, err := opts.Tools[0].Handler(location)

				fmt.Printf("gened question in %s\n", questionData)
			}
		}
	}

	// 构建响应
	if len(completion.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	// 提取 reasoning_content
	reasoningContent := ""
	if _, exists := completion.Choices[0].Message.JSON.ExtraFields["reasoning_content"]; exists {
		reasoningContent = completion.Choices[0].Message.JSON.ExtraFields["reasoning_content"].Raw()
	}

	return &CompletionResponse{
		Content:          completion.Choices[0].Message.Content,
		ReasoningContent: reasoningContent,
		Usage: CompletionUsage{
			PromptTokens:     completion.Usage.PromptTokens,
			CompletionTokens: completion.Usage.CompletionTokens,
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
