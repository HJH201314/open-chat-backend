package services

import (
	"context"
	"errors"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
	"time"
)

type ScheduleTask struct {
	Duration time.Duration
	Schedule func() error
	JobID    uuid.UUID
	Job      *gocron.Job
}

type ScheduleService struct {
	BaseService *BaseService
	db          *gorm.DB
	redis       *redis.Client
	scheduler   gocron.Scheduler
	tasks       map[string]*ScheduleTask
}

var (
	scheduleServiceInstance *ScheduleService
	scheduleServiceOnce     sync.Once
)

func InitScheduleService(base *BaseService) *ScheduleService {
	scheduleServiceOnce.Do(
		func() {
			sd, err := gocron.NewScheduler()
			if err != nil {
				panic(err)
			}
			scheduleServiceInstance = &ScheduleService{
				BaseService: base,
				db:          base.Gorm,
				redis:       base.Redis,
				scheduler:   sd,
				tasks:       make(map[string]*ScheduleTask),
			}
		},
	)
	return scheduleServiceInstance
}

func GetScheduleService() *ScheduleService {
	if scheduleServiceInstance == nil {
		panic("schedule service not initialized")
	}
	return scheduleServiceInstance
}

// StartSchedule 启动所有任务
func (s *ScheduleService) StartSchedule() {
	GetScheduleService().scheduler.Start()
}

// StopSchedule 停止所有任务
func (s *ScheduleService) StopSchedule() error {
	err := scheduleServiceInstance.scheduler.StopJobs()
	if err != nil {
		return err
	}
	return nil
}

// RegisterSchedule 注册一个任务
func (s *ScheduleService) RegisterSchedule(name string, desc string, duration time.Duration, schedule func() error) error {
	// 1. 保存信息
	if err := s.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"description"}),
		},
	).Create(
		&schema.Schedule{
			Name:        name,
			Description: desc,
			Duration:    int64(duration.Seconds()),
		},
	).Error; err != nil {
		return err
	}

	// 2. 创建任务
	s.tasks[name] = &ScheduleTask{
		Duration: duration,
		Schedule: schedule,
		JobID:    uuid.Nil,
	}
	err := s.StartJob(name)
	if err != nil {
		return err
	}

	return nil
}

func (s *ScheduleService) StartJob(name string) error {
	task, ok := s.tasks[name]
	if !ok {
		return errors.New("job not found")
	}
	if task.JobID != uuid.Nil {
		return errors.New("job already started")
	}

	job, err := s.scheduler.NewJob(
		gocron.DurationJob(task.Duration),
		gocron.NewTask(
			func(ctx context.Context) {
				start := time.Now()
				err := task.Schedule()
				if err != nil {
					return
				}
				s.redis.Set(
					ctx,
					"schedule_run:"+name,
					map[string]any{
						"time":        time.Now(),
						"used_millis": time.Since(start).Milliseconds(),
					},
					0,
				)
				s.db.WithContext(ctx).Where("name = ?", name).Updates(
					&schema.Schedule{
						Name:        name,
						LastRunTime: time.Now().UnixMilli(),
						Status:      schema.ScheduleStatusRunning,
					},
				)
			},
		),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		return err
	}
	task.JobID = job.ID()
	task.Job = &job
	return nil
}

// StopJob 停止一个任务
func (s *ScheduleService) StopJob(name string) error {
	task, ok := s.tasks[name]
	if !ok {
		return errors.New("job not found")
	}

	err := s.scheduler.RemoveJob(task.JobID)
	if err != nil {
		return err
	}

	task.JobID = uuid.Nil
	task.Job = nil

	return nil
}

// RunJobNow 立即执行一个任务
func (s *ScheduleService) RunJobNow(name string) error {
	task, ok := s.tasks[name]
	if !ok || task.Job == nil {
		return errors.New("job not running")
	}

	return (*task.Job).RunNow()
}

// IsJobRunning 判断一个任务是否正在运行
func (s *ScheduleService) IsJobRunning(name string) bool {
	task, ok := s.tasks[name]
	if !ok {
		return false
	}
	return task.JobID != uuid.Nil
}
