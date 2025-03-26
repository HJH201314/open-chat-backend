package helper

import (
	"github.com/fcraft/open-chat/internal/schema"
)

// CreatePreset 创建预设并缓存
func (s *HandlerHelper) CreatePreset(role *schema.Preset) error {
	// 创建角色
	if err := s.GormStore.CreatePreset(role); err != nil {
		return err
	}

	// 缓存角色
	if err := s.RedisStore.CachePreset(role); err != nil {
		// 缓存失败不影响主流程
	}

	return nil
}

// GetPreset 获取预设，优先从缓存获取
func (s *HandlerHelper) GetPreset(id uint64) (*schema.Preset, error) {
	// 先尝试从缓存获取
	if role, err := s.RedisStore.GetCachedPreset(id); err == nil {
		return role, nil
	}

	// 从数据库获取
	role, err := s.GormStore.GetPreset(id)
	if err != nil {
		return nil, err
	}

	// 缓存角色
	if err := s.RedisStore.CachePreset(role); err != nil {
		// 缓存失败不影响主流程
	}

	return role, nil
}

// ListPresets 获取预设列表，优先从缓存获取
func (s *HandlerHelper) ListPresets() ([]schema.Preset, error) {
	// 先尝试从缓存获取
	if roles, err := s.RedisStore.GetCachedPresets(); err == nil {
		return roles, nil
	}

	// 从数据库获取
	roles, err := s.GormStore.ListPresets()
	if err != nil {
		return nil, err
	}

	// 缓存角色列表
	if err := s.RedisStore.CachePresets(roles); err != nil {
		// 缓存失败不影响主流程
	}

	return roles, nil
}

// UpdatePreset 更新预设并更新缓存
func (s *HandlerHelper) UpdatePreset(role *schema.Preset) error {
	// 更新角色
	if err := s.GormStore.UpdatePreset(role); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeletePresetCache(role.ID); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}

// DeletePreset 删除预设并删除缓存
func (s *HandlerHelper) DeletePreset(id uint64) error {
	// 删除角色
	if err := s.GormStore.DeletePreset(id); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeletePresetCache(id); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}
