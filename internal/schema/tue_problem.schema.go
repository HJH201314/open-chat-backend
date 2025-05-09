package schema

// ProblemType 题目类型枚举
type ProblemType string

const (
	SingleChoice   ProblemType = "single_choice"
	MultipleChoice ProblemType = "multiple_choice"
	FillBlank      ProblemType = "fill_blank"
	ShortAnswer    ProblemType = "short_answer"
	TrueFalse      ProblemType = "true_false"
	AnyProblemType ProblemType = "any" // 任意，仅内部使用
)

// ProblemOption 题目选项结构
type ProblemOption struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	Correct bool   `json:"correct"` // 是否正确答案
}

// ProblemAnswer 题目答案结构（JSON格式存储），多加这层包装否则储存不了 true/false
type ProblemAnswer struct {
	// 选择题：存储正确选项ID []uint
	// 填空题：存储多个填空关键词 []string
	// 判断题：true/false
	// 简答题：文本答案 string
	Answer any `json:"answer"`
}

// Problem 题目模型
type Problem struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        ProblemType     `gorm:"type:varchar(20);index;not null" json:"type"`
	Description string          `gorm:"type:text;not null" json:"description"`                       // 支持HTML/Markdown
	Options     []ProblemOption `gorm:"type:json;serializer:json" json:"options"`                    // 选项（JSON存储ProblemOption数组）
	Answer      ProblemAnswer   `gorm:"type:json;serializer:json;not null" json:"answer"`            // 答案（JSON存储ProblemAnswer）
	Explanation string          `gorm:"type:text" json:"explanation"`                                // 答案解析
	Difficulty  int             `gorm:"type:int;default:3" json:"difficulty" validate:"min=1,max=5"` // 难度等级 1-5
	Subject     string          `gorm:"type:varchar(100)" json:"subject"`                            // 所属科目/分类

	AutoCreateUpdateDeleteAt
}

func (p *Problem) TableName() string {
	return "tue_problems"
}

// ProblemUserRecord 用户答题记录
type ProblemUserRecord struct {
	ID        uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID uint64        `gorm:"index;not null" json:"problem_id"`
	UserID    uint64        `gorm:"index;not null" json:"user_id"`
	Answer    ProblemAnswer `gorm:"type:json;serializer:json;not null" json:"answer"`
	Score     uint64        `gorm:"not null;default:0" json:"score"`
	Comment   string        `json:"comment"`
	AutoCreateAt
	// 组装
	Problem Problem `json:"problem"`
}

func (r *ProblemUserRecord) TableName() string {
	return "tue_problem_user_records"
}

// ProblemMakeRecord 出题记录
type ProblemMakeRecord struct {
	ID                       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PresetCompletionRecordID uint64 `gorm:"index;not null" json:"preset_completion_record_id"`
	ProblemID                uint64 `gorm:"index;not null" json:"problem_id"`

	AutoCreateUpdateAt
}

func (r *ProblemMakeRecord) TableName() string {
	return "tue_problem_make_records"
}
