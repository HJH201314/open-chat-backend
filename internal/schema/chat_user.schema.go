package schema

type User struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Roles    []Role `gorm:"many2many:user_roles;" json:"roles"` // 用户与角色之间的多对多关系
	AutoCreateUpdateDeleteAt
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
