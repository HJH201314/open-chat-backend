package services

import (
	"encoding/json"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"gorm.io/datatypes"
	"sync"
)

type ChatService struct {
	TitleGenerateTasks []string
	BaseService
}

func (s *ChatService) GenerateTitleForSession(sessionID string, startMessageIndex int, messageCount int) error {
	// 1. 获取对话数据
	var schemaMessages []schema.Message
	if err := s.Gorm.Where(
		"session_id = ?",
		sessionID,
	).Find(&schemaMessages).Offset(startMessageIndex).Limit(messageCount).Order("id ASC").Error; err != nil {
		return err
	}

	// 2. 将对话数据转换为合适的格式
	messages := chat_utils.ConvertSchemaToMessages(schemaMessages)
	if len(messages) == 0 {
		return nil
	}
	strMessages, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	// 3. 执行预设
	title, _, err := BuiltinPresetCompletion(
		ChatSessionTitleGeneratePresetName,
		map[string]string{
			"CONTENT": string(strMessages),
		},
	)
	if err != nil {
		return err
	}

	// 4. 更新对话标题
	return s.Gorm.Model(&schema.Session{}).Where(
		"id = ?",
		sessionID,
	).Updates(
		map[string]any{
			"name":      chat_utils.ExtractTagContent(title, "title"),
			"name_type": schema.SessionNameTypeSystem,
		},
	).Error
}

var (
	chatServiceInstance *ChatService
	chatServiceOnce     sync.Once
)

// InitChatService 初始化对话服务
func InitChatService(base *BaseService) *ChatService {
	chatServiceOnce.Do(
		func() {
			chatServiceInstance = &ChatService{
				BaseService: *base,
			}
			registerSystemConfig()
			registerBuiltinPreset()
		},
	)
	return chatServiceInstance
}

const (
	ChatOnlineSearchServiceBaseURL      = "chat_online_search_searxng_service"
	ChatSessionTitleGeneratePresetName  = "chat_session_title_generate"
	ChatSearchKeywordGeneratePresetName = "chat_search_keyword_generate"
)

func registerSystemConfig() {
	err := GetSystemConfigService().RegisterSystemConfig(
		RegisterConfigParams{
			Name:        ChatOnlineSearchServiceBaseURL,
			DisplayName: "联网搜索 SearXNG 服务",
			Schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":        "string",
					"description": "service base url",
				},
			},
			Default:  datatypes.NewJSONType[any]([]string{""}),
			IsPublic: false,
		},
	)
	if err != nil {
		return
	}
}

func registerBuiltinPreset() {
	// 对话标题生成
	GetPresetService().RegisterBuiltinPresetsSimple(
		ChatSessionTitleGeneratePresetName, "对话标题生成", 3, "", []chat_utils.Message{
			chat_utils.UserMessage(
				`
你的任务是根据下方对话内容，总结一句 25 字以下的标题，尽量简洁。
对话内容：{CONTENT}
标题语言：内容中的主要自然语言（不包含代码）
标题输出在<title></title>中
`,
			),
		},
	)

	// 提炼搜索词
	GetPresetService().RegisterBuiltinPresetsSimple(
		ChatSearchKeywordGeneratePresetName, "搜索词提炼", 2, "", []chat_utils.Message{
			chat_utils.UserMessage(
				`
你的任务是根据下方用户消息，提炼出用于提交到在线搜索引擎的搜索词。若你认为该问题无需联网搜索，搜索词为空。
对话内容：{CONTENT}
搜索词输出在<search></search>中
`,
			),
		},
	)
}

// GetChatService 获取对话服务
func GetChatService() *ChatService {
	if chatServiceInstance == nil {
		panic("ChatService not initialized")
	}
	return chatServiceInstance
}
