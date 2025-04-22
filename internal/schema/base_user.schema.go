package schema

type UserType string

const (
	UserTypeNormal     UserType = "normal"      // 使用普通注册（后续绑定第三方，类型不变）
	UserTypeThirdParty UserType = "third_party" // 使用第三方登录（后续可通过设置密码来转为普通）
)

type User struct {
	ID       uint64   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Nickname string   `gorm:"varchar(32)" json:"nickname"`
	Type     UserType `gorm:"default:normal" json:"type"`
	Roles    []Role   `gorm:"many2many:user_roles;" json:"roles"` // 用户与角色之间的多对多关系
	AutoCreateUpdateDeleteAt
}

type UserRole struct {
	UserID uint64 `gorm:"primaryKey;" json:"user_id"`
	RoleID uint64 `gorm:"primaryKey;" json:"role_id"`
	AutoCreateAt
}

func (UserRole) TableName() string {
	return "user_roles" // 显式声明表名
}
