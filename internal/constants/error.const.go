package constants

import (
	"errors"
	"net/http"
)

// 预定义错误类型
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrBadRequest   = errors.New("invalid request")
	ErrInternal     = errors.New("internal server error")
)

// ErrStatusMap 错误类型与 HTTP 状态码的映射
var ErrStatusMap = map[error]int{
	ErrNotFound:     http.StatusNotFound,
	ErrUnauthorized: http.StatusUnauthorized,
	ErrBadRequest:   http.StatusBadRequest,
	ErrInternal:     http.StatusInternalServerError,
}
