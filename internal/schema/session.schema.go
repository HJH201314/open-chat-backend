package schema

import (
	"time"
)

type Session struct {
	// 原始数据
	ID            string      `gorm:"default:gen_random_uuid();foreignKey:SessionID" json:"id"`
	UserID        uint64      `gorm:"index" json:"user_id"`
	LastActive    time.Time   `json:"last_active"`
	EnableContext bool        `json:"enable_context"`                                     // 上下文开关
	ModelParams   ModelParams `gorm:"embedded;embeddedPrefix:param_" json:"model_params"` // 模型参数
	AutoCreateDeleteAt

	// 组装数据
	Messages []Message `json:"messages"`
}

type ModelParams struct {
	Model       string  `json:"schema"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}
