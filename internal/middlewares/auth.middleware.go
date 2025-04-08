package middlewares

import (
	"slices"
	"strings"

	"github.com/fcraft/open-chat/internal/constants"
	"github.com/fcraft/open-chat/internal/entity"
	redisstore "github.com/fcraft/open-chat/internal/storage/redis"
	"github.com/fcraft/open-chat/internal/utils/auth_utils"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
)

var ignorePaths = []string{
	"/swagger", "/base/public-key", "/user/refresh", "/user/login", "/user/backdoor/login", "/user/logout",
	"/user/register",
}

// AuthMiddleware 鉴权中间件
func AuthMiddleware(redisStore *redisstore.RedisStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 部分路径不需要鉴权
		if slices.ContainsFunc(
			ignorePaths, func(path string) bool {
				return strings.HasPrefix(c.FullPath(), path)
			},
		) {
			c.Set(constants.AuthIgnoredKey, true)
			c.Next()
			return
		}

		// 1. 解析 auth_token
		token := auth_utils.ValidateAuthToken(c)
		if token == nil || !token.Valid {
			ctx_utils.HttpError(c, constants.ErrUnauthorized)
			return
		}

		// 2. 通过缓存验证 token 是否被清理
		userId, err := redisStore.FindUserIdByToken(token.Raw)
		if err != nil {
			ctx_utils.HttpError(c, constants.ErrUnauthorized)
			return
		}

		// 3. 转换 claims
		claims, ok := token.Claims.(*entity.UserClaims)
		if !ok {
			ctx_utils.HttpError(c, constants.ErrUnauthorized)
			return
		}

		// 4. 续期 token
		if err := redisStore.RenewUserToken(userId, token.Raw, constants.RefreshTokenExpire); err != nil {
			ctx_utils.HttpError(c, constants.ErrInternal)
			return
		}

		// 将信息写入上下文
		c.Set("claims", claims)
		c.Next()
	}
}
