package entity

type TimeRangeParam struct {
	StartTime int64 `json:"start_time" form:"start_time" binding:"-"`
	EndTime   int64 `json:"end_time" form:"end_time" binding:"-"`
}
