package models

import (
	"time"
)

type Provider struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"not null;unique" json:"name"` // 提供商名称
	BaseURL     string    `gorm:"not null" json:"base_url"`    // API 的基本 URL
	Description string    `gorm:"" json:"description"`         // 额外提供商描述
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	APIKeys     []APIKey  `gorm:"foreignKey:ProviderID" json:"api_keys"` // 一对多关系，与 APIKey 模型关联
	Models      []Model   `gorm:"foreignKey:ProviderID" json:"models"`   // 一对多关系，与 Model 模型关联
}

type APIKey struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Key        string    `gorm:"not null" json:"key"` // API 密钥
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	ProviderID uint      `gorm:"not null" json:"provider_id"` // 外键，指向 Provider
}

type Model struct {
	ID         uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	ProviderID uint        `gorm:"not null" json:"provider_id"` // 关联的 Provider ID
	Name       string      `gorm:"not null" json:"name"`        // 模型名称
	Config     ModelConfig `gorm:"type:json" json:"config"`     // 使用 JSON 储存配置
	CreatedAt  time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

// ModelConfig 定义了模型的默认配置
type ModelConfig struct {
	DefaultTemperature float32 `json:"default_temperature"`
	SystemPrompt       string  `json:"system_prompt"`
	MaxTokens          int     `json:"max_tokens"`
	TopP               float32 `json:"top_p"`
	FrequencyPenalty   float32 `json:"frequency_penalty"`
	PresencePenalty    float32 `json:"presence_penalty"`
}

var DefaultModelConfig = ModelConfig{
	DefaultTemperature: 0.6,
	SystemPrompt:       "",
	MaxTokens:          4096,
	TopP:               1.0,
	FrequencyPenalty:   0.0,
	PresencePenalty:    0.0,
}
