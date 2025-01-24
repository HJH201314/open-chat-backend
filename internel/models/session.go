package models

import (
	"time"
)

type Session struct {
	ID            string      `gorm:"default:gen_random_uuid()" json:"id"`
	UserID        string      `gorm:"index" json:"user_id"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	LastActive    time.Time   `json:"last_active"`
	EnableContext bool        `json:"enable_context"`               // 上下文开关
	ModelParams   ModelParams `gorm:"embedded" json:"model_params"` // 模型参数
}

type ModelParams struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}
