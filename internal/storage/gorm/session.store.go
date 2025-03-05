package gorm

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"gorm.io/gorm"
)

// CreateSession 创建会话
func (s *GormStore) CreateSession(userId uint64, session *schema.Session) error {
	return s.Db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(session).Error; err != nil {
				return err
			}
			if userId > 0 {
				userSession := schema.UserSession{
					UserID:    userId,
					SessionID: session.ID,
					Type:      schema.OWNER,
				}
				if err := tx.Create(&userSession).Error; err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// GetSession 获取会话
func (s *GormStore) GetSession(sessionId string) (*schema.Session, error) {
	var session schema.Session
	return &session, s.Db.Where(
		"id = ?",
		sessionId,
	).First(&session).Error
}

// FindSessionWithUser 联合 sessionId 和 userId 获取会话
func (s *GormStore) FindSessionWithUser(sessionId string, userId uint64) (*schema.Session, error) {
	var session schema.Session
	return &session, s.Db.Where(
		"id = ? AND user_id = ?",
		sessionId,
		userId,
	).First(&session).Error
}

// DeleteSession 从数据库中删除会话
func (s *GormStore) DeleteSession(sessionId string) error {
	return s.Db.Transaction(
		func(tx *gorm.DB) error {
			// 删除会话
			if err := tx.Delete(&schema.Session{ID: sessionId}).Error; err != nil {
				return err
			}
			// 删除消息
			if err := tx.Where("session_id = ?", sessionId).Delete(&schema.Message{}).Error; err != nil {
				return err
			}
			// 删除权限
			if err := tx.Where("session_id = ?", sessionId).Delete(&schema.UserSession{}).Error; err != nil {
				return err
			}
			return nil
		},
	)
}

// GetSessionsByPage 分页获取会话
func (s *GormStore) GetSessionsByPage(userId uint64, page int, pageSize int, sort entity.SortParam) ([]schema.Session, *int, error) {
	offset := (page - 1) * pageSize
	// 多查询一条以判断是否存在下一页
	limit := pageSize + 1

	var userSessions []schema.UserSession
	err := s.Db.Where("user_id = ?", userId).
		Order(sort.WithDefault("id DESC", "id").SafeExpr([]string{})).
		Offset(offset).
		Limit(limit).
		Preload("Session").
		Preload(
			"Session.Messages",
			"id IN (SELECT MIN(id) FROM messages GROUP BY session_id)",
		).
		Find(&userSessions).Error
	if err != nil {
		return nil, nil, err
	}

	sessions := slice.Map(
		userSessions, func(_ int, userSession schema.UserSession) schema.Session {
			return *userSession.Session
		},
	)
	// 分页逻辑处理
	hasNext := len(sessions) > pageSize
	if hasNext {
		sessions = sessions[:pageSize]
		nextPage := page + 1
		return sessions, &nextPage, nil
	}

	return sessions, nil, nil
}

// ToggleContext 更新会话的上下文开关状态
func (s *GormStore) ToggleContext(sessionID string, enable bool) error {
	return s.Db.Model(&schema.Session{}).
		Where("id = ?", sessionID).
		Update("enable_context", enable).Error
}

// CreateMessages 批量创建信息
func (s *GormStore) CreateMessages(msg *[]schema.Message) error {
	return s.Db.Create(msg).Error
}

// SaveMessages 批量保存消息
func (s *GormStore) SaveMessages(msg *[]schema.Message) error {
	return s.Db.Save(msg).Error
}

// DeleteMessages 批量删除消息
func (s *GormStore) DeleteMessages(sessionId string, messageIds []uint64) error {
	return s.Db.Where("session_id = ? AND id in (?)", sessionId, messageIds).Delete(&schema.Message{}).Error
}

// GetLatestMessages 获取会话的最新消息
func (s *GormStore) GetLatestMessages(sessionID string, limit int) ([]schema.Message, error) {
	var messages []schema.Message
	err := s.Db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetMessagesByPage 分页获取消息
func (s *GormStore) GetMessagesByPage(sessionID string, page int, pageSize int, sort entity.SortParam) ([]schema.Message, *int, error) {
	var messages []schema.Message
	offset := (page - 1) * pageSize
	// 多查询一条以判断是否存在下一页
	limit := pageSize + 1

	err := s.Db.Where("session_id = ?", sessionID).
		Order(sort.WithDefault("id ASC", "id").SafeExpr([]string{})). // 保持与现有排序一致
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, nil, err
	}

	// 分页逻辑处理
	hasNext := len(messages) > pageSize
	if hasNext {
		messages = messages[:pageSize]
		nextPage := page + 1
		return messages, &nextPage, nil
	}

	return messages, nil, nil
}
