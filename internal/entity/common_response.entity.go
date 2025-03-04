package entity

import (
	"github.com/fcraft/open-chat/internal/constants"
)

type CommonResponse[T any] struct {
	Code int    `json:"code"` // 代码
	Msg  string `json:"msg"`  // 消息
	Data T      `json:"data"` // 数据
}

func (c *CommonResponse[T]) WithCode(code int) *CommonResponse[T] {
	c.Code = code
	return c
}

func (c *CommonResponse[T]) WithMsg(msg string) *CommonResponse[T] {
	c.Msg = msg
	return c
}

func (c *CommonResponse[T]) WithData(data T) *CommonResponse[T] {
	c.Data = data
	return c
}

func (c *CommonResponse[T]) WithError(err error) *CommonResponse[T] {
	if status, ok := constants.ErrStatusMap[err]; ok {
		c.Code = status
	}
	c.Msg = err.Error()
	return c
}

var (
	// OK 正常
	OK = CommonResponse[any]{
		Code: 200,
		Msg:  "ok",
	}
	// ERR 错误
	ERR = CommonResponse[any]{
		Code: 500,
		Msg:  "err",
	}
)
