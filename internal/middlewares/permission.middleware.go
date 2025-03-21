package middlewares

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	handlers "github.com/fcraft/open-chat/internal/storage/helper"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(helper *handlers.HandlerHelper) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 忽略权限检查
		isIgnoreAuth, exists := c.Get(constants.AuthIgnoredKey)
		if exists && isIgnoreAuth.(bool) == true {
			c.Next()
			return
		}

		// 获取用户信息
		claims, exists := c.Get("claims")
		if !exists {
			ctx_utils.HttpError(c, constants.ErrUnauthorized)
			return
		}

		userClaims, ok := claims.(*entity.UserClaims)
		if !ok {
			ctx_utils.HttpError(c, constants.ErrUnauthorized)
			return
		}

		// 构建当前请求的权限路径
		currentPath := c.Request.Method + ":" + c.FullPath()

		// 查询用户角色
		userRoles, err := helper.GetUserRoles(userClaims.ID)
		if err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}

		// 检查用户是否有权限访问
		hasPermission := false
		for _, role := range userRoles {
			if role.Name == "SUPER_ADMIN" {
				hasPermission = true
				c.Set(constants.PermissionSuperAdminKey, true)
				break
			}
			for _, permission := range role.Permissions {
				if permission.Path == currentPath {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			ctx_utils.BizError(c, constants.ErrNoPermission)
			return
		}

		c.Next()
	}
}
