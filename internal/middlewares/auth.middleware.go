package middlewares

import (
	"github.com/fcraft/open-chat/internal/shared/constant"
	"github.com/fcraft/open-chat/internal/shared/entity"
	"github.com/fcraft/open-chat/internal/shared/util"
	"github.com/fcraft/open-chat/internal/utils/auth_utils"
	"github.com/gin-gonic/gin"
	"slices"
	"strings"
)

var ignorePaths = []string{"/swagger", "/user/refresh", "/user/login", "/user/register"}

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 部分路径不需要鉴权
		if slices.ContainsFunc(
			ignorePaths, func(path string) bool {
				return strings.HasPrefix(c.FullPath(), path)
			},
		) {
			c.Next()
			return
		}

		// 解析 auth_token
		token := auth_utils.ValidateAuthToken(c)
		if token == nil || !token.Valid {
			util.HttpErrorResponse(c, constant.ErrUnauthorized)
			return
		}
		// 转换 claims
		claims, ok := token.Claims.(*entity.UserClaims)
		if !ok {
			util.HttpErrorResponse(c, constant.ErrUnauthorized)
			return
		}

		// 将信息写入上下文
		c.Set("claims", claims)
		c.Next()
	}
}
