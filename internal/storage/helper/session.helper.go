package helper

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/schema"
	"time"
)

// GetSharedSession 获取分享的会话
func (s *QueryHelper) GetSharedSession(sessionId string, code string) (*schema.UserSession, error) {
	// 验证是否分享及 Code 是否正确
	var userSession schema.UserSession
	if err := s.Gorm.Preload("Session", s.Gorm.Select("id", "name")).Find(
		&userSession,
		"session_id = ? AND type = ?",
		sessionId,
		schema.UserSessionTypeOwner,
	).Error; err != nil {
		return nil, err
	}
	shareInfo := userSession.ShareInfo
	if shareInfo.Permanent || shareInfo.ExpiredAt > time.Now().UnixMilli() {
		// 验证分享码
		if userSession.ShareInfo.Code != "" && userSession.ShareInfo.Code != code {
			return nil, constants.BizErrNoPermission
		}
	} else {
		if shareInfo.ExpiredAt > 0 {
			return nil, constants.BizErrOutdated
		} else {
			return nil, constants.BizErrNoRecord
		}
	}

	return &userSession, nil
}
