package schema

// CourseSortedDataItem 定义了可排序数据的接口
type CourseSortedDataItem interface {
	GetType() string
	GetSortOrder() int
}

// SortedResource 排序后的资源数据
type SortedResource struct {
	Type      string   `json:"type"`
	SortOrder int      `json:"sort_order"`
	Resource  Resource `json:"resource"`
}

// SortedExam 排序后的考试数据
type SortedExam struct {
	Type      string `json:"type"`
	SortOrder int    `json:"sort_order"`
	Exam      Exam   `json:"exam"`
}

func (sr *SortedResource) GetType() string {
	return "resource"
}

func (sr *SortedResource) GetSortOrder() int {
	return sr.SortOrder
}

func (se *SortedExam) GetType() string {
	return "exam"
}

func (se *SortedExam) GetSortOrder() int {
	return se.SortOrder
}

type Course struct {
	// 原始数据
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"` // 课程名称
	Description string `json:"description"`                            // 课程描述
	CourseRelatedData
	AutoCreateUpdateDeleteAt
}

type CourseRelatedData struct {
	// 关联数据
	Resources []CourseResource `gorm:"foreignKey:CourseID" json:"resources"` // 课程资源
	Exams     []CourseExam     `gorm:"foreignKey:CourseID" json:"exams"`     // 课程考试
	// 排好序的数据
	SortedData []CourseSortedDataItem `gorm:"-" json:"sorted_data"` // 使用 gorm:"-" 标记为非数据库字段
}

// CourseResource 课程-资源关联模型
type CourseResource struct {
	ID         uint64   `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID   uint64   `gorm:"not null" json:"course_id"`             // 关联课程ID
	ResourceID uint64   `gorm:"not null" json:"resource_id"`           // 关联资源ID
	SortOrder  int      `gorm:"not null;default:0" json:"sort_order"`  // 资源排序
	Resource   Resource `gorm:"foreignKey:ResourceID" json:"resource"` // 资源详细信息
}

// CourseExam 课程-考试关联模型
type CourseExam struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseID  uint64 `gorm:"not null" json:"course_id"`            // 关联课程ID
	ExamID    uint64 `gorm:"not null" json:"exam_id"`              // 关联考试ID
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"` // 考试排序
	Exam      Exam   `gorm:"foreignKey:ExamID" json:"exam"`        // 考试详细信息
}

func (c *Course) TableName() string {
	return "tue_courses"
}

func (cr *CourseResource) TableName() string {
	return "tue_course_resources"
}

func (ce *CourseExam) TableName() string {
	return "tue_course_exams"
}
