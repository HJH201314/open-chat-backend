package storage

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
	); err != nil {
		panic("failed to migrate database")
	}
	// 初始化 GORM 存储
	return &GormStore{Db: db}
}

// CreateUser 创建用户
func (s *GormStore) CreateUser(user *models.User) error {
	return s.Db.Create(user).Error
}

// CreateSession 创建会话
func (s *GormStore) CreateSession(session *models.Session) error {
	return s.Db.Create(session).Error
}

// DeleteSession 删除会话
func (s *GormStore) DeleteSession(sessionId string) error {
	return s.Db.Delete(&models.Session{ID: sessionId}).Error
}

// SaveMessage 保存消息
func (s *GormStore) SaveMessage(msg *models.Message) error {
	return s.Db.Create(msg).Error
}

// ToggleContext 更新会话的上下文开关状态
func (s *GormStore) ToggleContext(sessionID string, enable bool) error {
	return s.Db.Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("enable_context", enable).Error
}

// GetLatestMessages 获取会话的最新消息
func (s *GormStore) GetLatestMessages(sessionID string, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := s.Db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
