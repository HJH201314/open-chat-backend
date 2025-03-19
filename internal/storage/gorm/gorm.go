package gorm

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/storage/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormStore struct {
	Db     *gorm.DB
	Logger *slog.Logger
	Redis  *redis.RedisStore
}

func NewGormStore(redisStore *redis.RedisStore) *GormStore {
	store := &GormStore{
		Logger: slog.Default(),
		Redis:  redisStore,
	}
	// 初始化 Postgres 连接（在 .env 文件中配置）
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s TimeZone=%s sslmode=disable",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DBNAME"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_TIMEZONE"),
	)
	if os.Getenv("PG_DSN") != "" {
		dsn = os.Getenv("PG_DSN")
	}
	db, err := gorm.Open(
		postgres.Open(dsn), &gorm.Config{
			// 取消外键
			DisableForeignKeyConstraintWhenMigrating: true,
			// 日志
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		store.Logger.Error("failed to connect database")
		panic(err)
	}
	db.Set("gorm:table_options", "AUTO_INCREMENT=100000000")
	// 注册自定义序列化器
	InitSerializer()
	// 自动迁移表结构
	if err := db.AutoMigrate(
		&schema.Session{},
		&schema.Message{},
		&schema.User{},
		&schema.Role{},
		&schema.Permission{},
		&schema.RolePermission{},
		&schema.UserRole{},
		&schema.Provider{},
		&schema.APIKey{},
		&schema.Model{},
		&schema.BotRole{},
		&schema.UserSession{},
		&schema.UserUsage{},
		&schema.Problem{}, &schema.Resource{},
		&schema.Exam{}, &schema.ExamProblem{},
		&schema.Course{}, &schema.CourseResource{}, &schema.CourseExam{},
	); err != nil {
		store.Logger.Error("failed to migrate database")
	}
	// 初始化 GORM 存储
	store.Db = db
	store.Logger.Info("connected to postgres")
	return store
}
