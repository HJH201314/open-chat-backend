package entity

type SearchParam[T any] struct {
	SearchData T `json:"search_data"`
}
