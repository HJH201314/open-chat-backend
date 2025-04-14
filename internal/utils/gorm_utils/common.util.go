package gorm_utils

import (
	"errors"
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

// GetByName 通用获取实体方法
//
//	Parameters:
//		db - 数据库连接
//		name - 实体 name
func GetByName[T any](db *gorm.DB, name string) (*T, error) {
	var result T
	err := db.Where("name = ?", name).First(&result).Error
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

// Delete 通用删除实体方法
//
//	Parameters:
//		db - 数据库连接
//		id - ID / 实体
func Delete[T any](db *gorm.DB, entityOrId interface{}) error {
	switch entityOrId.(type) {
	case T:
	case *T:
		return db.Delete(entityOrId).Error
	case uint64:
		var value T
		return db.Where("id = ?", entityOrId).Delete(&value).Error
	}
	return errors.New("incorrect entityOrId")
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
func GetByPageContinuous[T any](db *gorm.DB, param entity.ParamPagingSort) ([]T, *int64, error) {
	sort := param.SortParam
	paging := param.PagingParam
	timeRange := param.TimeRangeParam

	var results []T
	pageNum, pageSize := paging.GetPage(20, 100)
	offset := (pageNum - 1) * pageSize
	// 多查询一条以判断是否存在下一页
	limit := pageSize + 1

	tx := db
	if sort.SafeExpr() != "" {
		tx = db.Order(sort.SafeExpr())
	}
	if timeRange.StartTime > 0 {
		tx.Where(db.Where("created_at >= ?", timeRange.StartTime))
	}
	if timeRange.EndTime > 0 {
		tx.Where(db.Where("created_at <= ?", timeRange.EndTime))
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
	pageNum, pageSize := param.GetPage(20, 100)
	offset := (pageNum - 1) * pageSize
	limit := pageSize

	// 查询 total
	var total int64
	var model T

	if err := db.Model(&model).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	tx := db
	if sort.SafeExpr() != "" {
		tx = db.Order(sort.SafeExpr())
	}
	if err := tx.
		Offset(offset).
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, nil, err
	}

	// 分页逻辑处理
	return results, &total, nil
}
