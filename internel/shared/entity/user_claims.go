package entity

import "github.com/golang-jwt/jwt/v5"

// UserType 用户类型枚举
type UserType string

const (
	User  UserType = "user"
	Guest UserType = "guest"
)

// UserClaims 用户凭证
type UserClaims struct {
	ID   int32    `json:"id"`
	Type UserType `json:"type"`
	jwt.RegisteredClaims
}
