package schema

import (
	"encoding/json"
	"time"
)

// Session 会话，一系列消息的集合
type Session struct {
	// 原始数据g
	ID            string    `gorm:"primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string    `json:"name"`
	EnableContext bool      `json:"enable_context"` // 上下文开关
	SystemPrompt  string    `json:"system_prompt"`  // 系统提示词
	LastActive    time.Time `json:"last_active"`
	AutoCreateDeleteAt

	// 组装数据
	Messages []Message `gorm:"foreignKey:SessionID;references:ID" json:"messages"`
}

type UserSessionType int

const (
	OWNER UserSessionType = iota + 1
	INVITEE
)

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

func (UserSession) TableName() string {
	return "sessions_users"
}
