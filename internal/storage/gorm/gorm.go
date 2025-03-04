package gorm

import (
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type GormStore struct {
	Db     *gorm.DB
	Logger *log.Logger
}

func NewGormStore() *GormStore {
	store := &GormStore{
		Logger: log.New(log.Writer(), "GormStore", log.LstdFlags),
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		store.Logger.Fatal("failed to connect database")
	}
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
	); err != nil {
		store.Logger.Fatal("failed to migrate database")
	}
	// 初始化 GORM 存储
	store.Db = db
	store.Logger.Println("connected to postgres")
	return store
}
