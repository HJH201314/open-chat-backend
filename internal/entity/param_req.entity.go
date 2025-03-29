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

// ReqUpdateBody 更新数据的请求体，用于支持更新部分字段
type ReqUpdateBody[T any] struct {
	Data    T        `json:"data" binding:"required"`
	Updates []string `json:"updates" binding:"required" validate:"min=1"`
}
