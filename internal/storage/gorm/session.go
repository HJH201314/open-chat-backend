package gorm

import (
	"github.com/fcraft/open-chat/internal/models"
	"gorm.io/gorm"
)

// CreateSession 创建会话
func (s *GormStore) CreateSession(session *models.Session) error {
	return s.Db.Create(session).Error
}

// DeleteSession 删除会话
func (s *GormStore) DeleteSession(sessionId string) error {
	return s.Db.Transaction(func(tx *gorm.DB) error {
		// 删除会话
		if err := tx.Delete(&models.Session{ID: sessionId}).Error; err != nil {
			return err
		}
		// 删除消息
		if err := tx.Where("session_id = ?", sessionId).Delete(&models.Message{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ToggleContext 更新会话的上下文开关状态
func (s *GormStore) ToggleContext(sessionID string, enable bool) error {
	return s.Db.Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("enable_context", enable).Error
}

// CreateMessages 批量创建信息
func (s *GormStore) CreateMessages(msg *[]models.Message) error {
	return s.Db.Create(msg).Error
}

// SaveMessages 批量保存消息
func (s *GormStore) SaveMessages(msg *[]models.Message) error {
	return s.Db.Save(msg).Error
}

// DeleteMessages 批量删除消息
func (s *GormStore) DeleteMessages(sessionId string, messageIds []uint64) error {
	return s.Db.Where("session_id = ? AND id in (?)", sessionId, messageIds).Delete(&models.Message{}).Error
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
