package gorm

import (
	"github.com/fcraft/open-chat/internal/schema"
	"gorm.io/gorm"
)

// CreateBotRole 创建机器人角色
func (s *GormStore) CreateBotRole(role *schema.BotRole) error {
	return s.Db.Create(role).Error
}

// GetBotRole 获取机器人角色
func (s *GormStore) GetBotRole(id uint64) (*schema.BotRole, error) {
	var role schema.BotRole
	err := s.Db.Preload("PromptSession").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListBotRoles 获取机器人角色列表
func (s *GormStore) ListBotRoles() ([]schema.BotRole, error) {
	var roles []schema.BotRole
	err := s.Db.Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateBotRole 更新机器人角色
func (s *GormStore) UpdateBotRole(role *schema.BotRole) error {
	return s.Db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(role).Error
}

// DeleteBotRole 删除机器人角色
func (s *GormStore) DeleteBotRole(id uint64) error {
	return s.Db.Delete(&schema.BotRole{}, id).Error
}

// GetBotRoleByPromptSessionId 根据 prompt session id 获取机器人角色
func (s *GormStore) GetBotRoleByPromptSessionId(sessionId string) (*schema.BotRole, error) {
	var role schema.BotRole
	err := s.Db.Preload("PromptSession").Preload("PromptSession.Messages").Where(
		"prompt_session_id = ?",
		sessionId,
	).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
