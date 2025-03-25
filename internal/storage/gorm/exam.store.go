package gorm

import "github.com/fcraft/open-chat/internal/schema"

// UpdateExamTotalScore 更新测验总分
func (s *GormStore) UpdateExamTotalScore(examId uint64) error {
	var totalScore uint64
	err := s.Db.Model(&schema.ExamProblem{}).
		Where("exam_id = ?", examId).
		Select("sum(score)").
		Row().
		Scan(&totalScore)
	if err != nil {
		return err
	}
	return s.Db.Model(&schema.Exam{}).Where("id = ?", examId).Update("total_score", totalScore).Error
}
