package entity

import "github.com/duke-git/lancet/v2/slice"

// PathParamId 路径参数ID
type PathParamId struct {
	ID uint64 `uri:"id" binding:"required"`
}

// PathParamName 路径参数Name
type PathParamName struct {
	Name string `uri:"name" binding:"required"`
}

type ParamPagingSort struct {
	SortParam
	PagingParam
	TimeRangeParam
}

type ParamSearchPagingSort[T any] struct {
	SearchParam[T]
	ParamPagingSort
}

// ReqUpdateBody 更新数据的请求体，用于支持更新部分字段
type ReqUpdateBody[T any] struct {
	Data    T        `json:"data" binding:"required"`
	Updates []string `json:"updates" binding:"required" validate:"min=1"`
}

// WithWhitelist 根据给定的更新字段列表，过滤掉不在白名单中的字段
func (r *ReqUpdateBody[T]) WithWhitelist(updates ...string) {
	filteredUpdates := slice.Filter(
		updates, func(_ int, update string) bool {
			return slice.Contain(r.Updates, update)
		},
	)
	r.Updates = filteredUpdates
}
