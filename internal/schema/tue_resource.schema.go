package schema

type Resource struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	FileKey        string `gorm:"index;default:gen_random_uuid()" json:"file_key"`    // 文件的 uuid key
	FileName       string `gorm:"type:varchar(255);not null" json:"file_name"`        // OSS 中的文件名
	OriginFileName string `gorm:"type:varchar(255);not null" json:"origin_file_name"` // 原始文件名
	Description    string `json:"description"`                                        // 资源描述

	AutoCreateUpdateDeleteAt
}

func (r *Resource) TableName() string {
	return "tue_resources"
}
