package chat_utils

import (
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/fcraft/open-chat/internal/schema"
)

// GetCommonCompletionOptions returns the common completion options for the given provider model.
func GetCommonCompletionOptions(providerModel *schema.Model, options *CompletionOptions) CompletionOptions {
	completionOptions := CompletionOptions{}
	err := convertor.CopyProperties(&completionOptions, options)
	if err != nil || providerModel.Provider == nil || len(providerModel.Provider.APIKeys) < 1 {
		return CompletionOptions{}
	}
	completionOptions.Provider = Provider{
		BaseUrl: providerModel.Provider.BaseURL,
		ApiKey:  providerModel.Provider.APIKeys[0].Key,
	}
	completionOptions.Model = providerModel.Name
	return completionOptions
}

// ConvertMessagesToSchema 将 chat_utils.Message 转换为 schema.Message
//
//	Parameters:
//		- chatMessages: chat_utils.Message 数组
//		- args: args[0]应为 map[string]string 类型的模板数据
//	Returns:
//		- schema.Message 数组
func ConvertMessagesToSchema(chatMessages []Message, args ...interface{}) []schema.Message {
	schemaMessages := make([]schema.Message, 0)
	for _, chatMessage := range chatMessages {
		// 从参数中取出模板数据
		var templateData map[string]string
		if len(args) > 0 {
			if template, ok := args[0].(map[string]string); ok {
				templateData = template
			}
		}
		schemaMessage := schema.Message{
			Role:    chatMessage.Role,
			Content: strutil.TemplateReplace(chatMessage.Content, templateData),
		}
		schemaMessages = append(schemaMessages, schemaMessage)
	}
	return schemaMessages
}

// ConvertSchemaToMessages 将 schema.Message 转换为 chat_utils.Message
//
//	Parameters:
//		- schemaMessages: schema.Message 数组
//		- args: args[0]应为 map[string]string 类型的模板数据
//	Returns:
//		- chat_utils.Message 数组
func ConvertSchemaToMessages(schemaMessages []schema.Message, args ...interface{}) []Message {
	chatMessages := make([]Message, 0)
	for _, schemaMessage := range schemaMessages {
		// 从参数中取出模板数据
		var templateData map[string]string
		if len(args) > 0 {
			if template, ok := args[0].(map[string]string); ok {
				templateData = template
			}
		}
		chatMessage := Message{
			Role:    schemaMessage.Role,
			Content: strutil.TemplateReplace(schemaMessage.Content, templateData),
		}
		chatMessages = append(chatMessages, chatMessage)
	}
	return chatMessages
}

func ConvertToolsToMap(tools []CompletionTool) map[string]CompletionTool {
	toolMap := make(map[string]CompletionTool)
	for _, tool := range tools {
		toolMap[tool.Param.Function.Name] = tool
	}
	return toolMap
}
