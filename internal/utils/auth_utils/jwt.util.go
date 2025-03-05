package auth_utils

import (
	"github.com/fcraft/open-chat/internal/entity"
	"github.com/fcraft/open-chat/internal/utils/ctx_utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

// SignAuthTokenForUser 为用户签发 token
func SignAuthTokenForUser(userId uint64) (string, error) {
	// 1. 创建 Claims
	claims := entity.UserClaims{
		ID: userId,
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
		return "", err
	}
	return tokenString, nil
}

// SignRefreshTokenForUser 为用户签发刷新 token
func SignRefreshTokenForUser(userId uint64) (string, error) {
	// 1. 创建 Claims
	claims := entity.UserClaims{
		ID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                         // 生效时间
			Issuer:    "open-chat",                                            // 签发者
			Subject:   "user-refresh",                                         // 主题
		},
	}

	// 2. 签发 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("AUTH_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateAuthToken(c *gin.Context) *jwt.Token {
	// 1. 从请求头中获取 token
	authTokenString := ctx_utils.GetRawAuthToken(c)
	if authTokenString == "" {
		return nil
	}

	// 2. 解析 auth_token
	token, err := jwt.ParseWithClaims(
		authTokenString, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("AUTH_SECRET")), nil
		},
	)
	if err != nil {
		return nil
	}

	return token
}

// ValidateRefreshToken 联合验证刷新 token 是否合法，合法返回 refresh_token
func ValidateRefreshToken(c *gin.Context, authClaims *entity.UserClaims) *jwt.Token {
	// 1. 从请求头中获取 token
	refreshTokenString := ctx_utils.GetRawRefreshToken(c)
	if refreshTokenString == "" {
		return nil
	}

	// 2. 解析 refresh_token
	token, err := jwt.ParseWithClaims(
		refreshTokenString, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("AUTH_SECRET")), nil
		},
	)
	if err != nil || !token.Valid {
		return nil
	}

	// 3. 验证 ID 是否相同
	refreshClaims, ok := token.Claims.(*entity.UserClaims)
	if !ok || refreshClaims.ID != authClaims.ID {
		return nil
	}

	return token
}
