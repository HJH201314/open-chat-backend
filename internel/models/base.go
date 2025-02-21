package models

import (
	"gorm.io/gorm"
	"time"
)

type AutoCreateAt struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type AutoCreateUpdateAt struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type AutoCreateDeleteAt struct {
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 软删除
}

type AutoCreateUpdateDeleteAt struct {
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 软删除
}
