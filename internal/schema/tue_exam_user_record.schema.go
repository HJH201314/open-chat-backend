package schema

// ScoreStatus 评分状态
type ScoreStatus string

const (
	StatusPending   ScoreStatus = "pending"   // 待评分
	StatusScoring   ScoreStatus = "scoring"   // 评分中
	StatusCompleted ScoreStatus = "completed" // 评分完成
	StatusFailed    ScoreStatus = "failed"    // 评分失败
)

// ExamUserRecordAnswer 用户单题答案
type ExamUserRecordAnswer struct {
	RecordID  uint64      `gorm:"primaryKey" json:"exam_id"`               // 测验 ID
	ProblemID uint64      `gorm:"primaryKey" json:"problem_id"`            // 题目 ID
	Answer    interface{} `gorm:"type:json;serializer:json" json:"answer"` // 用户答案
	Score     uint64      `gorm:"default:0" json:"score"`                  // 得分（单位：0.01分）
	Comments  string      `json:"comments"`                                // 评语/反馈
	Status    ScoreStatus `gorm:"default:pending" json:"status"`           // 评分状态
}

func (e *ExamUserRecordAnswer) TableName() string {
	return "tue_exam_user_records_answers"
}

// ExamUserRecord 考试用户提交记录
type ExamUserRecord struct {
	// 普通字段
	ID         uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64      `gorm:"index;not null" json:"user_id"`                             // 用户ID
	ExamID     uint64      `gorm:"index;not null" json:"exam_id"`                             // 考试ID
	TotalScore uint64      `gorm:"default:0" json:"total_score"`                              // 总得分（单位：0.01分）
	Status     ScoreStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // 总评分状态
	TimeSpent  int         `gorm:"default:0" json:"time_spent"`                               // 答题用时（单位：秒）

	AutoCreateUpdateDeleteAt
	// 组装字段
	Exam    Exam                   `gorm:"foreignKey:ID;references:ExamID" json:"exam"`      // 考试信息
	Answers []ExamUserRecordAnswer `gorm:"foreignKey:RecordID;references:ID" json:"answers"` // 用户答案（关联表储存）
}

func (e *ExamUserRecord) TableName() string {
	return "tue_exam_user_records"
}
