package schema

type Message struct {
	// 默认结构
	ID               uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID        string `gorm:"index" json:"session_id"`
	Role             string `json:"role"`        // user/assistant/system
	ModelID          uint64 `json:"model_id"`    // 回复所使用的模型
	BotRoleID        uint64 `json:"bot_role_id"` // 回复所使用的模型
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
	TokenUsage       int64  `gorm:"default:0" json:"token_usage"`
	AutoCreateDeleteAt

	// 组装结构
	Model   *Model   `gorm:"foreignKey:ID;references:ModelID" json:"model"`
	BotRole *BotRole `gorm:"foreignKey:ID;references:BotRoleID" json:"bot_role"`
}

func (m *Message) TableName() string {
	return "messages"
}
