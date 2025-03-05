package schema

import (
	"encoding/json"
	"time"
)

type Session struct {
	// 原始数据
	ID            string      `gorm:"primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string      `json:"title"`
	LastActive    time.Time   `json:"last_active"`
	EnableContext bool        `json:"enable_context"`                                     // 上下文开关
	ModelParams   ModelParams `gorm:"embedded;embeddedPrefix:param_" json:"model_params"` // 模型参数
	AutoCreateDeleteAt

	// 组装数据
	Messages []Message `gorm:"foreignKey:SessionID;references:ID" json:"messages"`
}

type ModelParams struct {
	Model       string  `json:"schema"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type UserSessionType int

func (u UserSessionType) MarshalJSON() ([]byte, error) {
	var str string
	switch u {
	case OWNER:
		str = "owner"
	case INVITEE:
		str = "invitee"
	}
	return json.Marshal(str)
}

const (
	OWNER UserSessionType = iota + 1
	INVITEE
)

// UserSession 用户-会话
type UserSession struct {
	// 原始数据
	ID        uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64          `gorm:"index" json:"user_id"`
	SessionID string          `gorm:"index" json:"session_id"`
	Type      UserSessionType `json:"type"`
	AutoCreateAt

	// 组装数据
	Session *Session `gorm:"foreignKey:ID;references:SessionID" json:"session"`
}
