package gorm

import (
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
)

// GetCourseWithDetails 获取课程详情
func (s *GormStore) GetCourseWithDetails(courseId uint64) (*schema.Course, error) {
	var course schema.Course
	result := s.Db.
		Preload("Resources.Resource"). // 加载课程资源及其详细信息
		Preload("Exams.Exam").         // 加载课程考试及其详细信息
		Preload(
			"Resources", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC")
			},
		). // 按排序加载资源
		Preload(
			"Exams", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC")
			},
		). // 按排序加载考试
		First(&course, courseId)

	if result.Error != nil {
		return nil, result.Error
	}

	// 构建排序数据
	course.SortedData = make([]schema.CourseSortedDataItem, 0)

	// 添加资源
	for _, resource := range course.Resources {
		course.SortedData = append(
			course.SortedData, &schema.SortedResource{
				Type:      "resource",
				SortOrder: resource.SortOrder,
				Resource:  resource.Resource,
			},
		)
	}

	// 添加考试
	for _, exam := range course.Exams {
		course.SortedData = append(
			course.SortedData, &schema.SortedExam{
				Type:      "exam",
				SortOrder: exam.SortOrder,
				Exam:      exam.Exam,
			},
		)
	}

	// 按照 sort_order 排序
	sort.Slice(
		course.SortedData, func(i, j int) bool {
			return course.SortedData[i].GetSortOrder() < course.SortedData[j].GetSortOrder()
		},
	)

	return &course, nil
}

// GetCourses 获取课程列表
func (s *GormStore) GetCourses(pageParam entity.PagingParam, sortParam entity.SortParam) ([]schema.Course, int64, error) {
	// 构建查询
	query := s.Db.
		Preload("Resources.Resource").
		Preload("Exams.Exam").
		Preload(
			"Exams.Exam.Problems", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC")
			},
		).
		Preload("Exams.Exam.Problems.Problem")

	// 使用 GetByPageTotal 获取分页数据
	courses, total, err := gorm_utils.GetByPageTotal[schema.Course](query, pageParam, sortParam)
	if err != nil {
		return nil, 0, err
	}

	// 为每个课程构建排序数据
	for i := range courses {
		courses[i].SortedData = make([]schema.CourseSortedDataItem, 0)

		// 添加资源
		for _, resource := range courses[i].Resources {
			courses[i].SortedData = append(
				courses[i].SortedData, &schema.SortedResource{
					Type:      "resource",
					SortOrder: resource.SortOrder,
					Resource:  resource.Resource,
				},
			)
		}

		// 添加考试
		for _, exam := range courses[i].Exams {
			courses[i].SortedData = append(
				courses[i].SortedData, &schema.SortedExam{
					Type:      "exam",
					SortOrder: exam.SortOrder,
					Exam:      exam.Exam,
				},
			)
		}

		// 按照 sort_order 排序
		sort.Slice(
			courses[i].SortedData, func(j, k int) bool {
				return courses[i].SortedData[j].GetSortOrder() < courses[i].SortedData[k].GetSortOrder()
			},
		)
	}

	return courses, total, nil
}

// CreateCourse 全新创建课程
func (s *GormStore) CreateCourse(course *schema.Course) error {
	return s.Db.Create(course).Error
}

// UpdateCourse 完全更新课程
func (s *GormStore) UpdateCourse(course *schema.Course) error {
	return s.Db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(course).Error
}

// DeleteCourse 完全删除课程
func (s *GormStore) DeleteCourse(courseId uint64) error {
	return s.Db.Select(clause.Associations).Delete(&schema.Course{ID: courseId}).Error
}

// GetExamWithDetails 获取测验详情
func (s *GormStore) GetExamWithDetails(examId uint64) (*schema.Exam, error) {
	var exam schema.Exam
	result := s.Db.
		Preload("Problems.Problem"). // 加载 Sections 及其 Questions 和关联的 Problem
		Preload(
			"Problems", func(db *gorm.DB) *gorm.DB {
				return db.Where("exam_id = ?", examId).Order("sort_order ASC")
			},
		). // 确保所有 Sections 被加载
		First(&exam, examId)

	if result.Error != nil {
		return nil, result.Error
	}
	return &exam, nil
}
