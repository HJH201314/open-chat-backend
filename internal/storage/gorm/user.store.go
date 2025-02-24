package gorm

import (
	"github.com/fcraft/open-chat/internal/models"
)

// CreateUser 创建用户
func (s *GormStore) CreateUser(user *models.User) error {
	return s.Db.Create(user).Error
}

// GetUser 获取用户
func (s *GormStore) GetUser(userId uint64) (*models.User, error) {
	var user models.User
	return &user, s.Db.Where("id = ?", userId).First(&user).Error
}

// BindRolesToUser 绑定角色到用户
func (s *GormStore) BindRolesToUser(userId uint64, roleIds []uint64) error {
	var userRoles []models.UserRole
	for _, roleId := range roleIds {
		userRoles = append(
			userRoles, models.UserRole{
				UserID: userId,
				RoleID: roleId,
			},
		)
	}
	return s.Db.Create(&userRoles).Error
}

// UnbindRolesFromUser 解绑角色从用户
func (s *GormStore) UnbindRolesFromUser(userId uint64, roleIds []uint64) error {
	return s.Db.Where("user_id = ? AND role_id in (?)", userId, roleIds).Delete(&models.UserRole{}).Error
}

// UpdateUserRoles 更新用户角色
func (s *GormStore) UpdateUserRoles(userId uint64, roleIds []uint64) error {
	// 先删除用户所有角色
	if err := s.Db.Where("user_id = ?", userId).Delete(&models.UserRole{}).Error; err != nil {
		return err
	}
	// 再绑定新的角色
	return s.BindRolesToUser(userId, roleIds)
}

// AddRole 添加角色
func (s *GormStore) AddRole(role *models.Role) error {
	return s.Db.Create(role).Error
}

// DelRole 删除角色
func (s *GormStore) DelRole(roleId uint64) error {
	return s.Db.Where("id = ?", roleId).Delete(&models.Role{}).Error
}

// AddPermission 添加权限
func (s *GormStore) AddPermission(permission *models.Permission) error {
	return s.Db.Create(permission).Error
}

// DelPermission 删除权限
func (s *GormStore) DelPermission(permissionId uint64) error {
	return s.Db.Where("id = ?", permissionId).Delete(&models.Permission{}).Error
}

// BindPermissionsToRole 绑定权限到角色
func (s *GormStore) BindPermissionsToRole(roleId uint64, permissionIds []uint64) error {
	var rolePermissions []models.RolePermission
	for _, permissionId := range permissionIds {
		rolePermissions = append(
			rolePermissions, models.RolePermission{
				RoleID:       roleId,
				PermissionID: permissionId,
			},
		)
	}
	return s.Db.Create(&rolePermissions).Error
}

// UnbindPermissionsFromRole 解绑权限从角色
func (s *GormStore) UnbindPermissionsFromRole(roleId uint64, permissionIds []uint64) error {
	return s.Db.Where(
		"role_id = ? AND permission_id in (?)",
		roleId,
		permissionIds,
	).Delete(&models.RolePermission{}).Error
}

// UpdateRolePermissions 更新角色权限
func (s *GormStore) UpdateRolePermissions(roleId uint64, permissionIds []uint64) error {
	// 先删除角色所有权限
	if err := s.Db.Where("role_id = ?", roleId).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	// 再绑定新的权限
	return s.BindPermissionsToRole(roleId, permissionIds)
}
