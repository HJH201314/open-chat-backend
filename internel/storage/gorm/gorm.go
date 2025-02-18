package gorm

import (
	"github.com/fcraft/open-chat/internel/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormStore struct {
	Db *gorm.DB
}

func InitGormStore() *GormStore {
	// 初始化 Postgres 连接
	// TODO: 请将下面的 DSN 替换为你自己的数据库连接
	dsn := "host=localhost user=postgres password=123456 dbname=open_chat port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 自动迁移表结构
	if err := db.AutoMigrate(
		&models.Session{},
		&models.Message{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},
		&models.Provider{},
		&models.APIKey{},
		&models.Model{},
	); err != nil {
		panic("failed to migrate database")
	}
	// 初始化 GORM 存储
	return &GormStore{Db: db}
}
