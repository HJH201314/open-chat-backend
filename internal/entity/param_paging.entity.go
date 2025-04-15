package entity

type PagingParam struct {
	// 分页参数
	PageNum  int `form:"page_num" json:"page_num" binding:"-"`
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

func (q *PagingParam) GetPageNum() int {
	if q.PageNum == 0 {
		q.PageNum = 1
	}
	return q.PageNum
}

func (q *PagingParam) GetPage(defaultSize int, maxSize int) (int, int) {
	if q.PageSize == 0 {
		q.PageSize = defaultSize
	}
	if q.PageSize > maxSize {
		q.PageSize = maxSize
	}
	return q.GetPageNum(), q.PageSize
}

// PaginatedContinuationResponse 用于基于游标的分页（支持连续获取下一页）
type PaginatedContinuationResponse[T any] struct {
	List     []T    `json:"list"`
	NextPage *int64 `json:"next_page"`
}

func NewPaginatedContinuationResponse[T any](list []T, nextPage *int64) *PaginatedContinuationResponse[T] {
	return &PaginatedContinuationResponse[T]{
		List:     list,
		NextPage: nextPage,
	}
}

// PaginatedSyncListResponse 用于基于游标的同步查询分页（支持连续获取下一页）
type PaginatedSyncListResponse[T any] struct {
	Updated  []T    `json:"updated"`
	Deleted  []T    `json:"deleted"`
	NextPage *int64 `json:"next_page"`
}

// PaginatedTotalResponse 用于传统页码分页（支持跳页和总页数感知）
type PaginatedTotalResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}
