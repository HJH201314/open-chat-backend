package entities

type PagingParam struct {
	PageNum  int `form:"page_num" json:"page_num" binding:"required"`
	PageSize int `form:"page_size" json:"page_size" binding:"-"`
}

func (q *PagingParam) GetPageSize(defaultSize int, maxSize int) int {
	if q.PageSize == 0 {
		q.PageSize = defaultSize
	}
	if q.PageSize > maxSize {
		q.PageSize = maxSize
	}
	return q.PageSize
}
