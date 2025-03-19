package schema

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
