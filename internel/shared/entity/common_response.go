package entity

import (
	"github.com/fcraft/open-chat/internel/shared/constant"
)

type CommonResponse struct {
	Code int         `json:"code"` // 代码
	Msg  string      `json:"msg"`  // 消息
	Data interface{} `json:"data"` // 数据
}

func (c *CommonResponse) WithCode(code int) *CommonResponse {
	c.Code = code
	return c
}

func (c *CommonResponse) WithMsg(msg string) *CommonResponse {
	c.Msg = msg
	return c
}

func (c *CommonResponse) WithData(data interface{}) *CommonResponse {
	c.Data = data
	return c
}

func (c *CommonResponse) WithError(err error) *CommonResponse {
	if status, ok := constant.ErrStatusMap[err]; ok {
		c.Code = status
	}
	c.Msg = err.Error()
	return c
}

// ResponseWithData 创建一个 CommonResponse
func ResponseWithData(code int, msg string, data interface{}) *CommonResponse {
	return &CommonResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Response(code int, msg string) *CommonResponse {
	return ResponseWithData(code, msg, nil)
}

func ErrorResponse(err error) *CommonResponse {
	return Response(constant.ErrStatusMap[err], err.Error())
}

var (
	// OK 正常
	OK = Response(200, "ok")
	// ERR 错误
	ERR = Response(500, "err")
)
