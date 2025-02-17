package middlewares

import (
	"github.com/fcraft/open-chat/internel/shared/constant"
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/fcraft/open-chat/internel/shared/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 登录和注册接口不需要鉴权
		if c.FullPath() == "/user/login" || c.FullPath() == "/user/register" {
			c.Next()
			return
		}
		// 1. 从请求头中获取 token
		tokenString := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1)

		if tokenString == "" {
			util.HttpErrorResponse(c, constant.ErrUnauthorized)
			return
		}

		// 2. 解析并验证 token
		token, err := jwt.ParseWithClaims(tokenString, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("AUTH_SECRET")), nil
		})

		if err != nil || !token.Valid {
			util.HttpErrorResponse(c, constant.ErrUnauthorized)
			return
		}

		// 3. 将解析后的 token 存入上下文
		claims, _ := token.Claims.(*entity.UserClaims)
		c.Set("claims", claims)

		c.Next()
	}
}
