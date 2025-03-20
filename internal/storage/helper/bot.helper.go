package helper

import (
	"github.com/fcraft/open-chat/internal/schema"
)

// CreateBotRole 创建机器人角色并缓存
func (s *HandlerHelper) CreateBotRole(role *schema.BotRole) error {
	// 创建角色
	if err := s.GormStore.CreateBotRole(role); err != nil {
		return err
	}

	// 缓存角色
	if err := s.RedisStore.CacheBotRole(role); err != nil {
		// 缓存失败不影响主流程
	}

	return nil
}

// GetBotRole 获取机器人角色，优先从缓存获取
func (s *HandlerHelper) GetBotRole(id uint64) (*schema.BotRole, error) {
	// 先尝试从缓存获取
	if role, err := s.RedisStore.GetCachedBotRole(id); err == nil {
		return role, nil
	}

	// 从数据库获取
	role, err := s.GormStore.GetBotRole(id)
	if err != nil {
		return nil, err
	}

	// 缓存角色
	if err := s.RedisStore.CacheBotRole(role); err != nil {
		// 缓存失败不影响主流程
	}

	return role, nil
}

// ListBotRoles 获取机器人角色列表，优先从缓存获取
func (s *HandlerHelper) ListBotRoles() ([]schema.BotRole, error) {
	// 先尝试从缓存获取
	if roles, err := s.RedisStore.GetCachedBotRoles(); err == nil {
		return roles, nil
	}

	// 从数据库获取
	roles, err := s.GormStore.ListBotRoles()
	if err != nil {
		return nil, err
	}

	// 缓存角色列表
	if err := s.RedisStore.CacheBotRoles(roles); err != nil {
		// 缓存失败不影响主流程
	}

	return roles, nil
}

// UpdateBotRole 更新机器人角色并更新缓存
func (s *HandlerHelper) UpdateBotRole(role *schema.BotRole) error {
	// 更新角色
	if err := s.GormStore.UpdateBotRole(role); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeleteBotRoleCache(role.ID); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}

// DeleteBotRole 删除机器人角色并删除缓存
func (s *HandlerHelper) DeleteBotRole(id uint64) error {
	// 删除角色
	if err := s.GormStore.DeleteBotRole(id); err != nil {
		return err
	}

	// 删除缓存
	if err := s.RedisStore.DeleteBotRoleCache(id); err != nil {
		// 缓存删除失败不影响主流程
	}

	return nil
}
