package chat_utils

import (
	"github.com/duke-git/lancet/v2/convertor"
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
