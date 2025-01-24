package middlewares

import (
	"github.com/fcraft/open-chat/internel/shared/constant"
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
	"time"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		unauthorizedResponse := entity.ErrorResponse(constant.ErrUnauthorized)

		// 1. 从请求头中获取 token
		tokenString := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1)

		if tokenString == "" {
			SignTempAuthToken(c)
			c.AbortWithStatusJSON(unauthorizedResponse.Code, unauthorizedResponse)
			return
		}

		// 2. 解析并验证 token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("AUTH_SECRET")), nil
		})

		if err != nil || !token.Valid {
			SignTempAuthToken(c)
			c.AbortWithStatusJSON(unauthorizedResponse.Code, unauthorizedResponse)
			return
		}

		c.Next()
	}
}

// SignTempAuthToken 临时签发 token
func SignTempAuthToken(c *gin.Context) {
	// 1. 创建 Claims
	claims := entity.UserClaims{
		ID: -1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                     // 生效时间
			Issuer:    "open-chat",                                        // 签发者
			Subject:   "user-auth",                                        // 主题
		},
	}

	// 2. 签发 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("AUTH_SECRET")))
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	// 3. 将 token 写入 header
	c.Writer.Header().Set("Temp-Auth-Token", tokenString)
}
