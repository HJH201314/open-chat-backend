package entity

type PagingParam struct {
	PageNum  int `form:"page_num" json:"page_num" binding:"required"`
	PageSize int `form:"page_size" json:"page_size" binding:"-"`
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

type PagingResponse[T any] struct {
	List     []T  `json:"list"`
	NextPage *int `json:"next_page"`
}
