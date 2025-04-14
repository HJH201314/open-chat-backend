package schema

type ScheduleStatus int

const (
	ScheduleStatusStopped ScheduleStatus = iota
	ScheduleStatusRunning
	ScheduleStatusPending
)

type Schedule struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"not null;unique" json:"name"`
	Description string         `json:"description"`
	Duration    int64          `json:"duration"` // 执行间隔，单位：秒
	LastRunTime int64          `json:"last_run_time"`
	Status      ScheduleStatus `gorm:"default:1" json:"status"`
	AutoCreateAt
}
