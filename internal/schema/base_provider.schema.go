package schema

type Provider struct {
	ID          uint64   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string   `gorm:"not null;unique" json:"name"`           // 提供商名称
	DisplayName string   `gorm:"" json:"display_name"`                  // 对外展示提供商名称
	BaseURL     string   `gorm:"not null" json:"base_url"`              // API 的基本 URL
	Description string   `gorm:"" json:"description"`                   // 额外提供商描述
	APIKeys     []APIKey `gorm:"foreignKey:ProviderID" json:"api_keys"` // 一对多关系，与 APIKey 模型关联
	Models      []Model  `gorm:"foreignKey:ProviderID" json:"models"`   // 一对多关系，与 Model 模型关联
	AutoCreateUpdateAt
}

type APIKey struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProviderID uint64 `gorm:"index;not null" json:"provider_id"` // 外键，指向 Provider
	Key        string `gorm:"not null" json:"key"`               // API 密钥
	AutoCreateAt
}

type Model struct {
	// 原始数据
	ID          uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProviderID  uint64      `gorm:"index;not null" json:"provider_id"`       // 关联的 Provider ID
	Name        string      `gorm:"not null" json:"name"`                    // 模型名称
	DisplayName string      `gorm:"" json:"display_name"`                    // 对外展示模型名称
	Description string      `gorm:"" json:"description"`                     // 额外模型描述
	Config      ModelConfig `gorm:"type:json;serializer:json" json:"config"` // 使用 JSON 储存配置
	Active      bool        `gorm:"default:true" json:"active"`              // 是否启用

	// 组装数据
	Provider *Provider `gorm:"foreignKey:ProviderID" json:"provider"`
	AutoCreateUpdateAt
}

// ModelCache 缓存中的 schema 数据
type ModelCache struct {
	Model
	ProviderName        string `json:"provider_name"`         // 关联的 Provider FileName
	ProviderDisplayName string `json:"provider_display_name"` // 关联的 Provider DisplayName
}

// ModelConfig 定义了模型的默认配置
type ModelConfig struct {
	DefaultTemperature float64 `json:"default_temperature"` // 默认温度
	AllowSystemPrompt  bool    `json:"allow_system_prompt"` // 是否允许用户自行修改系统提示
	SystemPrompt       string  `json:"system_prompt"`       // 预设系统提示
	MaxTokens          int64   `json:"max_tokens"`
	TopP               float32 `json:"top_p"`
	FrequencyPenalty   float32 `json:"frequency_penalty"`
	PresencePenalty    float32 `json:"presence_penalty"`
}

var DefaultModelConfig = ModelConfig{
	DefaultTemperature: 0.6,
	AllowSystemPrompt:  true,
	SystemPrompt:       "",
	MaxTokens:          4096,
	TopP:               1.0,
	FrequencyPenalty:   0.0,
	PresencePenalty:    0.0,
}

// ModelCollection 是多个相似模型的集合，用于负载均衡/统一名称，这在对接多个供应商服务的时候很有用
type ModelCollection struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"not null;unique" json:"name"`                       // 唯一标识名称
	DisplayName string  `json:"display_name"`                                      // 展示名称
	Description string  `json:"description"`                                       // 额外描述
	Models      []Model `gorm:"many2many:model_collections_models;" json:"models"` // 关联的模型
	AutoCreateAt
}
