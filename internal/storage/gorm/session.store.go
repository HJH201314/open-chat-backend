package gorm

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/gorm_utils"
	"gorm.io/gorm"
	"time"
)

func ScopeWithUserId(userId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userId)
	}
}

func ScopePreloadSessionWithOneMessage(db *gorm.DB) *gorm.DB {
	return db.Preload("Session").
		Preload(
			"Session.Messages",
			"id IN (SELECT MIN(id) FROM messages GROUP BY session_id)",
		)
}

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

// UpdateSession 更新会话
func (s *GormStore) UpdateSession(session *schema.Session) error {
	return s.Db.Omit("LastActive").Updates(session).Error
}

// UpdateUserSessionShare 更新用户会话的分享信息
func (s *GormStore) UpdateUserSessionShare(userSession *schema.UserSession) error {
	updateData := map[string]interface{}{
		"share_permanent":  userSession.ShareInfo.Permanent,
		"share_title":      userSession.ShareInfo.Title,
		"share_code":       userSession.ShareInfo.Code,
		"share_expired_at": time.UnixMilli(userSession.ShareInfo.ExpiredAt),
	}
	return s.Db.Model(&schema.UserSession{}).Where(
		"session_id = ? AND user_id = ?",
		userSession.SessionID,
		userSession.UserID,
	).Updates(updateData).Error
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

// filterUserSessions 过滤没有关联 session 的数据并取出 session
func filterUserSessions(userSessions []schema.UserSession) []schema.UserSession {
	return slice.Filter(
		userSessions, func(_ int, userSession schema.UserSession) bool {
			return userSession.Session != nil
		},
	)
}

// GetSessionsByPage 分页获取会话
func (s *GormStore) GetSessionsByPage(userId uint64, page entity.PagingParam, sort entity.SortParam) ([]schema.UserSession, *int64, error) {
	userSessions, nextPage, err := gorm_utils.GetByPageContinuous[schema.UserSession](
		s.Db.Scopes(ScopeWithUserId(userId), ScopePreloadSessionWithOneMessage), page, sort,
	)
	if err != nil {
		return nil, nil, err
	}

	// 过滤没有关联 session 的数据并取出 session
	sessions := filterUserSessions(userSessions)

	return sessions, nextPage, nil
}

// GetSessionsForSync 分页获取用于同步数据的会话
func (s *GormStore) GetSessionsForSync(userId uint64, since time.Time, page entity.PagingParam, sort entity.SortParam) ([]schema.UserSession, *int64, error) {
	// 关闭排序
	sort.WithForceOrder("")
	sessionTable := (&schema.Session{}).TableName()
	messageTable := (&schema.Message{}).TableName()
	userSessions, nextPage, err := gorm_utils.GetByPageContinuous[schema.UserSession](
		s.Db.Unscoped().Where("user_id = ?", userId).
			// 使用手动 JOIN 查询到符合条件的 session
			Joins(
				fmt.Sprintf(
					"INNER JOIN %s AS session ON sessions_users.session_id = session.id",
					sessionTable,
				),
			).
			Joins(
				fmt.Sprintf(
					"LEFT JOIN %s AS message ON session.id = message.session_id AND message.id IN (SELECT MIN(m.id) FROM %s m GROUP BY m.session_id)",
					messageTable, messageTable,
				),
			).
			Where(
				"sessions_users.updated_at > ? OR sessions_users.deleted_at > ? OR session.updated_at > ? OR sessions_users.created_at > ?",
				since,
				since,
				since,
				since,
			).
			Order("COALESCE(sessions_users.deleted_at, sessions_users.updated_at, session.updated_at, sessions_users.created_at) DESC").
			// 使用预加载读取会话和消息的数据
			Preload("Session").Preload("Session.Messages"),
		page,
		sort,
	)

	if err != nil {
		return nil, nil, err
	}

	// 过滤没有关联 session 的数据并取出 session
	sessions := filterUserSessions(userSessions)

	return sessions, nextPage, nil
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
func (s *GormStore) GetMessagesByPage(sessionID string, page entity.PagingParam, sort entity.SortParam) ([]schema.Message, *int64, error) {
	return gorm_utils.GetByPageContinuous[schema.Message](
		s.Db.Preload("Model").Where("session_id = ?", sessionID),
		page,
		sort,
	)
}
