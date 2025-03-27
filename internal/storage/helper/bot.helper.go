package helper

import (
	"github.com/fcraft/open-chat/internal/schema"
)

// CreatePreset 创建预设并缓存
func (s *QueryHelper) CreatePreset(role *schema.Preset) error {
	// 创建预设
	if err := s.GormStore.CreatePreset(role); err != nil {
		return err
	}

	// 缓存预设
	if err := s.RedisStore.CachePreset(role); err != nil {
		// 缓存失败不影响主流程
	}

	return nil
}

// GetPreset 获取预设，优先从缓存获取
func (s *QueryHelper) GetPreset(id uint64) (*schema.Preset, error) {
	// 先尝试从缓存获取
	if role, err := s.RedisStore.GetCachedPresetByID(id); err == nil {
		return role, nil
	}

	// 从数据库获取
	role, err := s.GormStore.GetPreset(id)
	if err != nil {
		return nil, err
	}

	// 缓存预设
	if err := s.RedisStore.CachePreset(role); err != nil {
		// 缓存失败不影响主流程
	}

	return role, nil
}

// ListPresets 获取预设列表，优先从缓存获取
func (s *QueryHelper) ListPresets() ([]schema.Preset, error) {
	// 先尝试从缓存获取
	if roles, err := s.RedisStore.GetCachedPresets(); err == nil {
		return roles, nil
	}

	// 从数据库获取
	roles, err := s.GormStore.ListPresets()
	if err != nil {
		return nil, err
	}

	// 缓存预设列表
	if err := s.RedisStore.CachePresets(roles); err != nil {
		// 缓存失败不影响主流程
	}

	return roles, nil
}

// UpdatePreset 更新预设并更新缓存
func (s *QueryHelper) UpdatePreset(preset *schema.Preset) error {
	// 更新预设
	if err := s.GormStore.UpdatePreset(preset); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeletePresetCache(preset.ID, preset.Name); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}

// DeletePreset 删除预设并删除缓存
func (s *QueryHelper) DeletePreset(id uint64) error {
	// 查询预设基础信息
	preset, err := s.GormStore.GetPreset(id)
	if err != nil {
		return err
	}

	// 删除预设
	if err := s.GormStore.DeletePreset(id); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeletePresetCache(id, preset.Name); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}
