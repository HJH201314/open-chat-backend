package models

import (
	"time"
)

type Message struct {
	ID        string    `gorm:"default:gen_random_uuid()" json:"id"`
	SessionID string    `gorm:"index" json:"session_id"`
	Role      string    `json:"role"` // user/assistant/system
	Content   string    `json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
