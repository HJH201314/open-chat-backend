package schema

// Exam 考试模型
type Exam struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Description string        `json:"description"`                       // 考试描述
	TotalScore  uint64        `json:"total_score"`                       // 考试总分（单位：0.01分）
	Subjects    string        `gorm:"type:varchar(100)" json:"subjects"` // 所属科目分类
	LimitTime   int           `gorm:"default:0" json:"limit_time"`       // 考试限时（单位：秒）
	Problems    []ExamProblem `gorm:"foreignKey:ExamID" json:"problems"` // 考试包含的大题
	AutoCreateUpdateDeleteAt
}

func (e *Exam) TableName() string {
	return "tue_exams"
}

// ExamProblem 大题题目关联模型
type ExamProblem struct {
	ExamID    uint64  `gorm:"primaryKey;autoIncrement:false" json:"exam_id"`    // 关联考试ID
	ProblemID uint64  `gorm:"primaryKey;autoIncrement:false" json:"problem_id"` // 关联题目ID
	Problem   Problem `gorm:"foreignKey:ProblemID" json:"problem"`              // 题目详细信息
	Score     uint64  `gorm:"not null;default:0" json:"score"`                  // 题目分值（1表示0.01分）
	SortOrder int     `gorm:"not null;default:0" json:"sort_order"`             // 题目排序
}

func (e *ExamProblem) TableName() string {
	return "tue_exam_problems"
}
