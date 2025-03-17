package schema

type UserUsage struct {
	ID     uint64 `gorm:"primaryKey" json:"id" binding:"-"`
	UserID uint64 `gorm:"unique" json:"user_id" binding:"required"`
	Token  int64  `gorm:"default:0" json:"token" binding:"required"`
	AutoCreateAt
}
