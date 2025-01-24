package storage

import (
	"github.com/fcraft/open-chat/internel/models"
	"gorm.io/gorm"
)

type GormStore struct {
	Db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{Db: db}
}

// CreateSession 创建会话
func (s *GormStore) CreateSession(session *models.Session) error {
	return s.Db.Create(session).Error
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
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
