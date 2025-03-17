package gorm

import (
	"github.com/fcraft/open-chat/internal/schema"
	"gorm.io/gorm"
)

// GetExamWithDetails 获取测验详情
func (s *GormStore) GetExamWithDetails(examId uint64) (*schema.Exam, error) {
	var exam schema.Exam
	result := s.Db.
		Preload("ExamProblems.Problem"). // 加载 Sections 及其 Questions 和关联的 Problem
		Preload(
			"ExamProblems", func(db *gorm.DB) *gorm.DB {
				return db.Where("exam_id = ?", examId).Order("sort_order ASC")
			},
		). // 确保所有 Sections 被加载
		First(&exam, examId)

	if result.Error != nil {
		return nil, result.Error
	}
	return &exam, nil
}
