package schema

type Permission struct {
	Name        string `gorm:"primaryKey;not null" json:"name"` // 权限名称
	Path        string `gorm:"not null" json:"path"`            // 权限路径（形如：POST:/user/create）
	Description string `json:"description"`                     // 权限描述
	Module      string `json:"module"`                          // 所属模块（handler名称）
	Active      bool   `gorm:"default:true" json:"active"`      // 是否启用（唯一可设置的字段）
	AutoCreateUpdateAt
}

type Role struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"unique;not null" json:"name"` // 角色名称
	DisplayName string `json:"display_name"`                // 角色名称
	Description string `json:"description"`                 // 角色描述
	Active      bool   `gorm:"default:true" json:"active"`  // 是否启用
	AutoCreateUpdateDeleteAt
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permissions"` // 多对多关联
}
