package schema

// SystemConfig 模型定义
type SystemConfig struct {
	Name        string `gorm:"uniqueIndex;not null"` // 配置标识名
	DisplayName string `gorm:"not null"`             // 显示名称
	Schema      string `gorm:"type:json;not null"`   // JSON Schema格式
	Default     string `gorm:"type:json"`            // 配置默认值(JSON格式)
	Value       string `gorm:"type:json"`            // 配置值(JSON格式)
	Description string `gorm:"type:text"`            // 配置描述
	IsPublic    bool   `gorm:"default:false"`        // 是否公开配置

	AutoCreateUpdateDeleteAt
}
