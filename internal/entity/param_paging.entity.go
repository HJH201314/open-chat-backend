package entity

type PagingParam struct {
	// 分页参数
	PageNum  int `form:"page_num" json:"page_num" binding:"required"`
	PageSize int `form:"page_size" json:"page_size" binding:"-"`

	// 内部参数
	DefaultSize int `json:"-"`
	MaxSize     int `json:"-"`
}

// WithDefaultSize 链式设置默认分页大小
func (q *PagingParam) WithDefaultSize(defaultSize int) *PagingParam {
	q.DefaultSize = defaultSize
	return q
}

// WithMaxSize 链式设置最大分页大小
func (q *PagingParam) WithMaxSize(maxSize int) *PagingParam {
	q.MaxSize = maxSize
	return q
}

func (q *PagingParam) GetPageSize(defaultSize int, maxSize int) (int, int) {
	if q.PageSize == 0 {
		q.PageSize = defaultSize
	}
	if q.PageSize > maxSize {
		q.PageSize = maxSize
	}
	return q.PageNum, q.PageSize
}

// PaginatedContinuationResponse 用于基于游标的分页（支持连续获取下一页）
type PaginatedContinuationResponse[T any] struct {
	List     []T    `json:"list"`
	NextPage *int64 `json:"next_page"`
}

// PaginatedTotalResponse 用于传统页码分页（支持跳页和总页数感知）
type PaginatedTotalResponse[T any] struct {
	List     []T    `json:"list"`
	LastPage *int64 `json:"last_page"`
}
