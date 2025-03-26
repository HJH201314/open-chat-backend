package gorm

import (
	"github.com/fcraft/open-chat/internal/schema"
	"gorm.io/gorm"
)

// CreatePreset 创建预设
func (s *GormStore) CreatePreset(role *schema.Preset) error {
	return s.Db.Create(role).Error
}

// GetPreset 获取预设
func (s *GormStore) GetPreset(id uint64) (*schema.Preset, error) {
	var role schema.Preset
	err := s.Db.Preload("PromptSession").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListPresets 获取预设列表
func (s *GormStore) ListPresets() ([]schema.Preset, error) {
	var roles []schema.Preset
	err := s.Db.Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdatePreset 更新预设
func (s *GormStore) UpdatePreset(role *schema.Preset) error {
	return s.Db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(role).Error
}

// DeletePreset 删除预设
func (s *GormStore) DeletePreset(id uint64) error {
	return s.Db.Delete(&schema.Preset{}, id).Error
}

// GetPresetByPromptSessionId 根据 prompt session id 获取预设
func (s *GormStore) GetPresetByPromptSessionId(sessionId string) (*schema.Preset, error) {
	var role schema.Preset
	err := s.Db.Preload("PromptSession").Preload("PromptSession.Messages").Where(
		"prompt_session_id = ?",
		sessionId,
	).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
