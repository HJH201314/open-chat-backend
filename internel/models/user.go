package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint64         `gorm:"autoIncrement" json:"id"`
	Username  string         `json:"username"`
	Password  string         `json:"-"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
