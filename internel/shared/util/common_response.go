package util

import (
	"github.com/fcraft/open-chat/internel/shared/entity"
	"github.com/gin-gonic/gin"
)

// SuccessResponse 调用此函数，通过中间件统一成功返回
func SuccessResponse(c *gin.Context, data interface{}) {
	c.Set("data", data)
	c.Next()
}

// ErrorResponse 调用此函数，中断并返回格式化错误
func ErrorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, entity.ERR.WithCode(code).WithMsg(msg))
}
