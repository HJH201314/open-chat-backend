package models

import (
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID string         `gorm:"index" json:"session_id"`
	Role      string         `json:"role"` // user/assistant/system
	Content   string         `json:"content"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 软删除
}
