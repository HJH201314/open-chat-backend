package schema

type Bucket struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	DisplayName     string `gorm:"not null" json:"display_name"`
	EndpointURL     string `gorm:"not null" json:"endpoint_url"`
	Region          string `gorm:"not null" json:"region"`
	AccessKeyID     string `gorm:"not null" json:"access_key_id"`
	SecretAccessKey string `gorm:"not null" json:"secret_access_key"`
	BucketName      string `gorm:"not null" json:"bucket_name"`
	AutoCreateUpdateDeleteAt
}

type File struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketID uint64 `gorm:"not null" json:"bucket_id"`
	Name     string `gorm:"not null" json:"name"`               // 文件名
	Size     int64  `gorm:"not null" json:"size"`               // 文件大小（字节）
	Type     string `gorm:"not null" json:"type"`               // 文件类型（如 image/jpeg）
	Module   string `gorm:"not null" json:"module"`             // 文件所属模块
	S3Path   string `gorm:"not null;unique" json:"s3_path"`     // S3 存储路径（如 "uploads/abc123.jpg"）
	OwnerID  uint64 `gorm:"not null;default:0" json:"owner_id"` // 文件所有者ID（可选）
	AutoCreateDeleteAt

	Bucket *Bucket `gorm:"foreignKey:ID;references:BucketID" json:"bucket"`
}
