package schema

import (
	"encoding/json"
	"time"
)

// Session 会话，一系列消息的集合
type Session struct {
	// 原始数据
	ID            string          `gorm:"primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string          `json:"name"`
	NameType      SessionNameType `gorm:"default:0" json:"name_type"` // 标题来源
	EnableContext bool            `json:"enable_context"`             // 上下文开关
	ContextSize   int             `json:"context_size"`               // 上下文大小
	SystemPrompt  string          `json:"system_prompt"`              // 系统提示词
	LastActive    time.Time       `json:"last_active"`
	AutoCreateUpdateDeleteAt

	// 组装数据
	Messages []Message `gorm:"foreignKey:SessionID;references:ID" json:"messages"`
}

type SessionNameType int

const (
	SessionNameTypeNone SessionNameType = iota
	SessionNameTypeTemp
	SessionNameTypeSystem
)

func (s *Session) TableName() string {
	return "sessions"
}

type UserSessionType int

const (
	UserSessionTypeOwner UserSessionType = iota + 1
	UserSessionTypeInvitee
)

func (u UserSessionType) MarshalJSON() ([]byte, error) {
	var str string
	switch u {
	case UserSessionTypeOwner:
		str = "owner"
	case UserSessionTypeInvitee:
		str = "invitee"
	}
	return json.Marshal(str)
}

// UserSession 用户-会话
type UserSession struct {
	// 原始数据
	UserID    uint64           `gorm:"primaryKey;index" json:"user_id"`
	SessionID string           `gorm:"primaryKey;index" json:"session_id"`
	Type      UserSessionType  `json:"type"`
	ShareInfo SessionShareInfo `gorm:"embedded;embeddedPrefix:share_" json:"share_info"` // 分享字段
	FlagInfo  SessionFlagInfo  `gorm:"embedded;embeddedPrefix:flag_" json:"flag_info"`

	AutoCreateUpdateDeleteAt

	// 组装数据
	Session *Session `gorm:"foreignKey:ID;references:SessionID" json:"session"`
}

type SessionShareInfo struct {
	Permanent bool   `gorm:"default:false" json:"permanent"`                          // 是否永久分享
	Title     string `json:"title"`                                                   // 分享标题
	Code      string `gorm:"type:varchar(32)" json:"code,omitempty"`                  // 邀请码（可选）
	ExpiredAt int64  `gorm:"index;type:time;serializer:unixmstime" json:"expired_at"` // 邀请过期时间
}

type SessionFlagInfo struct {
	Star bool `gorm:"default:false" json:"star"` // 标星
}

func (UserSession) TableName() string {
	return "sessions_users"
}
