package helper

import "github.com/fcraft/open-chat/internal/schema"

// BindRolesToUser 绑定角色到用户并删除缓存
func (s *HandlerHelper) BindRolesToUser(userId uint64, roleIds []uint64) error {
	err := s.GormStore.BindRolesToUser(userId, roleIds)
	if err != nil {
		return err
	}
	// 删除用户角色缓存
	err = s.RedisStore.DeleteUserRolesCache(userId)
	if err != nil {
		return err
	}
	return nil
}

// UnbindRolesFromUser 取消绑定角色到用户并删除缓存
func (s *HandlerHelper) UnbindRolesFromUser(userId uint64, roleIds []uint64) error {
	err := s.GormStore.UnbindRolesFromUser(userId, roleIds)
	if err != nil {
		return err
	}
	// 删除用户角色缓存
	err = s.RedisStore.DeleteUserRolesCache(userId)
	if err != nil {
		return err
	}
	return nil
}

// GetUserRoles 从数据库和缓存中获取用户角色
func (s *HandlerHelper) GetUserRoles(userId uint64) ([]schema.Role, error) {
	// 先尝试从缓存获取
	if roles, err := s.RedisStore.GetCachedUserRoles(userId); err == nil && len(roles) > 0 {
		return roles, nil
	}
	// 从数据库获取用户角色
	roles, err := s.GormStore.GetUserRoles(userId)
	if err != nil {
		return nil, err
	}
	// 写入缓存
	if len(roles) > 0 {
		_ = s.RedisStore.CacheUserRoles(userId, roles)
	}
	return roles, nil
}
