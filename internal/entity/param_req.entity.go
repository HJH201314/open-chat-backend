package entity

// PathParamId 路径参数ID
type PathParamId struct {
	ID uint64 `uri:"id" binding:"required"`
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
