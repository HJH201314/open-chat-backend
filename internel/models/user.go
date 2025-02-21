package models

type User struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Roles    []Role `gorm:"many2many:user_roles;" json:"roles"` // 用户与角色之间的多对多关系
	AutoCreateUpdateDeleteAt
}

type Permission struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"unique;not null" json:"name"` // 权限名称
	Path        string `gorm:"not null" json:"path"`        // 权限路径（一般与名称相同）
	Description string `json:"description"`                 // 权限描述
	AutoCreateUpdateDeleteAt
}

type Role struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name"`                    // 角色名称
	Description string       `json:"description"`                                    // 角色描述
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"` // 多对多关联
	AutoCreateUpdateDeleteAt
}

type RolePermission struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint64 `gorm:"not null" json:"role_id"`
	PermissionID uint64 `gorm:"not null" json:"permission_id"`
	AutoCreateAt
}

func (RolePermission) TableName() string {
	return "role_permissions" // 显式声明表名
}

type UserRole struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID uint64 `gorm:"not null" json:"user_id"`
	RoleID uint64 `gorm:"not null" json:"role_id"`
	AutoCreateAt
}

func (UserRole) TableName() string {
	return "user_roles" // 显式声明表名
}
