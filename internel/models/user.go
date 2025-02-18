package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `json:"username"`
	Password  string         `json:"-"`
	Roles     []Role         `gorm:"many2many:user_roles;" json:"roles"` // 用户与角色之间的多对多关系
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Permission struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"unique;not null" json:"name"` // 权限名称
	Description string         `json:"description"`                 // 权限描述
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Role struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"unique;not null" json:"name"`                    // 角色名称
	Description string         `json:"description"`                                    // 角色描述
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions"` // 多对多关联
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type RolePermission struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint64    `gorm:"not null" json:"role_id"`
	PermissionID uint64    `gorm:"not null" json:"permission_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions" // 显式声明表名
}

type UserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	RoleID    uint64    `gorm:"not null" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRole) TableName() string {
	return "user_roles" // 显式声明表名
}
