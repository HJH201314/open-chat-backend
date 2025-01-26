package util

import (
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) uint64 {
	claims, exists := c.Get("claims")
	if !exists {
		return 0
	}
	return claims.(*entity.UserClaims).ID
}
