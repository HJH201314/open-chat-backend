package gorm_utils

import (
	"github.com/fcraft/open-chat/internal/entity"
	"gorm.io/gorm"
)

// GetByID 通用获取实体方法
//
//	Parameters:
//		db - 数据库连接
//		id - 实体 ID
func GetByID[T any](db *gorm.DB, id uint64) (*T, error) {
	var result T
	err := db.Where("id = ?", id).First(&result).Error
	return &result, err
}

// Save 通用保存（完全保存）实体方法
//
//	Parameters:
//		db - 数据库连接
//		entity - 实体
func Save[T any](db *gorm.DB, entity *T) error {
	err := db.Save(&entity).Error
	return err
}

// Update 通用更新（部分保存）实体方法
//
//	Parameters:
//		db - 数据库连接
//		entity - 实体
func Update[T any](db *gorm.DB, entity *T) error {
	err := db.Updates(&entity).Error
	return err
}

// GetByPageContinuous 分页获取列表
//
//	Parameters:
//		db - 数据库连接
//		param - 分页参数
//		sort - 排序参数
//
//	Returns:
//		[]T - 实体列表
//		*int - 下一页页码
//		error - 错误信息
func GetByPageContinuous[T any](db *gorm.DB, param entity.PagingParam, sort entity.SortParam) ([]T, *int64, error) {
	var results []T
	pageNum, pageSize := param.GetPageSize(20, 100)
	offset := (pageNum - 1) * pageSize
	// 多查询一条以判断是否存在下一页
	limit := pageSize + 1

	tx := db
	if sort.SafeExpr([]string{}) != "" {
		tx = db.Order(sort.SafeExpr([]string{}))
	}
	err := tx.
		Offset(offset).
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, nil, err
	}

	// 分页逻辑处理
	hasNext := len(results) > pageSize
	if hasNext {
		results = results[:pageSize]
		nextPage := int64(pageNum) + 1
		return results, &nextPage, nil
	}

	return results, nil, nil
}

// GetByPageTotal 分页获取题目列表
//
//	Parameters:
//		db - 数据库连接
//		param - 分页参数
//		sort - 排序参数
//
//	Returns:
//		[]T - 实体列表
//		*int - 最后一页页码
//		error - 错误信息
func GetByPageTotal[T any](db *gorm.DB, param entity.PagingParam, sort entity.SortParam) ([]T, *int64, error) {
	var results []T
	pageNum, pageSize := param.GetPageSize(20, 100)
	offset := (pageNum - 1) * pageSize
	limit := pageSize

	// 查询 total
	var total int64
	var model T

	if err := db.Model(&model).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	tx := db
	if sort.SafeExpr([]string{}) != "" {
		tx = db.Order(sort.SafeExpr([]string{}))
	}
	if err := tx.
		Offset(offset).
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, nil, err
	}

	// 分页逻辑处理
	lastPage := (total + int64(limit) - 1) / int64(limit) // 向上取整
	return results, &lastPage, nil
}
