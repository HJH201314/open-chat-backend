package ctx_utils

import (
	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) uint64 {
	claims, exists := c.Get("claims")
	if !exists {
		return 0
	}
	return claims.(*entity.UserClaims).ID
}

func UserIsSuperAdmin(c *gin.Context) bool {
	isSuperAdmin, exists := c.Get(constants.PermissionSuperAdminKey)
	if !exists {
		return false
	}
	return isSuperAdmin.(bool)
}
