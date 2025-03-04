package schema

type Message struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID        string `gorm:"index" json:"session_id"`
	Role             string `json:"role"`     // user/assistant/system
	ModelID          uint64 `json:"model_id"` // 回复所使用的模型
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
	AutoCreateDeleteAt
}
