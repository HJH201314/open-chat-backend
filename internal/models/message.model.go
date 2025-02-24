package models

type Message struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID string `gorm:"index" json:"session_id"`
	Role      string `json:"role"` // user/assistant/system
	Content   string `json:"content"`
	AutoCreateDeleteAt
}
