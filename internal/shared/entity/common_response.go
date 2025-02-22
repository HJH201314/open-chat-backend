package entity

import (
	"github.com/fcraft/open-chat/internal/shared/constant"
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
	if status, ok := constant.ErrStatusMap[err]; ok {
		c.Code = status
	}
	c.Msg = err.Error()
	return c
}

// ResponseWithData 创建一个 CommonResponse
func ResponseWithData[T any](code int, msg string, data T) *CommonResponse[T] {
	return &CommonResponse[T]{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Response(code int, msg string) *CommonResponse[any] {
	return ResponseWithData[any](code, msg, nil)
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
