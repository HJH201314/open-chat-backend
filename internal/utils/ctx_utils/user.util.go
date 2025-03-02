package ctx_utils

import (
	"github.com/fcraft/open-chat/internal/entities"
	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) uint64 {
	claims, exists := c.Get("claims")
	if !exists {
		return 0
	}
	return claims.(*entities.UserClaims).ID
}
